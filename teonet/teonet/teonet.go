package teonet

//// CGO definition (don't delay or edit it):
//#include "packet.h"
import "C"
import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

// MODULE Teonet module name for using in logging
var MODULE = teokeys.Color(teokeys.ANSILightCyan, "(teonet)")

const (
	localhostIP   = "127.0.0.1"
	localhostIPv6 = "::1"
)

// Parameters is Teonet parameters
type Parameters struct {
	Name             string // this host client name
	Port             int    // local port
	RAddr            string // remote host address
	RPort, RChan     int    // remote host port and channel(for TRUdp only)
	Network          string // teonet network name
	LogLevel         string // show log messages level
	LogFilter        string // log messages filter
	L0tcpPort        int    // L0 Server TCP port number (default 9000)
	ForbidHotkeysF   bool   // forbid hotkeys menu
	ShowTrudpStatF   bool   // show trudp statistic
	ShowPeersStatF   bool   // show peers table
	ShowClientsStatF bool   // show clients table
	ShowHelpF        bool   // show usage
	IPv6Allow        bool   // Allow IPv6 support (not supported in Teonet-C)
	L0allow          bool   // Allow l0 server
	DisallowEncrypt  bool   // Disable teonet packets encryption

}

// Teonet teonet connection data structure
type Teonet struct {
	td         *trudp.TRUDP        // TRUdp connection
	param      *Parameters         // Teonet parameters
	cry        *crypt              // Crypt module
	com        *command            // Commands module
	wcom       *waitCommand        // Command wait module
	arp        *arp                // Arp module
	rhost      *rhostData          // R-host module
	split      *splitPacket        // Solitter module
	l0         *l0                 // L0 server module
	menu       *teokeys.HotkeyMenu // Hotkey menu
	ticker     *time.Ticker        // Idle timer ticker (to use in hokeys)
	chanKernel chan func()         // Channel to execute function on kernel level
	ctrlc      bool                // Ctrl+C is on flag (for use in reconnect)
	running    bool                // Teonet running flag
	reconnect  bool                // Teonet reconnect flag
	wg         sync.WaitGroup      // Wait stopped
}

// Connect initialize Teonet
func Connect(param *Parameters) (teo *Teonet) {

	// Create Teonet connection structure and Init logger
	teo = &Teonet{param: param, running: true}
	teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile, param.LogFilter)

	// Timer ticker and kernel channel init
	teo.ticker = time.NewTicker(250 * time.Millisecond)
	teo.chanKernel = make(chan func())

	// Command, Command wait and Crypto modules init
	teo.com = &command{teo}
	teo.wcom = teo.waitFromNew()
	teo.cry = teo.cryptNew(param.Network)

	// Trudp init
	teo.td = trudp.Init(&param.Port)
	teo.td.AllowEvents(1) // \TODO: set events connected by '||'' to allow it
	teo.td.ShowStatistic(param.ShowTrudpStatF)

	// Arp module init
	teo.arp = &arp{teo: teo, m: make(map[string]*arpData)}
	teo.arp.peerAdd(param.Name, teo.Version())

	// R-host module init and Connect to remote host (r-host)
	teo.rhost = &rhostData{teo: teo}

	// Splitter modules
	teo.split = teo.splitNew()

	// L0 server module init
	teo.l0 = teo.l0New()

	// Hotkeys CreateMenu
	teo.createMenu()

	return
}

// Reconnect reconnects Teonet
func (teo *Teonet) Reconnect() {
	teo.reconnect = true
	teo.Close()
}

// Run start Teonet event loop
func (teo *Teonet) Run() {
	for teo.running {

		// Reader
		go func() {
			defer teo.td.ChanEventClosed()
			teo.wg.Add(1)
			for teo.running {
				rd, err := teo.read()
				if err != nil || rd == nil {
					teolog.Error(MODULE, rd, err)
					continue
				}
				teolog.DebugVf(MODULE, "got packet: cmd %d from %s, data len: %d, data: %v\n",
					rd.Cmd(), rd.From(), len(rd.Data()), rd.Data())
			}
			teo.wg.Done()
		}()

		// Start running
		teo.rhost.run()
		teo.td.Run()
		teo.running = false
		teo.wg.Wait()
		teolog.Connect(MODULE, "stopped")

		// Reconnect
		if teo.reconnect {
			appType := teo.GetType()
			ctrlc := teo.ctrlc
			//teolog.Connect(MODULE, "reconnect...")
			fmt.Println("reconnect...")
			param := teo.param
			teo = nil
			time.Sleep(1 * time.Second)
			teo = Connect(param)
			teo.SetType(appType)
			if ctrlc {
				teo.CtrlC()
			}
			teo.reconnect = false
			teo.running = true
		}
	}
}

// Close stops Teonet running
func (teo *Teonet) Close() {
	fmt.Println("teo.running:", teo.running)
	if !teo.running {
		return
	}
	teo.running = false

	teo.menu.Quit()
	teo.l0.destroy()
	teo.arp.deleteAll()
	teo.rhost.destroy()
	teo.td.Close()

	close(teo.chanKernel)
	teo.ticker.Stop()

	teo.cry.destroy()
}

// kernel run function in trudp kernel (main process)
func (teo *Teonet) kernel(f func()) {
	teo.chanKernel <- f
}

// read reads and parse network packet
func (teo *Teonet) read() (rd *C.ksnCorePacketData, err error) {
FOR:
	for teo.running {
		select {
		// Trudp event
		case ev, ok := <-teo.td.ChanEvent():
			if !ok {
				break FOR
			}
			packet := ev.Data

			// Process trudp events
			switch ev.Event {

			case trudp.CONNECTED:
				teolog.Connect(MODULE, "got CONNECTED event, channel key: "+string(packet))

			case trudp.DISCONNECTED:
				teolog.Connect(MODULE, "got DISCONNECTED event, channel key: "+string(packet))
				// Reconnect to r-host
				teo.rhost.reconnect(ev.Tcd)
				// Delete peer from arp table
				teo.arp.deleteKey(string(packet))
				// Close l0 client
				if client, ok := teo.l0.findAddr(string(packet)); ok {
					teo.l0.close(client)
				}

			case trudp.RESET_LOCAL:
				err = errors.New("got RESET_LOCAL event, channel key: " + ev.Tcd.GetKey())
				teolog.Connect(MODULE, err.Error())
				//ev.Tcd.CloseChannel()
				//break FOR

			case trudp.GOT_DATA, trudp.GOT_DATA_NOTRUDP:
				teolog.DebugVvf(MODULE, "got %d bytes packet, channel key: %s\n",
					len(packet), ev.Tcd.GetKey())
				packet, err = teo.cry.decrypt(packet, ev.Tcd.GetKey())
				if err != nil && teo.l0.allow {
					// if packet does not decrypted than it may be l0 client trudp packet.
					// Check l0 packet and process it if this packet is valid teocli(l0)
					// packet
					if _, status := teo.l0.check(ev.Tcd, packet); status != 1 {
						continue FOR
					}
				}
				pac := &Packet{packet: packet} // Create Packet and parse it
				if rd, err = pac.Parse(); err == nil {
					//teolog.DebugVvf(MODULE, "got valid packet cmd: %d, name: %s, data_len: %d\n", pac.Cmd(), pac.From(), pac.DataLen())
					// \TODO don't return error on Parse err != nil, because error is interpreted as disconnect
					if !teo.com.process(&receiveData{rd, ev.Tcd}) {
						break FOR
					}
				} else {
					teolog.DebugVvf(MODULE, teokeys.Color(teokeys.ANSIRed,
						"got invalid (not teonet) packet")+", channel key: %s\n",
						ev.Tcd.GetKey())
					rd = nil
				}

			case trudp.GOT_ACK_PING:
				// triptime, _ := ev.Tcd.GetTriptime()
				// teolog.DebugVv(MODULE, "got GOT_ACK_PING, key:", ev.Tcd.GetKey(), "triptime:", triptime, "ms")
				teo.arp.print()

			default:
				var key string
				if ev.Tcd != nil {
					key = ev.Tcd.GetKey()
				}
				teolog.Logf(teolog.DEBUGvv, MODULE,
					"got unknown event: %d, channel key: %s\n", ev.Event, key)
			}

		// Execute function on Teonet kernel level
		case f, ok := <-teo.chanKernel:
			if !ok {
				return
			}
			f()

		// Timer iddle event
		case <-teo.ticker.C:
			//teolog.Debug(MODULE, "got ticker event")
			if teo.menu != nil && !teo.param.ForbidHotkeysF {
				teo.menu.Check()
			}
		}
	}
	return
}

// sendToHimself send command to this host
func (teo *Teonet) sendToHimself(to string, cmd byte, data []byte) (err error) {

	teolog.Debugf(MODULE, "send command to this host: %s, cmd: %d, data_len: %d\n",
		to, cmd, len(data))

	rd, err := teo.packetCreateNew(teo.param.Name, byte(cmd), data).Parse()
	if err != nil {
		err = errors.New("can't parse packet to himself")
		return
	}

	teo.com.process(&receiveData{rd, nil})
	return
}

// SendTo send command to Teonet peer
func (teo *Teonet) SendTo(to string, cmd byte, data []byte) (err error) {
	arp, ok := teo.arp.m[to]
	if !ok {
		err = errors.New("peer " + to + " not connected to this host")
		return
	}
	if arp.tcd == nil {
		//err = errors.New("send himself not implemented yet")
		return teo.sendToHimself(to, cmd, data)
	}
	return teo.sendToTcd(arp.tcd, cmd, data)
}

// SenToL0 send command to Teonet L0 client
func (teo *Teonet) SendToL0(l0Peer string, client string, cmd byte, data []byte) (err error) {
	teo.l0.sendToL0(l0Peer, client, cmd, data)
	return nil
}

// SendAnswer send command to Teonet peer by receiveData
func (teo *Teonet) SendAnswer(rec *receiveData, cmd byte, data []byte) (err error) {
	if !rec.rd.IsL0() {
		return teo.sendToTcd(rec.tcd, cmd, data)
	} else {
		addr, port := C.GoString(rec.rd.addr), int(rec.rd.port)
		//fmt.Printf("SEND TO CLIENT %s from l0 '%s:%d'\n", rec.rd.From(), addr, port)
		if addr == "" && port == teo.param.Port && teo.l0.allow {
			teo.l0.sendTo(teo.param.Name, rec.rd.From(), cmd, data)
			return
		}
		arp, ok := teo.arp.find(addr, port, 0)
		if ok {
			peer := arp.peer
			//fmt.Printf("SEND TO CLIENT %s from peer: %s\n", rec.rd.From(), peer)
			return teo.SendToL0(peer, rec.rd.From(), cmd, data)
		}
		return nil
	}
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcd(tcd *trudp.ChannelData, cmd byte, data []byte) (err error) {

	// makePac creates new teonet packet and show 'send' log message
	makePac := func(tcd *trudp.ChannelData, cmd byte, data []byte) []byte {
		pac := teo.packetCreateNew(teo.param.Name, cmd, data)
		to, _ := teo.arp.peer(tcd)
		teolog.DebugVf(MODULE, "send cmd: %d, to: %s, data_len: %d\n", cmd, to, len(data))
		return teo.cry.encrypt(pac.packet)
	}

	// send splitted packet or send whole packet
	if tcd != nil {
		_, err = teo.split.split(cmd, data, func(cmd byte, data []byte) {
			_, err = tcd.Write(makePac(tcd, cmd, data))
		})
	} else {
		teo.sendToHimself(teo.param.Name, cmd, data)
	}
	return
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcdUnsafe(tcd *trudp.ChannelData, cmd byte, data []byte) (int, error) {
	pac := teo.packetCreateNew(teo.param.Name, cmd, data)
	to, _ := teo.arp.peer(tcd)
	teolog.DebugVf(MODULE, "send cmd: %d, to: %s, data_len: %d (send direct udp)\n",
		cmd, to, len(data))
	// \TODO: split data!! We can't split Unsafe packet bekause we can't delivery
	// much unfsafe packets and than combine it. So sugest return err "too large
	// data packet" if the length more than 1024 - 1280 (more than mtu, more than
	// udp packet)... or lets him try send any size packets :-)
	return tcd.WriteToUnsafe(teo.cry.encrypt(pac.packet))
}

// GetType return this teonet application type (array of types)
func (teo *Teonet) GetType() []string {
	// Select this host in arp table
	peerArp, ok := teo.arp.m[teo.param.Name]
	if !ok {
		//err = errors.New("host " + teo.param.Name + " does not exist in arp table")
		return nil
	}
	return peerArp.appType
}

// Host return host name byte array with leading zerro
func (teo *Teonet) Host() []byte {
	return append([]byte(teo.param.Name), 0)
}

// SetType set this teonet application type (array of types)
func (teo *Teonet) SetType(appType []string) (err error) {
	// Select this host in arp table
	peerArp, ok := teo.arp.m[teo.param.Name]
	if !ok {
		err = errors.New("host " + teo.param.Name + " does not exist in arp table")
		return
	}
	// Set application type
	peerArp.appType = appType
	return
}

// SetVersion set this teonet application version
func (teo *Teonet) SetVersion(appVersion string) (err error) {
	// Select this host in arp table
	peerArp, ok := teo.arp.m[teo.param.Name]
	if !ok {
		err = errors.New("host " + teo.param.Name + " does not exist in arp table")
		return
	}
	// Set application version
	peerArp.appVersion = appVersion
	return
}

// Version return teonet version
func (teo *Teonet) Version() string {
	return Version
}

// CtrlC process Ctrl+C to close Teonet
func (teo *Teonet) CtrlC() {
	teo.ctrlc = true
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGINT, syscall.SIGKILL:
				teo.Close()
				close(c)
				return
			case syscall.SIGCLD:
				fallthrough
			default:
				fmt.Printf("sig: %x\n", sig)
			}
		}
	}()
}

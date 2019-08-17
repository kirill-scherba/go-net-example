package teonet

//// CGO definition (don't delay or edit it):
//#cgo LDFLAGS: -lcrypto
//#include "net_core.h"
//#include "crypt.h"
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
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

const (
	localhostIP   = "127.0.0.1"
	localhostIPv6 = "::1"
)

// MODULE Teonet module name for using in logging
var MODULE = teokeys.Color(teokeys.ANSILightCyan, "(teonet)")

// Parameters is Teonet parameters
type Parameters struct {
	Name           string // this host client name
	Port           int    // local port
	RAddr          string // remote host address
	RPort, RChan   int    // remote host port and channel(for TRUdp only)
	Network        string // teonet network name
	LogLevel       string // show log messages level
	LogFilter      string // log messages filter
	ForbidHotkeysF bool   // forbid hotkeys menu
	ShowTrudpStatF bool   // show trudp statistic
	ShowPeersStatF bool   // show peers table
	ShowHelpF      bool   // show usage
}

// Teonet teonet connection data structure
type Teonet struct {
	td        *trudp.TRUDP        // TRUdp connection
	param     *Parameters         // Teonet parameters
	kcr       *C.ksnCryptClass    // C crypt module
	com       *command            // Commands module
	arp       *arp                // Arp module
	rhost     *rhostData          // R-host module
	menu      *teokeys.HotkeyMenu // Hotkey menu
	ticker    *time.Ticker        // Idle timer ticker (to use in hokeys)
	ctrlc     bool                // Ctrl+C is on flag (for use in reconnect)
	running   bool                // Teonet running flag
	reconnect bool                // Teonet reconnect flag
	wg        sync.WaitGroup      // Wait stopped
}

// Connect initialize Teonet
func Connect(param *Parameters) (teo *Teonet) {

	// Create Teonet connection structure and Init logger
	teo = &Teonet{param: param, running: true}
	teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile, param.LogFilter)

	// Command and Crypto modules init
	teo.com = &command{teo}
	cnetwork := append([]byte(param.Network), 0)
	teo.kcr = C.ksnCryptInit((*C.char)(unsafe.Pointer(&cnetwork[0])))

	// Trudp init
	teo.td = trudp.Init(param.Port)
	teo.td.AllowEvents(1) // \TODO: set events connected by '||'' to allow it
	teo.td.ShowStatistic(param.ShowTrudpStatF)

	// Arp module init
	teo.arp = &arp{teo: teo, m: make(map[string]*arpData)}
	teo.arp.peerAdd(param.Name, teo.version())

	// R-host module init and Connect to remote host (r-host)
	teo.rhost = &rhostData{teo: teo}
	teo.rhost.run()

	// Timer ticker
	teo.ticker = time.NewTicker(250 * time.Millisecond)

	// Hotkeys CreateMenu
	teo.createMenu()

	return
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
					teolog.Error(MODULE, err)
					continue
				}
				teolog.DebugVf(MODULE, "got packet: cmd %d from %s, data len: %d, data: %v\n",
					rd.Cmd(), rd.From(), len(rd.Data()), rd.Data())
			}
			teo.wg.Done()
		}()

		// Start running
		teo.td.Run()
		teo.running = false
		teo.wg.Wait()
		teolog.Connect(MODULE, "stopped")

		// Reconnect
		if teo.reconnect {
			appType := teo.GetType()
			ctrlc := teo.ctrlc
			teolog.Connect(MODULE, "reconnect...")
			time.Sleep(1 * time.Second)
			teo = Connect(teo.param)
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
	teo.running = false
	teo.arp.deleteAll()
	teo.td.Close()
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
				teo.rhost.reconnect(ev.Tcd)
				teo.arp.deleteKey(string(packet))

			case trudp.RESET_LOCAL:
				err = errors.New("got RESET_LOCAL event, channel key: " + ev.Tcd.GetKey())
				teolog.Connect(MODULE, err.Error())
				//ev.Tcd.CloseChannel()
				//break FOR

			case trudp.GOT_DATA, trudp.GOT_DATA_NOTRUDP:
				teolog.DebugVvf(MODULE, "got %d bytes packet, channel key: %s\n", len(packet), ev.Tcd.GetKey())
				// Decrypt
				var decryptLen C.size_t
				packetPtr := unsafe.Pointer(&packet[0])
				C.ksnDecryptPackage(teo.kcr, packetPtr, C.size_t(len(packet)), &decryptLen)
				if decryptLen > 0 {
					packet = packet[2 : decryptLen+2]
					teolog.DebugVvf(MODULE, "decripted to %d bytes packet, channel key: %s\n", decryptLen, ev.Tcd.GetKey())
				} else {
					teolog.DebugVvf(MODULE, "can't decript %d bytes packet (try to use without decrypt), channel key: %s\n", len(packet), ev.Tcd.GetKey())
				}
				// Create Packet and parse it
				pac := &Packet{packet: packet}
				if rd, err = pac.Parse(); err == nil {
					//teolog.DebugVvf(MODULE, "got valid packet cmd: %d, name: %s, data_len: %d\n", pac.Cmd(), pac.From(), pac.DataLen())
					// \TODO don't return error on Parse err != nil, because error is interpreted as disconnect
					if !teo.com.process(&receiveData{rd, ev.Tcd}) {
						break FOR
					}
				} else {
					teolog.DebugVvf(MODULE, teokeys.Color(teokeys.ANSIRed, "got invalid (not teonet) packet")+", channel key: %s\n", ev.Tcd.GetKey())
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
				teolog.Logf(teolog.DEBUGvv, MODULE, "got unknown event: %d, channel key: %s\n", ev.Event, key)
			}

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

// SendTo send command to Teonet peer
func (teo *Teonet) SendTo(to string, cmd int, data []byte) (err error) {
	arp, ok := teo.arp.m[to]
	if !ok {
		err = errors.New("peer " + to + " not connected to this host")
		return
	}
	return teo.sendToTcd(arp.tcd, cmd, data)
}

// SendAnswer send command to Teonet peer by receiveData
func (teo *Teonet) SendAnswer(rec *receiveData, cmd int, data []byte) (err error) {
	return teo.sendToTcd(rec.tcd, cmd, data)
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcd(tcd *trudp.ChannelData, cmd int, data []byte) (err error) {
	pac := teo.packetCreateNew(cmd, teo.param.Name, data)
	// \TODO: encrypt data
	return tcd.WriteTo(pac.packet)
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcdUnsafe(tcd *trudp.ChannelData, cmd int, data []byte) (int, error) {
	pac := teo.packetCreateNew(cmd, teo.param.Name, data)
	return tcd.WriteToUnsafe(pac.packet)
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

// version return teonet version
func (teo *Teonet) version() string {
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

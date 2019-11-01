// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teonet contain Teonet server functions and data structures.
package teonet

// #include "packet.h"
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

	"github.com/kirill-scherba/teonet-go/services/teoapi"

	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

// MODULE Teonet module name for using in logging
var MODULE = teokeys.Color(teokeys.ANSILightCyan, "(teonet)")

const (
	localhostIP   = "127.0.0.1"
	localhostIPv6 = "::1"
)

// Teonet teonet connection data structure
type Teonet struct {
	td         *trudp.TRUDP        // TRUdp connection
	param      *Parameters         // Teonet parameters
	cry        *crypt              // Crypt module
	ev         *event              // Event module
	com        *command            // Commands module
	wcom       *waitCommand        // Command wait module
	arp        *arp                // Arp module
	rhost      *rhostData          // R-host module
	split      *splitPacket        // Solitter module
	l0         *l0Conn             // L0 server module
	api        *teoapi.Teoapi      // Teonet registry api module
	menu       *teokeys.HotkeyMenu // Hotkey menu
	ticker     *time.Ticker        // Idle timer ticker (to use in hokeys)
	chanKernel chan func()         // Channel to execute function on kernel level
	running    bool                // Teonet running flag
	reconnect  bool                // Teonet reconnect flag
	wg         sync.WaitGroup      // Wait stopped
}

// Logo print teonet logo
func Logo(title, ver string) {
	fmt.Println("" +
		" _____                     _   \n" +
		"|_   _|__  ___  _ __   ___| |_ \n" +
		"  | |/ _ \\/ _ \\| '_ \\ / _ \\ __|\n" +
		"  | |  __/ (_) | | | |  __/ |_ \n" +
		"  |_|\\___|\\___/|_| |_|\\___|\\__|\n" +
		"\n" +
		title + " ver " + ver + ", based on teonet ver " + Version +
		"\n",
	)
}

// Connect initialize Teonet
// Note: The forth parameter may be added to this function. There is value of
// '*teoapi.Teoapi'. If this api parammeter is set than api menu item adding
// to the hotkey menu. The api definition shoud be done before the
// 'teonet.Params(api)' calls.
func Connect(param *Parameters, appType []string, appVersion string,
	ii ...interface{}) (teo *Teonet) {

	// Create Teonet connection structure and Init logger
	teo = &Teonet{param: param, running: true}
	teolog.Init(param.Loglevel, log.Lmicroseconds|log.Lshortfile,
		param.LogFilter, param.LogToSyslogF, param.Name)

	// Timer ticker and kernel channel init
	teo.ticker = time.NewTicker(250 * time.Millisecond)
	teo.chanKernel = make(chan func())

	// Command, Command wait, Crypto and Event modules init
	teo.com = &command{teo}
	teo.wcom = teo.waitFromNew()
	teo.cry = teo.cryptNew(param.Network)
	teo.ev = teo.eventNew()

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

	// Process Ctrl+C to close Teonet
	teo.ctrlC()

	// Set app type
	teo.setType(appType)

	// Set app version
	teo.setAppVersion(appVersion)

	// Teonet api registry
	if len(ii) > 0 {
		if v, ok := ii[0].(*teoapi.Teoapi); ok {
			teo.api = v
			teo.Menu().Add('a', "show teonet application api", func() {
				fmt.Printf("\b%s\n", v.Sprint())
			})
		}
	}

	return
}

// Menu is Hotkey menu getter
func (teo *Teonet) Menu() *teokeys.HotkeyMenu {
	return teo.menu
}

// Reconnect reconnects Teonet
func (teo *Teonet) Reconnect() {
	teo.reconnect = true
	teo.Close()
}

// Run start Teonet event loop
func (teo *Teonet) Run(proccess func(*Teonet)) {
	fmt.Print("\0337" + "\033[r" + "\0338") // reset terminal scrolling
	for teo.running {
		// Reader
		go func() {
			defer teo.td.ChanEventClosed()
			teo.wg.Add(1)
			teo.ev.send(EventStarted, nil)
			for teo.running {
				rd, err := teo.read()
				if err != nil || rd == nil {
					//teolog.Error(MODULE, rd, err)
					continue
				}
				// teolog.DebugVf(MODULE,
				// 	"got packet: cmd %d from %s, data len: %d\n",
				// 	rd.Cmd(), rd.From(), len(rd.Data()),
				// )
				// teo.ev.send(EventReceived, rd.Packet())
			}
			teo.ev.send(EventStoppedBefore, nil)
			teo.wg.Done()
			teo.ev.send(EventStopped, nil)
			teo.ev.close()
		}()

		// Users level event loop process (or empty loop if proccess parameter skipped)
		if proccess == nil {
			proccess = func(teo *Teonet) {
				for range teo.ev.ch {
				}
			}
		}
		go func() { teo.wg.Add(1); proccess(teo); teo.wg.Done() }()

		// Start running
		teo.rhost.run()
		teo.td.Run()
		teo.running = false
		teo.wg.Wait()
		teolog.Connect(MODULE, "stopped")

		// Reconnect
		if teo.reconnect {
			param := teo.param
			appType := teo.Type()
			appVersion := teo.AppVersion()
			fmt.Println("reconnect...")
			teo = nil
			time.Sleep(1 * time.Second)
			teo = Connect(param, appType, appVersion)
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

	fmt.Print("\0337" + "\033[r" + "\0338") // reset terminal scrolling
}

// Event returns pointer to EventCH channel
func (teo *Teonet) Event() <-chan *EventData {
	return teo.ev.ch
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
				rd = nil
				break FOR
			}
			packet := ev.Data

			// Process trudp events
			switch ev.Event {

			case trudp.EvConnected:
				teolog.Connect(MODULE, "got CONNECTED event, channel key: "+
					string(packet))

			case trudp.EvDisconnected:
				teolog.Connect(MODULE, "got DISCONNECTED event, channel key: "+
					string(packet))
				// Reconnect to r-host
				teo.rhost.reconnect(ev.Tcd)
				// Delete peer from arp table
				teo.arp.deleteKey(string(packet))
				// Close l0 client
				if client, ok := teo.l0.findAddr(string(packet)); ok {
					teo.l0.close(client)
				}

			case trudp.EvResetLocal:
				err = errors.New("got RESET_LOCAL event, channel key: " +
					ev.Tcd.GetKey())
				teolog.Connect(MODULE, err.Error())
				//ev.Tcd.CloseChannel()
				//break FOR

			case trudp.EvGotData, trudp.EvGotDataNotrudp:
				teolog.DebugVvf(MODULE, "got %d bytes packet, channel key: %s\n",
					len(packet), ev.Tcd.GetKey())
				packet, err = teo.cry.decrypt(packet, ev.Tcd.GetKey())
				if err != nil && teo.l0.allow {
					// if packet does not decrypted than it may be l0 client
					// trudp packet. Check l0 packet and process it if this
					// packet is valid teocli(l0) packet
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

			case trudp.EvGotAckPing:
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
	if !teo.running {
		rd = nil
	}
	return
}

// sendToHimself send command to this host
func (teo *Teonet) sendToHimself(to string, cmd byte, data []byte) (length int,
	err error) {
	teolog.DebugVf(MODULE,
		"send command to this host: '%s', cmd: %d, data_len: %d\n",
		to, cmd, len(data),
	)
	rd, err := teo.PacketCreateNew(teo.param.Name, byte(cmd), data).Parse()
	if err != nil {
		err = errors.New("can't parse packet to himself")
		return
	}
	length = len(data)
	teo.com.process(&receiveData{rd, nil})
	return
}

// SendTo send command to Teonet peer
func (teo *Teonet) SendTo(to string, cmd byte, data []byte) (length int,
	err error) {
	arp, ok := teo.arp.m[to]
	if !ok && to != "" {
		err = errors.New("peer " + to + " not connected to this host")
		return
	}
	if arp == nil || arp.tcd == nil {
		return teo.sendToHimself(to, cmd, data)
	}
	return teo.sendToTcd(arp.tcd, cmd, data)
}

// SendToClient send command to Teonet L0 client
func (teo *Teonet) SendToClient(l0Peer string, client string, cmd byte,
	data []byte) (length int, err error) {
	return teo.l0.sendToL0(l0Peer, client, cmd, data)
}

// SendToClientAddr send command to Teonet L0 client by address
func (teo *Teonet) SendToClientAddr(l0 *L0PacketData, client string, cmd byte,
	data []byte) (length int, err error) {
	return teo.sendToClient(l0.addr, l0.port, client, cmd, data)
}

// sendToClient send command to Teonet L0 client by address
func (teo *Teonet) sendToClient(addr string, port int, client string, cmd byte,
	data []byte) (length int, err error) {
	arp, ok := teo.arp.find(addr, port, 0)
	if !ok {
		err = fmt.Errorf("can't find l0 server %s:%d in arp table", addr, port)
		return
	}
	return teo.SendToClient(arp.peer, client, cmd, data)
}

// SendAnswer send (answer) command to Teonet peer by received Packet
func (teo *Teonet) SendAnswer(ipac interface{}, cmd byte, data []byte) (length int,
	err error) {
	pac := ipac.(*Packet)
	if addr, port, ok := pac.L0(); ok {
		if teo.isL0Local(addr, port) {
			return teo.l0.sendTo(teo.param.Name, pac.From(), cmd, data)
		}
		return teo.sendToClient(addr, port, pac.From(), cmd, data)
	}
	return teo.SendTo(pac.From(), cmd, data)
}

// isL0Local return true if the addres and port refered to this server
func (teo *Teonet) isL0Local(addr string, port int) bool {
	return teo.l0.allow &&
		port == teo.param.Port &&
		(addr == "" || addr == "127.0.0.1" || addr == "localhost")
}

// sendAnswer send command to Teonet peer by receiveData
func (teo *Teonet) sendAnswer(rec *receiveData, cmd byte, data []byte) (length int,
	err error) {
	// Answer to peer
	if !rec.rd.IsL0() {
		return teo.sendToTcd(rec.tcd, cmd, data)
	}
	// Answer to L0 client on this or on another L0 server
	addr, port := C.GoString(rec.rd.addr), int(rec.rd.port)
	if teo.isL0Local(addr, port) {
		return teo.l0.sendTo(teo.param.Name, rec.rd.From(), cmd, data)
	} else if length, err = teo.sendToClient(addr, port, rec.rd.From(), cmd, data); err != nil {
		err = errors.New("can't find whom answer to")
	}
	return
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcd(tcd *trudp.ChannelData, cmd byte, data []byte) (length int,
	err error) {

	// makePac creates new teonet packet and show 'send' log message
	makePac := func(tcd *trudp.ChannelData, cmd byte, data []byte) []byte {
		pac := teo.PacketCreateNew(teo.param.Name, cmd, data)
		to, _ := teo.arp.peer(tcd)
		teolog.DebugVf(MODULE, "send cmd: %d, to: %s, data_len: %d\n", cmd, to,
			len(data))
		return teo.cry.encrypt(pac.packet)
	}

	// send splitted packet or send whole packet
	if tcd == nil {
		return teo.sendToHimself(teo.param.Name, cmd, data)
	}
	_, err = teo.split.split(cmd, data, func(cmd byte, data []byte) {
		var l int
		l, err = tcd.Write(makePac(tcd, cmd, data))
		if err != nil {
			return
		}
		length += l
	})

	return
}

// sendToTcd send command to Teonet peer by known trudp channel
func (teo *Teonet) sendToTcdUnsafe(tcd *trudp.ChannelData, cmd byte,
	data []byte) (int, error) {
	pac := teo.PacketCreateNew(teo.param.Name, cmd, data)
	to, _ := teo.arp.peer(tcd)
	teolog.DebugVf(MODULE, "send cmd: %d, to: %s, data_len: %d (send direct udp)\n",
		cmd, to, len(data))
	// \TODO: split data!! We can't split Unsafe packet bekause we can't delivery
	// much unfsafe packets and than combine it. So sugest return err "too large
	// data packet" if the length more than 1024 - 1280 (more than mtu, more than
	// udp packet)... or lets him try send any size packets :-)
	return tcd.WriteUnsafe(teo.cry.encrypt(pac.packet))
}

// Type return this teonet application type (array of types)
func (teo *Teonet) Type() []string {
	// Select this host in arp table
	peerArp, ok := teo.arp.m[teo.param.Name]
	if !ok {
		//err = errors.New("host " + teo.param.Name + " does not exist in arp table")
		return nil
	}
	return peerArp.appType
}

// AppVersion return this teonet application version
func (teo *Teonet) AppVersion() string {
	// Select this host in arp table
	peerArp, ok := teo.arp.m[teo.param.Name]
	if !ok {
		//err = errors.New("host " + teo.param.Name + " does not exist in arp table")
		return ""
	}
	return peerArp.appVersion
}

// Host return host name byte array with leading zerro
func (teo *Teonet) Host() []byte {
	return append([]byte(teo.param.Name), 0)
}

// SetType set this teonet application type (array of types)
func (teo *Teonet) setType(appType []string) (err error) {
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
func (teo *Teonet) setAppVersion(appVersion string) (err error) {
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

// ctrlC process Ctrl+C to close Teonet
func (teo *Teonet) ctrlC() {
	if !teo.param.CtrlcF {
		return
	}
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

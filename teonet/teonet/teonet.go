package teonet

//// CGO definition (don't delay or edit it):
//#cgo LDFLAGS: -lcrypto
//#include <stdlib.h>
//#include "crypt.h"
//#include "net_core.h"
/*
 */
import "C"
import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

// MODULE Teonet module name for using in logging
var MODULE = teokeys.Color(teokeys.ANSILightCyan, "(teonet)")

// Parameters teonet parameters
type Parameters struct {
	Name           string // this host client name
	Port           int    // local port
	RAddr          string // remote host address
	RPort, RChan   int    // remote host port and channel(for TRUdp only)
	Network        string // teonet network name
	LogLevel       string // show log messages level
	ForbidHotkeysF bool   // forbid hotkeys menu
	ShowTrudpStatF bool   // show trudp statistic
	ShowPeersStatF bool   // show peers table
	ShowHelpF      bool   // show usage
}

// Packet is Teonet packet container
type Packet struct {
	packet []byte
}

// packetCreateNew create teonet packet
func packetCreateNew(cmd int, from string, data []byte) (packet *Packet) {
	fromC := C.CString(from)
	var dataC unsafe.Pointer
	var packetLen C.size_t
	var dataLen C.size_t
	if data != nil {
		dataC = unsafe.Pointer(&data[0])
		dataLen = C.size_t(len(data))
	}

	packetC := C.createPacketFrom(C.uint8_t(cmd), fromC, C.size_t(len(from)+1),
		dataC, dataLen, &packetLen)
	pac := C.GoBytes(packetC, C.int(packetLen))
	packet = &Packet{packet: pac}

	C.free(packetC)
	C.free(unsafe.Pointer(fromC))
	return
}

// Len return packet length
func (pac *Packet) Len() int {
	return len(pac.packet)
}

// Cmd return packets cmd number
func (pac *Packet) Cmd() int {
	return int(pac.packet[pac.FromLen()+1])
}

// From return packets from
func (pac *Packet) From() string {
	return C.GoString((*C.char)(unsafe.Pointer(&pac.packet[1])))
}

// FromLen return packets from length
func (pac *Packet) FromLen() int {
	return int(pac.packet[0])
}

// Data return packets data
func (pac *Packet) Data() (data []byte) {
	dataLength := pac.DataLen()
	if dataLength > 0 {
		dataPtr := unsafe.Pointer(&pac.packet[pac.FromLen()+C.PACKET_HEADER_ADD_SIZE])
		data = (*[1 << 28]byte)(dataPtr)[:dataLength:dataLength]
	}
	return
}

// DataLen return packets data len
func (pac *Packet) DataLen() int {
	return len(pac.packet) - pac.FromLen() - C.PACKET_HEADER_ADD_SIZE
}

type receiveData struct {
	rd  *C.ksnCorePacketData
	tcd *trudp.ChannelData
}

// Parse parse teonet packet to 'rd' structure and return it
func (pac *Packet) Parse() (rd *C.ksnCorePacketData, err error) {
	rd = &C.ksnCorePacketData{}
	packetC := unsafe.Pointer(&pac.packet[0])
	if C.parsePacket(packetC, C.size_t(pac.Len()), rd) == 0 {
		err = errors.New("not valid packet")
	}
	return
}

// Packet return packet
func (rd *C.ksnCorePacketData) Packet() (pac *Packet) {
	var data []byte
	dataLength := rd.data_len
	if dataLength > 0 {
		data = (*[1 << 28]byte)(rd.data)[:dataLength:dataLength]
	}
	pac = &Packet{packet: data}
	return
}

// PacketLen return packet length
func (rd *C.ksnCorePacketData) PacketLen() int {
	return int(rd.raw_data_len)
}

// Cmd return rd's cmd number
func (rd *C.ksnCorePacketData) Cmd() int {
	return int(rd.cmd)
}

// From return rd's from
func (rd *C.ksnCorePacketData) From() string {
	return C.GoString(rd.from)
}

// FromLen return rd's from length
func (rd *C.ksnCorePacketData) FromLen() int {
	return int(rd.from_len)
}

// Data return rd's data
func (rd *C.ksnCorePacketData) Data() (data []byte) {
	dataLength := rd.data_len
	if dataLength > 0 {
		data = (*[1 << 28]byte)(rd.data)[:dataLength:dataLength]
	}
	return
}

// Data return rd's data length
func (rd *C.ksnCorePacketData) DataLen() int {
	return int(rd.data_len)
}

// Teonet teonet connection data structure
type Teonet struct {
	td      *trudp.TRUDP        // TRUdp connection
	param   *Parameters         // Teonet parameters
	kcr     *C.ksnCryptClass    // C crypt module
	com     *command            // Commands module
	arp     *arp                // Arp module
	rhost   *rhostData          // R-host module
	running bool                // Teonet running flag
	wg      sync.WaitGroup      // Wait stopped
	ticker  *time.Ticker        // Idle timer ticker (to use in hokeys)
	menu    *teokeys.HotkeyMenu // Hotkey menu
}

// reconnect reconnect to r-host if selected in function parameters channel is
// r-host trudp channel
func (rhost *rhostData) reconnect(tcd *trudp.ChannelData) {
	if rhost.tcd == tcd {
		rhost.wg.Done()
	}
}

// rhostData r-host data
type rhostData struct {
	teo *Teonet            // Teonet connection
	tcd *trudp.ChannelData // TRUDP channel data
	wg  sync.WaitGroup     // Reconnect wait group
}

// Connect initialize Teonet
func Connect(param *Parameters) (teo *Teonet) {

	// Init logger and create Teonet connection structure
	teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	teo = &Teonet{param: param, running: true}

	// Command and Crypto modules init
	teo.com = &command{teo}
	teo.kcr = C.ksnCryptInit(nil)

	// Trudp init
	teo.td = trudp.Init(param.Port)
	teo.td.AllowEvents(1) // \TODO: set events to allow it
	teo.td.ShowStatistic(param.ShowTrudpStatF)

	// Arp module init
	teo.arp = &arp{teo: teo, m: make(map[string]*arpData)}
	teo.arp.peerAdd(param.Name, teo.version())

	// R-host module init and Connect to remote host (r-host)
	teo.rhost = &rhostData{teo: teo}
	if param.RPort > 0 {
		go func() {
			teo.wg.Add(1)
			for teo.running {
				teolog.Connectf(MODULE, "connecting to r-host %s:%d:%d\n", param.RAddr, param.RPort, 0)
				teo.rhost.tcd = teo.td.ConnectChannel(param.RAddr, param.RPort, 0)
				teo.sendToTcd(teo.rhost.tcd, 0, nil)
				teo.rhost.wg.Add(1)
				teo.rhost.wg.Wait()
				time.Sleep(2 * time.Second)
			}
			teo.wg.Done()
		}()
	}

	// Timer ticker
	teo.ticker = time.NewTicker(250 * time.Millisecond)

	// Hotkeys CreateMenu
	if !teo.param.ForbidHotkeysF {
		teo.menu = teokeys.CreateMenu("\bHot keys list:", "")
		teo.menu.Add([]int{'h', '?', 'H'}, "show this help screen", teo.menu.Usage)
		teo.menu.Add('p', "show peers", func() {
			var mode string
			if teo.param.ShowPeersStatF {
				teo.param.ShowPeersStatF = false
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowPeersStatF = true
				teo.param.ShowTrudpStatF = false
				teo.arp.print()
				mode = "on"
			}
			teo.td.ShowStatistic(param.ShowTrudpStatF)
			fmt.Println("\nshow peers", mode)
		})
		teo.menu.Add('u', "show trudp statistics", func() {
			var mode string
			if teo.param.ShowTrudpStatF {
				teo.param.ShowTrudpStatF = false
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowTrudpStatF = true
				teo.param.ShowPeersStatF = false
				mode = "on"
			}
			teo.td.ShowStatistic(param.ShowTrudpStatF)
			fmt.Println("\nshow trudp", mode)
		})
		teo.menu.Add('n', "show 'none' messages", func() {
			fmt.Print("\b")
			param.LogLevel = "NONE"
			teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		})
		teo.menu.Add('c', "show 'connect' messages", func() {
			fmt.Print("\b")
			param.LogLevel = "CONNECT"
			teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		})
		teo.menu.Add('d', "show 'debug' messages", func() {
			fmt.Print("\b")
			param.LogLevel = "DEBUG"
			teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		})
		teo.menu.Add('v', "show 'debug_v' messages", func() {
			fmt.Print("\b")
			param.LogLevel = "DEBUGv"
			teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		})
		teo.menu.Add('w', "show 'debug_vv' messages", func() {
			fmt.Print("\b")
			param.LogLevel = "DEBUGvv"
			teolog.Init(param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
		})
		teo.menu.Add('q', "quit this application", teo.menu.Quit)
	}

	return
}

// Run start Teonet event loop
func (teo *Teonet) Run() {
	go func() {
		teo.wg.Add(1)
		for teo.running {
			rd, err := teo.read()
			if err != nil {
				teolog.Error(MODULE, err)
				continue
			}
			teolog.DebugVf(MODULE, "got packet: cmd %d from %s, data len: %d, data: %v\n",
				rd.Cmd(), rd.From(), len(rd.Data()), rd.Data())
		}
		teo.wg.Done()
	}()
	teo.td.Run()
	teo.running = false
	teo.wg.Wait()
}

// read read and parse network packet
func (teo *Teonet) read() (rd *C.ksnCorePacketData, err error) {
FOR:
	for {
		select {
		// Trudp event
		case ev := <-teo.td.ChanEvent():
			packet := ev.Data

			// Process trudp events
			switch ev.Event {

			case trudp.CONNECTED:
				teolog.Connect(MODULE, "got event: channel with key "+string(packet)+" connected")

			case trudp.DISCONNECTED:
				teolog.Connect(MODULE, "got event: channel with key "+string(packet)+" disconnected")
				teo.rhost.reconnect(ev.Tcd)
				teo.arp.deleteKey(string(packet))

			case trudp.RESET_LOCAL:
				err = errors.New("need to reconnect " + ev.Tcd.GetKey())
				break FOR

			case trudp.GOT_DATA, trudp.GOT_DATA_NOTRUDP:
				teolog.DebugVvf(MODULE, "got %d bytes packet %v\n", len(packet), packet)
				// Decrypt
				var decryptLen C.size_t
				packetPtr := unsafe.Pointer(&packet[0])
				C.ksnDecryptPackage(teo.kcr, packetPtr, C.size_t(len(packet)), &decryptLen)
				if decryptLen > 0 {
					packet = packet[2 : decryptLen+2]
					teolog.DebugVvf(MODULE, "decripted %d bytes packet %v\n", decryptLen, packet)
				} else {
					teolog.DebugVvf(MODULE, "can't decript %d bytes packet (try to use without decrypt)\n", len(packet))
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
					teolog.DebugVv(MODULE, teokeys.Color(teokeys.ANSIRed, "got invalid (not teonet) packet"))
					rd = nil
				}

			case trudp.GOT_ACK_PING:
				triptime, _ := ev.Tcd.GetTriptime()
				teolog.DebugV(MODULE, "got GOT_ACK_PING, key:", ev.Tcd.GetKey(), "triptime:", triptime, "ms")
				teo.arp.print()

			default:
				teolog.Log(teolog.DEBUGvv, MODULE, "got event:", ev.Event)
			}

		// Timer iddle event
		case <-teo.ticker.C:
			//teolog.Debug(MODULE, "ticker event")
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
	pac := packetCreateNew(cmd, teo.param.Name, data)
	// \TODO: encrypt data
	return tcd.WriteTo(pac.packet)
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

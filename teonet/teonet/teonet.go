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
	"log"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

var MODULE = teokeys.Color(teokeys.ANSILightCyan, "(teonet)")

// Parameters
type Parameters struct {
	Name           string // this host client name
	Port           int    // local port
	RAddr          string // remote host address
	RPort, RChan   int    // remote host port and channel(for TRUdp only)
	Network        string // teonet network name
	ShowLogLevel   string // show log messages level
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
	td    *trudp.TRUDP     // TRUdp connection
	param *Parameters      // Teonet parameters
	kcr   *C.ksnCryptClass // C crypt module
	com   *command         // Commands module
	arp   *arp             // Arp module
}

// Connect initialize Teonet
func Connect(param *Parameters) (teo *Teonet) {
	teolog.Init(param.ShowLogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	teo = &Teonet{param: param}
	teo.com = &command{teo}
	teo.kcr = C.ksnCryptInit(nil)
	teo.td = trudp.Init(param.Port)
	teo.td.ShowStatistic(param.ShowTrudpStatF)
	teo.arp = &arp{teo: teo, m: make(map[string]*arpData)}
	teo.arp.peerAdd(param.Name)
	// Connect to remote host (r-host)
	if param.RPort > 0 {
		tcd := teo.td.ConnectChannel(param.RAddr, param.RPort, 0)
		teo.sendToTcd(tcd, 0, nil)
	}
	return
}

// Run start Teonet event loop
func (teo *Teonet) Run() {
	go func() {
		for {
			rd, err := teo.read()
			if err != nil {
				teolog.Error(MODULE, err)
				//break
				continue
			}
			teolog.DebugVf(MODULE, "got packet: cmd %d from %s, data len: %d, data: %v\n",
				rd.Cmd(), rd.From(), len(rd.Data()), rd.Data())
		}
	}()
	teo.td.Run()
}

// read read and parse network packet
func (teo *Teonet) read() (rd *C.ksnCorePacketData, err error) {
FOR:
	for {
		ev := <-teo.td.ChanEvent()
		packet := ev.Data

		// Process trudp events
		switch ev.Event {

		case trudp.CONNECTED:
			teolog.Connect(MODULE, "got event: channel with key "+string(packet)+" connected")
			//break FOR

		case trudp.DISCONNECTED:
			teolog.Connect(MODULE, "got event: channel with key "+string(packet)+" disconnected")
			teo.arp.deleteKey(string(packet))

		case trudp.RESET_LOCAL:
			err = errors.New("need to reconnect")
			break FOR

		case trudp.GOT_DATA, trudp.GOT_DATA_NOTRUDP:
			teolog.DebugVvf(MODULE, "got %d bytes packet %v\n", len(packet), packet)
			var decryptLen C.size_t
			packetPtr := unsafe.Pointer(&packet[0])
			C.ksnDecryptPackage(teo.kcr, packetPtr, C.size_t(len(packet)), &decryptLen)
			if decryptLen > 0 {
				packet = packet[2 : decryptLen+2]
				teolog.DebugVvf(MODULE, "decripted %d bytes packet %v\n", decryptLen, packet)
			}
			pac := &Packet{packet: packet}
			if rd, err = pac.Parse(); rd != nil {
				teolog.DebugVvf(MODULE, "got valid packet cmd: %d, name: %s, data_len: %d\n", pac.Cmd(), pac.From(), pac.DataLen())
				// \TODO don't return error on Parse err != nil, because error is interpreted as disconnect
				if !teo.com.process(&receiveData{rd, ev.Tcd}) {
					break FOR
				}
			} else {
				teolog.Error(MODULE, "got invalid packet")
			}

		default:
			teolog.Log(teolog.DEBUGvv, MODULE, "got event:", ev.Event)
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

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
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// Version Teonet version
const Version = "3.0.0"

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
	teo = &Teonet{param: param}
	teo.com = &command{teo}
	teo.kcr = C.ksnCryptInit(nil)
	teo.td = trudp.Init(param.Port)
	teo.td.ShowStatistic(param.ShowTrudpStatF)
	teo.arp = &arp{teo: teo, m: make(map[string]*arpData)}
	//ShowLogLevel
	teolog.Level(param.ShowLogLevel, true, log.LstdFlags|log.Lmicroseconds)
	if param.RPort > 0 {
		tcd := teo.td.ConnectChannel(param.RAddr, param.RPort, 0)
		teo.sendToTcd(tcd, 0, nil) //[]byte{0})
	}
	return
}

// Run start Teonet event loop
func (teo *Teonet) Run() {
	go func() {
		for {
			rd, err := teo.read()
			if err != nil {
				fmt.Println(err)
				//break
				continue
			}
			fmt.Printf("got packet: cmd %d from %s, data len: %d, data: %v\n",
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
			fmt.Println("got event: channel with key " + string(packet) + " connected")
			//break FOR

		case trudp.DISCONNECTED:
			fmt.Println("got event: channel with key " + string(packet) + " disconnected")
			// \TODO: remove peer from arp map connected to this channel

		case trudp.RESET_LOCAL:
			err = errors.New("need to reconnect")
			break FOR

		case trudp.GOT_DATA, trudp.GOT_DATA_NOTRUDP:
			fmt.Printf("got %d bytes packet %v\n", len(packet), packet)
			var decryptLen C.size_t
			C.ksnDecryptPackage(teo.kcr, unsafe.Pointer(&packet[0]), C.size_t(len(packet)),
				&decryptLen)
			if decryptLen > 0 {
				packet = packet[2 : decryptLen+2]
			}
			fmt.Printf("got(decripted) %d bytes packet %v\n", decryptLen, packet)
			pac := &Packet{packet: packet}
			fmt.Printf("(before parse) cmd: %d, name: %s, name len: %d\n", pac.Cmd(), pac.From(), pac.FromLen())
			if rd, err = pac.Parse(); rd != nil {
				// \TODO don't return error on Parse err != nil, because error is interpreted as disconnect
				if !teo.com.process(&receiveData{rd, ev.Tcd}) {
					break FOR
				}
			}

		default:
			fmt.Println("got event:", ev.Event)
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

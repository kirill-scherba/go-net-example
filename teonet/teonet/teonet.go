package teonet

//// CGO definition (don't delay or edit it):
//#include <stdlib.h>
//#include "net_core.h"
/*
 */
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

const Version = "3.0.0"

// Teonet packet container
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

// Parse parse teonet packet to 'rd' structure and return it
func (pac *Packet) Parse() (rd *C.ksnCorePacketData) {
	rd = &C.ksnCorePacketData{}
	packetC := unsafe.Pointer(&pac.packet[0])
	C.parsePacket(packetC, C.size_t(pac.Len()), rd)
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

type Teonet struct {
	td    *trudp.TRUDP
	name  string // this host name
	raddr string // r-host address
	rport int    // r-host port
}

var tcd *trudp.ChannelData

// Connect initialize Teonet
func Connect(name string, port int, raddr string, rport int) (teo *Teonet) {
	teo = &Teonet{name: name, raddr: raddr, rport: rport}
	teo.td = trudp.Init(port)
	if rport > 0 {
		tcd = teo.td.ConnectChannel(raddr, rport, 0)
		teo.SendTo(name, 0, []byte{0}) //[]byte(name))
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
				break
			}
			fmt.Printf("got cmd %d from %s, data len: %d, data: %v\n",
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
		switch ev.Event {
		case trudp.DISCONNECTED:
			err = errors.New("channel with key " + string(packet) + " disconnected")
			break FOR
		case trudp.RESET_LOCAL:
			err = errors.New("need to reconnect")
			break FOR
		case trudp.GOT_DATA:
			fmt.Printf("got %d bytes packet %v\n", len(packet), packet)
			pac := &Packet{packet: packet}
			fmt.Printf("cmd: %d, name: %s, name len: %d\n", pac.Cmd(), pac.From(), pac.FromLen())
			if rd = pac.Parse(); rd != nil {
				break FOR
			}
		default:
			fmt.Println("got event:", ev.Event)
		}
	}
	return
}

// SendTo
func (teo *Teonet) SendTo(to string, cmd int, data []byte) (err error) {
	pac := packetCreateNew(cmd, teo.name, data)
	return tcd.WriteTo(pac.packet)
}

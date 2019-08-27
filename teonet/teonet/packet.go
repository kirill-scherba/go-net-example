package teonet

// Packed and receive packet module

//#include <stdlib.h> /* free */
//#include "packet.h"
import "C"
import (
	"errors"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// packetCreateNew create teonet packet
func (teo *Teonet) packetCreateNew(from string, cmd byte, data []byte) (packet *Packet) {
	fromC := C.CString(from)
	var dataC unsafe.Pointer
	var packetLen C.size_t
	var dataLen C.size_t
	if data != nil && len(data) > 0 {
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

// Packet is Teonet packet container
type Packet struct {
	packet []byte
}

// Len return packet length
func (pac *Packet) Len() int {
	return len(pac.packet)
}

// Cmd return packets cmd number
func (pac *Packet) Cmd() byte {
	return pac.packet[pac.FromLen()+1]
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

// receiveData recived data structure
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
func (rd *C.ksnCorePacketData) Cmd() byte {
	return byte(rd.cmd)
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

func (rd *C.ksnCorePacketData) IsL0() bool {
	return rd.l0_f != 0
}

// setL0 sets l0 flag and l0 server address to received data
func (rd *C.ksnCorePacketData) setL0(addr string, port int) {
	addrB := append([]byte(addr), 0)
	addrC := (*C.char)(unsafe.Pointer(&addrB[0]))
	rd.addr = addrC
	rd.port = C.int(port)
	rd.l0_f = 1
}

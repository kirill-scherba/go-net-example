package teonet

// Packed and receive packet module

//#include <stdlib.h> /* free */
//#include "packet.h"
import "C"
import (
	"errors"
	"unsafe"

	"github.com/kirill-scherba/teonet-go/trudp/trudp"
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

// Packet is Teonet packet data and method receiver
type Packet struct {
	packet []byte
	l0     *L0PacketData
}

// l0PacketData is l0 data of Teonet packet
type L0PacketData struct {
	addr string
	port int
	ok   bool
}

// Addr l0 address getter
func (l *L0PacketData) Addr() string {
	return l.addr
}

// Port l0 port getter
func (l *L0PacketData) Port() int {
	return l.port
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

// L0 return l0 server address, port and ok == true if packer recived from l0 client
func (pac *Packet) L0() (addr string, port int, ok bool) {
	addr = pac.l0.addr
	port = pac.l0.port
	ok = pac.l0.ok
	return
}

// L0 return packets l0 structure
func (pac *Packet) GetL0() *L0PacketData {
	return pac.l0
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
	dataLength := rd.raw_data_len
	if dataLength > 0 {
		data = (*[1 << 28]byte)(rd.raw_data)[:dataLength:dataLength]
	}
	pac = &Packet{data, &L0PacketData{addr: C.GoString(rd.addr), port: int(rd.port), ok: rd.l0_f != 0}}
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

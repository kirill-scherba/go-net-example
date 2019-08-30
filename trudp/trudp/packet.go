package trudp

//// CGO definition (don't delay or edit it):
//#include <stdlib.h>
//#include "packet.h"
import "C"
import (
	"unsafe"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// \TODO rename to trudpPacket
type packetType struct {
	trudp      *TRUDP
	data       []byte
	sendQueueF bool // true - save to send queue (Data packet); false - don't save to send queue (Service packet)
	destoryF   bool // true - created in C and need to destroy; false - not need to destroy
}

// goBytesUnsafe makes a Go byte slice from a C array (without copying the original data)
func goBytesUnsafe(data unsafe.Pointer, length C.size_t) []byte {
	return (*[1 << 28]byte)(data)[:length:length]
}

// getTimestamp return current 32 bit timestamp in thousands of milliseconds (uSec)
func (trudp *TRUDP) getTimestamp() uint32 {
	return uint32(C.trudpGetTimestamp())
}

// dataCreateNew creates DATA package, it should be free with freeCreated
func (pac *packetType) dataCreateNew(id uint32, channel int, data []byte) *packetType {
	var length C.size_t
	packet := C.trudpPacketDATAcreateNew(C.uint32_t(id), C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), sendQueueF: true, destoryF: true}
}

// pingCreateNew Create PING package, it should be free with freeCreated
func (pac *packetType) pingCreateNew(channel int, data []byte) *packetType {
	var length C.size_t
	packet := C.trudpPacketPINGcreateNew(0, C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// ackCreateNew Create ACK to data package, it should be free with freeCreated
func (pac *packetType) ackCreateNew() *packetType {
	packetPtr := unsafe.Pointer(&pac.data[0])
	packet := C.trudpPacketACKcreateNew(packetPtr)
	length := C.trudpPacketGetHeaderLength(packetPtr)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// ackToPingCreateNew Create ACK to ping package, it should be free with freeCreated
func (pac *packetType) ackToPingCreateNew() *packetType {
	packetPtr := unsafe.Pointer(&pac.data[0])
	headerLength := C.trudpPacketGetHeaderLength(packetPtr)
	packet := C.trudpPacketACKtoPINGcreateNew(packetPtr)
	length := C.size_t(int(headerLength) + len(pac.getData()))
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// resetCreateNew Create RSET package, it should be free with freeCreated
func (pac *packetType) resetCreateNew() *packetType {
	packet := C.trudpPacketRESETcreateNew(0, C.uint(pac.getChannel()))
	length := C.trudpPacketGetHeaderLength(nil)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// ackToResetCreateNew Create ACK to reset package, it should be free with freeCreated
func (pac *packetType) ackToResetCreateNew() *packetType {
	packetPtr := unsafe.Pointer(&pac.data[0])
	packet := C.trudpPacketACKtoRESETcreateNew(packetPtr)
	length := C.trudpPacketGetHeaderLength(packetPtr)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// freeCreated frees packet created with functions dataCreateNew, pingCreateNew
// ackCreateNew or resetCreateNew
func (pac *packetType) freeCreated(packet []byte) {
	C.trudpPacketCreatedFree(unsafe.Pointer(&packet[0]))
}

// destroy packet
func (pac *packetType) destroy() {
	if pac.destoryF {
		pac.freeCreated(pac.data)
	}
}

// writeTo send packetData to trudp channel. Depend on type of created packet:
// Data or Service. Send Data packet to trudp channel and save it to sendQueue
// or Send Service packet to trudp channel and destroy it
func (pac *packetType) writeTo(tcd *ChannelData) {
	pac.trudp.proc.chanWriter <- &writerType{pac, tcd.addr}
	teolog.DebugVf(MODULE, "send %s packet id: %d, to channel: %s\n",
		pac.getTypeString(), pac.getID(), tcd.GetKey())
	if pac.sendQueueF {
		tcd.sendQueueAdd(pac)
		tcd.stat.send(len(pac.data))
		//tcd.trudp.sendEvent(tcd, SEND_DATA, pac.getData())
	}
}

// Check TR-UDP packet and return true if packet valid
func (pac *packetType) check(packet []byte) bool {
	return int(C.trudpPacketCheck(unsafe.Pointer(&packet[0]), C.size_t(len(packet)))) != 0
}

// getChannel return trudp packet channel number
func (pac *packetType) getChannel() int {
	return int(C._trudpPacketGetChannel(unsafe.Pointer(&pac.data[0])))
}

// getID reurn packet id
func (pac *packetType) getID() uint32 {
	return uint32(C.trudpPacketGetId(unsafe.Pointer(&pac.data[0])))
}

// getType return packet type
func (pac *packetType) getType() int {
	return int(C.trudpPacketGetType(unsafe.Pointer(&pac.data[0])))
}

// getTypeStr return packet type in string format
// DATA(0x0), ACK(0x1), RESET(0x2), ACK_RESET(0x3), PING(0x4), ACK_PING(0x5)
func (pac *packetType) getTypeString() string {
	switch int(C.trudpPacketGetType(unsafe.Pointer(&pac.data[0]))) {
	case 0:
		return "DATA"
	case 1:
		return "ACK"
	case 2:
		return "RESET"
	case 3:
		return "ACK_RESET"
	case 4:
		return "PING"
	case 5:
		return "ACK_PING"
	default:
		return "UNKNOWN"
	}
}

// getData return trudp packet data
func (pac *packetType) getData() []byte {
	return pac.data[int(C.trudpPacketGetHeaderLength(unsafe.Pointer(&pac.data[0]))):]
}

// getTimestamp return Timestamp (32 byte) contains sending time of DATA and RESET messages
func (pac *packetType) getTimestamp() uint32 {
	return uint32(C.trudpPacketGetTimestamp(unsafe.Pointer(&pac.data[0])))
}

// getTriptime return packets triptime
func (pac *packetType) getTriptime() (triptime float32) {
	triptime = float32(pac.trudp.getTimestamp()-pac.getTimestamp()) / 1000.0
	return
}

// copy trudp packet
func (pac *packetType) copy() *packetType {
	return &packetType{trudp: pac.trudp, data: append([]byte(nil), pac.data...)}
}

package trudp

//// CGO definition (don't delay or edit it):
//#include <stdlib.h>
//#include "packet.h"
import "C"
import (
	"unsafe"
)

// \TODO remane to trudpPacket
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
func (pac *packetType) dataCreateNew(id uint, channel int, data []byte) *packetType {
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
func (pac *packetType) ackCreateNew(packetInput []byte) *packetType {
	packetPtr := unsafe.Pointer(&packetInput[0])
	packet := C.trudpPacketACKcreateNew(packetPtr)
	length := C.trudpPacketGetHeaderLength(packetPtr)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// ackToPingCreateNew Create ACK to ping package, it should be free with freeCreated
func (pac *packetType) ackToPingCreateNew(packetInput []byte) *packetType {
	packetPtr := unsafe.Pointer(&packetInput[0])
	headerLength := C.trudpPacketGetHeaderLength(packetPtr)
	packet := C.trudpPacketACKtoPINGcreateNew(packetPtr)
	length := C.size_t(int(headerLength) + len(pac.getData(packetInput)))
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// resetCreateNew Create RSET package, it should be free with freeCreated
func (pac *packetType) resetCreateNew(channel int) *packetType {
	packet := C.trudpPacketRESETcreateNew(0, C.uint(channel))
	length := C.trudpPacketGetHeaderLength(nil)
	return &packetType{trudp: pac.trudp, data: goBytesUnsafe(packet, length), destoryF: true}
}

// ackToResetCreateNew Create ACK to reset package, it should be free with freeCreated
func (pac *packetType) ackToResetCreateNew(packetInput []byte) *packetType {
	packetPtr := unsafe.Pointer(&packetInput[0])
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
func (packet *packetType) destroy() {
	if packet.destoryF {
		packet.freeCreated(packet.data)
	}
}

// writeTo send packetData to trudp channel. Depend on type of created packet:
// Data or Service. Send Data packet to trudp channel and save it to sendQueue
// or Send Service packet to trudp channel and destroy it
func (packet *packetType) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	if packet.sendQueueF {
		tcd.sendQueueProcess(func() { tcd.sendQueueAdd(packet) })
	} else {
		packet.destroy()
	}
}

// Check TR-UDP packet and return true if packet valid
func (pac *packetType) check(packet []byte) bool {
	return int(C.trudpPacketCheck(unsafe.Pointer(&packet[0]), C.size_t(len(packet)))) != 0
}

// getChannel return trudp packet channel number
func (pac *packetType) getChannel(packet []byte) int {
	return int(C._trudpPacketGetChannel(unsafe.Pointer(&packet[0])))
}

// getID reurn packet id
func (pac *packetType) getID(packet []byte) uint {
	return uint(C.trudpPacketGetId(unsafe.Pointer(&packet[0])))
}

// getType reurn packet type
func (pac *packetType) getType(packet []byte) int {
	return int(C.trudpPacketGetType(unsafe.Pointer(&packet[0])))
}

// getData return trudp packet data
func (pac *packetType) getData(packet []byte) []byte {
	return packet[int(C.trudpPacketGetHeaderLength(unsafe.Pointer(&packet[0]))):]
}

// getTimestamp return Timestamp (32 byte) contains sending time of DATA and RESET messages
func (pac *packetType) getTimestamp(packet []byte) uint32 {
	return uint32(C.trudpPacketGetTimestamp(unsafe.Pointer(&packet[0])))
}

// getTriptime return packet triptime
func (pac *packetType) getTriptime(packet []byte) (triptime float32) {
	triptime = float32(pac.trudp.getTimestamp()-pac.getTimestamp(packet)) / 1000.0
	return
}

package trudp

//// CGO definition (don't delay or edit it):
//#include <stdlib.h>
//#include "packet.h"
import "C"
import (
	"net"
	"unsafe"
)

type packetType struct {
	trudp *TRUDP
}

// goBytesUnsafe makes a Go byte slice from a C array (without copying the original data)
func goBytesUnsafe(data unsafe.Pointer, length C.size_t) []byte {
	return (*[1 << 28]byte)(data)[:length:length]
}

// getTimestamp return current 32 bit timestamp in thousands of milliseconds (uSec)
func (trudp *TRUDP) getTimestamp() uint32 {
	return uint32(C.trudpGetTimestamp())
}

// autodestroy is helper to destroy packet for xxxCreateNew functions
// func (pac *packetType) autodestroy(sliceOrig []byte) (slice []byte, destroy func()) {
// 	slice = sliceOrig
// 	destroy = func() { pac.freeCreated(slice) }
// 	return
// }

// dataCreateNew creates DATA package, it should be free with freeCreated
func (pac *packetType) dataCreateNew(id uint, channel int, data []byte) []byte {
	var length C.size_t
	packet := C.trudpPacketDATAcreateNew(C.uint32_t(id), C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
	return goBytesUnsafe(packet, length)
}

// pingCreateNew Create PING package, it should be free with freeCreated
func (pac *packetType) pingCreateNew(channel int, data []byte) []byte {
	var length C.size_t
	packet := C.trudpPacketPINGcreateNew(0, C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
	return goBytesUnsafe(packet, length)
}

// ackCreateNew Create ACK to data package, it should be free with freeCreated
func (pac *packetType) ackCreateNew(packetInput []byte) []byte {
	packetPtr := unsafe.Pointer(&packetInput[0])
	packet := C.trudpPacketACKcreateNew(packetPtr)
	length := C.trudpPacketGetHeaderLength(packetPtr)
	return goBytesUnsafe(packet, length)
}

// ackToPingCreateNew Create ACK to ping package, it should be free with freeCreated
func (pac *packetType) ackToPingCreateNew(packetInput []byte) []byte {
	packetPtr := unsafe.Pointer(&packetInput[0])
	headerLength := C.trudpPacketGetHeaderLength(packetPtr)
	packet := C.trudpPacketACKtoPINGcreateNew(packetPtr)
	length := C.size_t(int(headerLength) + len(pac.getData(packetInput)))
	return goBytesUnsafe(packet, length)
}

// freeCreated frees packet created with functions dataCreateNew, pingCreateNew
// ackCreateNew or resetCreateNew
func (pac *packetType) freeCreated(packet []byte) {
	C.trudpPacketCreatedFree(unsafe.Pointer(&packet[0]))
}

// Check TR-UDP packet and return true if packet valid
func (pac *packetType) check(packet []byte) bool {
	return int(C.trudpPacketCheck(unsafe.Pointer(&packet[0]), C.size_t(len(packet)))) != 0
}

// writeTo writes to address and free packet
func (pac *packetType) writeTo(tcd *channelData, packet []byte, addr net.Addr, addToSendQueueF bool) {
	pac.trudp.conn.WriteTo(packet, addr)
	if addToSendQueueF {
		tcd.sendQueueAdd(packet)
	} else {
		pac.freeCreated(packet)
	}
}

// getChannel return trudp packet channel number
func (pac *packetType) getChannel(packet []byte) int {
	return int(C._trudpPacketGetChannel(unsafe.Pointer(&packet[0])))
}

// getId reurn packet id
func (pac *packetType) getId(packet []byte) uint {
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

package trudp

//// CGO definition (don't delay or edit it):
//#include <stdlib.h>
//#include "packet.h"
import "C"
import "unsafe"

type packetType struct{}

// getTimestamp return current 32 bit timestamp in thousands of milliseconds (uSec)
func (trudp *TRUDP) getTimestamp() uint32 {
	return uint32(C.trudpGetTimestamp())
}

// goBytesUnsafe makes a Go byte slice from a C array (without copying the original data)
func goBytesUnsafe(data unsafe.Pointer, length C.size_t) []byte {
	return (*[1 << 28]byte)(data)[:length:length]
}

// dataCreateNew creates DATA package
func (pac *packetType) dataCreateNew(id uint, channel int, data []byte) []byte {
	var length C.size_t
	packet := C.trudpPacketDATAcreateNew(C.uint32_t(id), C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
	return goBytesUnsafe(packet, length)
}

// pingCreateNew Create PING package
func (pac *packetType) pingCreateNew(id uint, channel int, data []byte) []byte {
	var length C.size_t
	packet := C.trudpPacketPINGcreateNew(C.uint32_t(id), C.uint(channel),
		unsafe.Pointer(&data[0]), C.size_t(len(data)), &length)
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

package teocli

//// CGO definition (don't delay or edit it):
//#include "teonet_l0_client.h"
/*
int packetGetPeerNameLength(teoLNullCPacket *packet) {
  return packet->peer_name_length;
}
char* packetGetPeerName(teoLNullCPacket *packet) {
  return packet->peer_name;
}
char* packetGetData(teoLNullCPacket *packet) {
  return packet->peer_name + packet->peer_name_length;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type teoLNull struct {
}

// PacketCreate create teonet l0 client packet
func (teocli *teoLNull) PacketCreate(command uint8, peer string, data []byte) (buffer []byte, err error) {

	bufferLength := C.teoLNullHeaderSize() + C.size_t(len(peer)+1+len(data))
	buffer = make([]byte, bufferLength)
	peerC := C.CString(peer)
	lengh := C.teoLNullPacketCreate(unsafe.Pointer(&buffer[0]), C.size_t(bufferLength), C.uint8_t(command), peerC,
		unsafe.Pointer(&data[0]), C.size_t(len(data)))
	C.free(unsafe.Pointer(peerC))
	if int(lengh) != len(buffer) {
		// \TODO: set error
	}

	packetC := (*C.teoLNullCPacket)(unsafe.Pointer(&buffer[0]))
	fmt.Println("PacketCreate:",
		buffer, "\n",
		C.packetGetPeerNameLength(packetC),
		packetC.peer_name_length,
		packetC.header_checksum,
		C.GoString(C.packetGetPeerName(packetC)),
		C.GoString(C.packetGetData(packetC)),
	)
	return
}

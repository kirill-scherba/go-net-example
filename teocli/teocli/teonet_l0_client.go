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
char* packetGetData(void *packetPtr) {
  teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
  return packet->peer_name + packet->peer_name_length;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/udp/trudp"
)

// TeoLNull connection data
type TeoLNull struct {
	tcp bool

	td  *trudp.TRUDP
	tcd *trudp.ChannelData
}

// packetCreate create teonet l0 client packet
func (teocli *TeoLNull) packetCreate(command uint8, peer string, data []byte) (buffer []byte, err error) {
	bufferLength := C.teoLNullHeaderSize() + C.size_t(len(peer)+1+len(data))
	buffer = make([]byte, bufferLength)
	peerC := C.CString(peer)
	defer C.free(unsafe.Pointer(peerC))
	lengh := C.teoLNullPacketCreate(unsafe.Pointer(&buffer[0]), C.size_t(bufferLength), C.uint8_t(command), peerC,
		unsafe.Pointer(&data[0]), C.size_t(len(data)))
	if int(lengh) != len(buffer) {
		err = fmt.Errorf("can't create packet: "+
			"the length of created packet %d not equal to packet buffer %d",
			lengh, len(buffer))
	}

	// packetC := (*C.teoLNullCPacket)(unsafe.Pointer(&buffer[0]))
	// fmt.Println("PacketCreate:",
	// 	buffer, "\n",
	// 	C.packetGetPeerNameLength(packetC),
	// 	packetC.peer_name_length,
	// 	packetC.header_checksum,
	// 	C.GoString(C.packetGetPeerName(packetC)),
	// 	C.GoString(C.packetGetData(packetC)),
	// )
	return
}

func (teocli *TeoLNull) packetCreateString(command uint8, peer string, data string) (buffer []byte, err error) {
	return teocli.packetCreate(command, peer, append([]byte(data), 0))
}

func (teocli *TeoLNull) packetCreateLogin(data string) (buffer []byte, err error) {
	return teocli.packetCreateString(0, "", data)
}

// packetCreateEcho create teonet l0 client echo packet
func (teocli *TeoLNull) packetCreateEcho(peer string, msg string) (buffer []byte, err error) {
	bufferLength := C.teoLNullHeaderSize() + C.size_t(len(peer)+1+len(msg)+1) + C.sizeof_int64_t
	buffer = make([]byte, bufferLength)
	peerC := C.CString(peer)
	msgC := C.CString(msg)
	defer func() { C.free(unsafe.Pointer(peerC)); C.free(unsafe.Pointer(msgC)) }()
	lengh := C.teoLNullPacketCreateEcho(unsafe.Pointer(&buffer[0]), C.size_t(bufferLength), peerC, msgC)
	if int(lengh) != len(buffer) {
		err = fmt.Errorf("can't create echo packet: "+
			"the length of created packet %d not equal to packet buffer %d",
			lengh, len(buffer),
		)
	}
	return
}

// send send packet to L0 server
func (teocli *TeoLNull) send(packet []byte) (length int, err error) {
	if teocli.tcp {
		err = errors.New("the teocli.send for TCP is not implemented yet")
	} else {
		teocli.tcd.WriteTo(packet)
	}
	return
}

// teoLNullConnectData * con = teoLNullConnectE(param.tcp_server, param.tcp_port,
// 	event_cb, &param, TCP)

// Connect connect to L0 server
func Connect(addr string, port int, tcp bool) (teo *TeoLNull, err error) {
	teo = &TeoLNull{tcp: tcp}
	if tcp {
		err = errors.New("the teocli.Connect for TCP is not implemented yet")
	} else {
		teo.td = trudp.Init(0)
		teo.tcd = teo.td.ConnectChannel(addr, port, 0)
		go teo.td.Run()
	}
	return
}

// Send send data to L0 server
func (teocli *TeoLNull) Send(command uint8, peer string, data []byte) (int, error) {
	packet, err := teocli.packetCreate(command, peer, data)
	if err != nil {
		return 0, err
	}
	return teocli.send(packet)
}

// SendEcho send echo packet to L0 server
func (teocli *TeoLNull) SendEcho(peer string, msg string) (int, error) {
	packet, err := teocli.packetCreateEcho(peer, msg)
	if err != nil {
		return 0, err
	}
	return teocli.send(packet)
}

// SendLogin send login packet to L0 server
func (teocli *TeoLNull) SendLogin(name string) (int, error) {
	packet, err := teocli.packetCreateLogin(name)
	if err != nil {
		return 0, err
	}
	return teocli.send(packet)
}

// Read wait for receiving data from trudp and return teocli packet
func (teocli *TeoLNull) Read() (packet []byte, err error) {
	if teocli.tcp {
		err = errors.New("the teocli.Read for TCP is not implemented yet")
	} else {
		ev := <-teocli.td.ChanEvent()
		packet = ev.Data
	}
	return
}

// ProccessEchoAnswer parse echo answer packet and return triptime in ms
func (teocli *TeoLNull) ProccessEchoAnswer(packet []byte) int64 {
	packetC := C.packetGetData(unsafe.Pointer(&packet[0]))
	return int64(C.teoLNullProccessEchoAnswer(packetC))
}

package teocli

//// CGO definition (don't delay or edit it):
//#include "teonet_l0_client.h"
/*
int packetGetCommand(void *packetPtr) {
	teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
  return packet->cmd;
}
int packetGetPeerNameLength(teoLNullCPacket *packet) {
  return packet->peer_name_length;
}
int packetGetDataLength(void *packetPtr) {
	teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
  return packet->data_length;
}
int packetGetLength(void *packetPtr) {
  teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
	return teoLNullHeaderSize() + packetGetPeerNameLength(packetPtr) +
	  packetGetDataLength(packetPtr);
}
char* packetGetPeerName(void *packetPtr) {
	teoLNullCPacket *packet = (teoLNullCPacket *)packetPtr;
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
	"net"
	"strconv"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

const (
	// CmdLEcho Echo command
	CmdLEcho = C.CMD_L_ECHO

	// CmdLEchoAnswer Answer to Echo command
	CmdLEchoAnswer = C.CMD_L_ECHO_ANSWER

	// CmdLPeers Get peers command
	CmdLPeers = C.CMD_L_PEERS

	// CmdLPeersAnswer Answer to get peers command
	CmdLPeersAnswer = C.CMD_L_PEERS_ANSWER
)

// TeoLNull teonet l0 client connection data
type TeoLNull struct {
	readBuffer []byte
	tcp        bool

	conn net.Conn // TCP connection

	td  *trudp.TRUDP       // TRUDP connection
	tcd *trudp.ChannelData // TRUDP channel
}

// packetCreate create teonet l0 client packet
func (teocli *TeoLNull) packetCreate(command uint8, peer string, data []byte) (buffer []byte, err error) {
	var dataLen int
	var dataPtr unsafe.Pointer
	if data != nil {
		dataPtr = unsafe.Pointer(&data[0])
		dataLen = len(data)
	}
	bufferLength := C.teoLNullHeaderSize() + C.size_t(len(peer)+1+dataLen)
	buffer = make([]byte, bufferLength)
	peerC := C.CString(peer)
	defer C.free(unsafe.Pointer(peerC))
	lengh := C.teoLNullPacketCreate(unsafe.Pointer(&buffer[0]), C.size_t(bufferLength), C.uint8_t(command), peerC,
		dataPtr, C.size_t(dataLen))
	if int(lengh) != len(buffer) {
		err = fmt.Errorf("can't create packet: "+
			"the length of created packet %d not equal to packet buffer %d",
			lengh, len(buffer))
	}

	return
}

// packetCreateString creates packet with string data
func (teocli *TeoLNull) packetCreateString(command uint8, peer string, data string) (buffer []byte, err error) {
	return teocli.packetCreate(command, peer, append([]byte(data), 0))
}

// packetCreateLogin creates login packet
func (teocli *TeoLNull) packetCreateLogin(data string) (buffer []byte, err error) {
	return teocli.packetCreateString(0, "", data)
}

// packetCreateEcho creates teonet l0 client echo packet
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

// packetCheck check received packet, combine packets and return valid packet
// return Valid packet or nil and status
// status  0 valid packet received
// status -1 packet not received yet (got part of packet)
// status  1 wrong packet received (drop it)
func (teocli *TeoLNull) packetCheck(packet []byte) (retpacket []byte, retval int) {

	// Skip empty packet
	if packet == nil || len(packet) == 0 {
		return
	}

	// Check packet length and checksums and parse return value (0, 1, -1, -2, -3)
	retval = int(C.packetCheck(unsafe.Pointer(&packet[0]), C.size_t(len(packet))))
	//fmt.Println("C.packetCheck(before):", retval, "buffer len:", len(teocli.readBuffer))
	switch {

	// valid packet
	case retval == 0 && len(teocli.readBuffer) == 0:
		//if len(teocli.readBuffer) > 0 {
		teocli.readBuffer = teocli.readBuffer[:0]
		//}
		retpacket = packet

	// First part of splitted packet
	case (retval == -1 || retval == -2) && len(teocli.readBuffer) == 0:
		teocli.readBuffer = append(teocli.readBuffer, packet...)
		retval = -1

	// next part of splitted packet
	case /*(retval == -3 || retval == -2 || retval == -1 || retval == 0) &&*/ len(teocli.readBuffer) > 0:
		teocli.readBuffer = append(teocli.readBuffer, packet...)
		bufPtr := unsafe.Pointer(&teocli.readBuffer[0])
		retval = int(C.packetCheck(bufPtr, C.size_t(len(teocli.readBuffer))))
		//fmt.Println("C.packetCheck(after):", retval, "buffer len:", len(teocli.readBuffer))
		switch retval {
		// valid packet received
		case 0:
			packetLength := C.packetGetLength(bufPtr)
			retpacket = append([]byte(nil), teocli.readBuffer[:packetLength]...)
			teocli.readBuffer = teocli.readBuffer[packetLength:]
		// invalid packet received
		case 1, -3:
			teocli.readBuffer = teocli.readBuffer[0:0]
			retval = 1
		// next part of packet received
		default:
			retval = -1
		}
	}
	//fmt.Println("packetCheck(end):", retval, "buffer len:", len(teocli.readBuffer))
	return
}

// packetGetData return packet data
func (teocli *TeoLNull) packetGetData(packet []byte) (data []byte) {
	packetPtr := unsafe.Pointer(&packet[0])
	dataC := C.packetGetData(packetPtr)
	dataLength := C.packetGetDataLength(packetPtr)
	data = (*[1 << 28]byte)(unsafe.Pointer(dataC))[:dataLength:dataLength]
	return
}

// send send packet to L0 server
func (teocli *TeoLNull) send(packet []byte) (length int, err error) {
	if teocli.tcp {
		length, err = teocli.conn.Write(packet)
	} else {
		length = len(packet)
		for {
			if len(packet) <= 512 {
				err = teocli.tcd.WriteTo(packet)
				break
			} else {
				err = teocli.tcd.WriteTo(packet[:512])
				packet = packet[512:]
			}
		}
	}
	return
}

// sendEchoAnswer send echo answer to echo command
func (teocli *TeoLNull) sendEchoAnswer(packet []byte) (length int, err error) {
	if packet != nil {
		packetPtr := unsafe.Pointer(&packet[0])
		if C.packetGetCommand(packetPtr) == C.CMD_L_ECHO {
			peerC := C.packetGetPeerName(packetPtr)
			data := teocli.packetGetData(packet)
			teocli.Send(C.CMD_L_ECHO_ANSWER, C.GoString(peerC), data)
		}
	}
	return
}

// Connect connect to L0 server
func Connect(addr string, port int, tcp bool) (teo *TeoLNull, err error) {
	teo = &TeoLNull{tcp: tcp, readBuffer: make([]byte, 0)}
	if tcp {
		teo.conn, err = net.Dial("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			return
		}
	} else {
		teo.td = trudp.Init(0)
		teo.tcd = teo.td.ConnectChannel(addr, port, 0)
		go teo.td.Run()
	}
	return
}

// Disconnect from L0 server
func (teocli *TeoLNull) Disconnect() {
	if teocli.tcp {
		teocli.conn.Close()
	} else {
		teocli.td.ChanEventClosed()
		teocli.td.Close()
	}
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
func (teocli *TeoLNull) Read() (pac *Packet, err error) {
	packetCheck := func(packet []byte) (pac *Packet) {
		packet, _ = teocli.packetCheck(packet)
		if packet != nil {
			teocli.sendEchoAnswer(packet)
			pac = teocli.packetNew(packet)
		}
		return
	}
FOR:
	for {
		if teocli.tcp {
			packet := make([]byte, 2048)
			length, _ := teocli.conn.Read(packet)
			if length == 0 {
				err = errors.New("server disconnected...")
				break FOR
			}
			packet = packet[:length]
			fmt.Printf("got %d bytes packet\n", len(packet))
			if pac = packetCheck(packet); pac != nil {
				break FOR
			}
		} else {
			ev := <-teocli.td.ChanEvent()
			packet := ev.Data
			switch ev.Event {
			case trudp.DISCONNECTED:
				err = errors.New("channel with key " + string(packet) + " disconnected")
				break FOR
			case trudp.RESET_LOCAL:
				err = errors.New("need to reconnect")
				break FOR
			case trudp.GOT_DATA:
				fmt.Printf("got %d bytes packet\n", len(packet))
				if pac = packetCheck(packet); pac != nil {
					break FOR
				}
			default:
				fmt.Println("got event:", ev.Event)
			}
		}
	}
	return
}

// proccessEchoAnswer parse echo answer packet and return triptime in ms
func (teocli *TeoLNull) proccessEchoAnswer(packet []byte) (t int64, err error) {
	packetPtr := unsafe.Pointer(&packet[0])
	command := C.packetGetCommand(packetPtr)
	if command != C.CMD_L_ECHO_ANSWER {
		err = errors.New("wrong packets command number")
		return
	}
	dataC := C.packetGetData(packetPtr)
	t = int64(C.teoLNullProccessEchoAnswer(dataC))
	return
}

// Packet is teocli packet data structure and container for methods
type Packet struct {
	packet []byte
	teocli *TeoLNull
}

// packetNew Creates new packet from packet slice
func (teocli *TeoLNull) packetNew(packet []byte) *Packet {
	return &Packet{packet: packet, teocli: teocli}
}

// Command return packets peer name
func (pac *Packet) Command() int {
	return int(C.packetGetCommand(unsafe.Pointer(&pac.packet[0])))
}

// Name return packets peer name
func (pac *Packet) Name() string {
	return C.GoString(C.packetGetPeerName(unsafe.Pointer(&pac.packet[0])))
}

// From return packets peer name (sinonim for Name nethod)
func (pac *Packet) From() string {
	return pac.Name()
}

// Data return packets data
func (pac *Packet) Data() []byte {
	return pac.teocli.packetGetData(pac.packet)
}

// TripTime return triptime for echo answer packet
func (pac *Packet) TripTime() (int64, error) {
	return pac.teocli.proccessEchoAnswer(pac.packet)
}

// PeersLength return number of peers in peerAnswer packet
func (pac *Packet) PeersLength() int {
	dataPtr := unsafe.Pointer(&pac.Data()[0])
	arpDataAr := (*C.ksnet_arp_data_ar)(dataPtr)
	return int(arpDataAr.length)
}

// Peers return string representation of peerAnswer packet
func (pac *Packet) Peers() string {
	dataPtr := unsafe.Pointer(&pac.Data()[0])
	arpDataAr := (*C.ksnet_arp_data_ar)(dataPtr)
	buf := C.arp_data_print(arpDataAr)
	defer C.free(unsafe.Pointer(buf))
	return C.GoString(buf)
}

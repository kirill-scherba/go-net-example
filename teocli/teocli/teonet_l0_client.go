// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocli

//// CGO definition (don't delay or edit it):
//#include <string.h>
//#include "teonet_l0_client.h"
/*
uint8_t packetGetCommand(void *packetPtr) {
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
	"math"
	"net"
	"strconv"
	"unsafe"

	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

// Version teocli version
const Version = "3.0.0"

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

// PacketCreate create teonet l0 client packet
func (teocli *TeoLNull) PacketCreate(command uint8, peer string, data []byte) (buffer []byte, err error) {
	var dataLen int
	var dataPtr unsafe.Pointer
	if data != nil && len(data) > 0 {
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
	return teocli.PacketCreate(command, peer, append([]byte(data), 0))
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

// PacketCheck check received packet, combine packets and return valid packet.
// After "valid packet received" run this function again to get next valid packed.
// return Valid packet or nil and status
// status  0 valid packet received
// status -1 packet not received yet (got part of packet)
// status  1 wrong packet received (drop it)
func (teocli *TeoLNull) PacketCheck(packet []byte) (retpacket []byte, retval int) {

	// Skip empty packet
	if len(teocli.readBuffer) == 0 && (packet == nil || len(packet) == 0) {
		retval = -1
		return
	}

	// Check packet length and checksums and parse return value (0, 1, -1, -2, -3)
	var packetPtr unsafe.Pointer
	if packet != nil && len(packet) > 0 {
		packetPtr = unsafe.Pointer(&packet[0])
	}
	retval = int(C.packetCheck(packetPtr, C.size_t(len(packet))))
	//fmt.Println("C.packetCheck(before):", retval, "buffer len:", len(teocli.readBuffer))
	switch {

	// valid packet
	case retval == 0 && len(teocli.readBuffer) == 0:
		//if len(teocli.readBuffer) > 0 {
		//teocli.readBuffer = teocli.readBuffer[:0]
		//}
		packetLength := C.packetGetLength(packetPtr)
		retpacket = packet[0:packetLength]
		teocli.readBuffer = packet[packetLength:]

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

// ResetReadBuf reset read buffer
func (teocli *TeoLNull) ResetReadBuf() {
	teocli.readBuffer = nil
}

// // PacketGetCmd return packets command
// func (teocli *TeoLNull) PacketGetCommand(packet []byte) (cmd int) {
// 	packetPtr := unsafe.Pointer(&packet[0])
// 	cmd = int(C.packetGetCommand(packetPtr))
// 	return
// }

// PacketGetData return packets data
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
				_, err = teocli.tcd.Write(packet)
				break
			} else {
				_, err = teocli.tcd.Write(packet[:512])
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
			teocli.SendTo(C.GoString(peerC), C.CMD_L_ECHO_ANSWER, data)
		}
	}
	return
}

// Init initialize teocli
func Init(tcp bool) (teo *TeoLNull, err error) {
	teo = &TeoLNull{tcp: tcp, readBuffer: make([]byte, 0)}
	return
}

// Connect connect to L0 server
func Connect(addr string, port int, tcp bool) (teo *TeoLNull, err error) {
	teo, err = Init(tcp)
	if tcp {
		teo.conn, err = net.Dial("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			return
		}
	} else {
		localport := 0
		teo.td = trudp.Init(&localport)
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
func (teocli *TeoLNull) SendTo(peer string, command byte, data []byte) (int, error) {
	packet, err := teocli.PacketCreate(command, peer, data)
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

// Read wait for receiving data from tcp or trudp and return teocli packet
func (teocli *TeoLNull) Read() (pac *Packet, err error) {
	packetCheck := func(packet []byte) (pac *Packet) {
		packet, _ = teocli.PacketCheck(packet)
		if packet != nil {
			teocli.sendEchoAnswer(packet)
			pac = teocli.PacketNew(packet)
		}
		return
	}
FOR:
	for {
		// Get next valid packet from teocli read buffer
		if pac = packetCheck(nil); pac != nil {
			break FOR
		}
		// Get received data from tcp or udp
		if teocli.tcp {
			packet := make([]byte, 2048)
			length, _ := teocli.conn.Read(packet)
			if length == 0 {
				err = errors.New("server disconnected")
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

// PacketNew Creates new packet from packet slice
func (teocli *TeoLNull) PacketNew(packet []byte) *Packet {
	return &Packet{packet: packet, teocli: teocli}
}

// Command return packets peer name
func (pac *Packet) Command() byte {
	return byte(C.packetGetCommand(unsafe.Pointer(&pac.packet[0])))
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
	if len(pac.Data()) == 0 {
		return ""
	}
	dataPtr := unsafe.Pointer(&pac.Data()[0])
	arpDataAr := (*C.ksnet_arp_data_ar)(dataPtr)
	buf := C.arp_data_print(arpDataAr)
	defer C.free(unsafe.Pointer(buf))
	return C.GoString(buf)
}

// PeerData create arp peer bynary data
func PeerData(mode int, peer, addr string, port int, triptime float32) (d []byte) {

	// Fill ksnet_arp data structure
	arpData := &C.ksnet_arp_data{}
	arpDataLen := C.sizeof_ksnet_arp_data
	arpData.mode = C.int16_t(mode)
	if len(addr) > 0 {
		C.memcpy(unsafe.Pointer(&arpData.addr[0]), unsafe.Pointer(&[]byte(addr)[0]),
			C.size_t(len(addr)))
	}
	arpData.port = C.int16_t(port)
	arpData.last_triptime = C.double(triptime)

	// Peer name
	cname := []byte(peer)
	if l := len(cname); l < int(C.ARP_TABLE_IP_SIZE) {
		cname = append(cname, make([]byte, C.ARP_TABLE_IP_SIZE-l)...)
	} else {
		cname = cname[:C.ARP_TABLE_IP_SIZE]
	}

	// Create slice from unsafe C raw pointer (the data does not copy)
	d = (*[1 << 28]byte)(unsafe.Pointer(arpData))[:arpDataLen:arpDataLen]
	d = append(cname, d...)

	return
}

// ParsePeerData parse peer data from binary buffer
func ParsePeerData(d []byte) (mode int, peer, addr string, port int, triptime float32) {
	cpeer := (*C.char)(unsafe.Pointer(&d[0]))
	dptr := unsafe.Pointer(&d[C.ARP_TABLE_IP_SIZE])
	arpData := (*C.ksnet_arp_data)(dptr)
	mode = int(arpData.mode)
	peer = C.GoString(cpeer)
	addr = C.GoString(&arpData.addr[0])
	port = int(arpData.port)
	triptime = float32(math.Round(float64(arpData.last_triptime)*1000) / 1000)
	//triptime = float32(arpData.last_triptime)
	return
}

// PeerDataLength return length of binary PeerData buffer
func PeerDataLength() int {
	return C.ARP_TABLE_IP_SIZE + C.sizeof_ksnet_arp_data
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Packet module of teocli packag

package teocli

// #cgo CFLAGS: -g -Wall
// #include "packet.h"
import "C"

import (
	"errors"
	"unsafe"
)

// Packet is teocli packet data structure and container for methods
type Packet struct {
	packet []byte
}

// NewPacket Creates new teocli packet from packet slice
func (teocli *TeoLNull) NewPacket(packet []byte) *Packet {
	return &Packet{packet: packet}
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
	packetPtr := unsafe.Pointer(&pac.packet[0])
	dataC := C.packetGetData(packetPtr)
	dataLength := C.packetGetDataLength(packetPtr)
	return (*[1 << 28]byte)(unsafe.Pointer(dataC))[:dataLength:dataLength]
}

// Triptime return triptime for echo answer packet
func (pac *Packet) Triptime() (t int64, err error) {
	packetPtr := unsafe.Pointer(&pac.packet[0])
	command := C.packetGetCommand(packetPtr)
	if command != C.CMD_L_ECHO_ANSWER {
		err = errors.New("wrong packet command number")
		return
	}
	dataC := C.packetGetData(packetPtr)
	t = int64(C.teoLNullProccessEchoAnswer(dataC))
	return
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

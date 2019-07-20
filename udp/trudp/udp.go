// Copyright 2019 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This module contain functions and structures to work with UDP

package trudp

import (
	"fmt"
	"net"
	"strconv"
)

type udp struct {
	conn *net.UDPConn // Listner UDP address
}

// resolveAddr returns an address of UDP end point
func (udp *udp) resolveAddr(network, address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(network, address)
}

// listen Connect to UDP with selected port (the port incremented if busy)
func (udp *udp) listen(port int) *net.UDPConn {

	// Combine service from host name and port
	service := hostName + ":" + strconv.Itoa(port)

	// Resolve the UDP address so that we can make use of ListenUDP
	// with an actual IP and port instead of a name (in case a
	// hostname is specified).
	udpAddr, err := udp.resolveAddr(network, service)
	if err != nil {
		panic(err)
	}

	// Start listen UDP port
	udp.conn, err = net.ListenUDP(network, udpAddr)
	if err != nil {
		port++
		fmt.Println("the", port-1, "is busy, try next port:", port)
		udp.conn = udp.listen(port)
	}

	return udp.conn
}

// localAddr return string with udp local address
func (udp *udp) localAddr() string {
	return udp.conn.LocalAddr().String()
}

// readFrom acts like ReadFrom but returns a UDPAddr.
func (udp *udp) readFrom(b []byte) (int, *net.UDPAddr, error) {
	return udp.conn.ReadFromUDP(b)
}

// WriteToUDP acts like WriteTo but takes a UDPAddr.
func (udp *udp) writeTo(b []byte, addr *net.UDPAddr) (int, error) {
	return udp.conn.WriteToUDP(b, addr)
}

// Copyright 2019 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This module contain functions and structures to work with UDP

package trudp

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
)

const USESYSCALL = true

type udp struct {
	conn *net.UDPConn // Listner UDP address
	fd   int          // Listen UDP syscall socket
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

	// Start listen UDP port (using go net library)
	fn := func() {
		udp.conn, err = net.ListenUDP(network, udpAddr)
		if err != nil {
			port++
			fmt.Println("the", port-1, "is busy, try next port:", port)
			udp.conn = udp.listen(port)
		}
	}
	//fn()

	// Create UDP syscall socket
	fs := func() {
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			panic(err)
		}

		//  Bind UDP socket to local port so we can receive pings
		if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: port, Addr: [4]byte{0, 0, 0, 0}}); err != nil {
			port++
			fmt.Println("the", port-1, "is busy, try next port:", port)
			udp.conn = udp.listen(port)
		}
	}
	//fs()

	if !USESYSCALL {
		fs()
	} else {
		fn()
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

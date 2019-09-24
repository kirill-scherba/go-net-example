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

// USESYSCALL flag select udp libray or syscall (may be syscall will be
// deprekated bekause library wotk good)
const USESYSCALL = false

type udp struct {
	conn *net.UDPConn // Listner UDP address

	fd   int                   // Listen UDP syscall socket
	addr syscall.SockaddrInet4 // Listen UDP syscall port
}

// resolveAddr returns an address of UDP end point
func (udp *udp) resolveAddr(network, address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(network, address)
}

// listen Connect to UDP with selected port (the port incremented if busy)
func (udp *udp) listen(port *int) *net.UDPConn {

	// Combine service from host name and port
	service := hostName + ":" + strconv.Itoa(*port)

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
			*port++
			fmt.Println("the", *port-1, "is busy, try next port:", *port)
			udp.conn = udp.listen(port)
		}
	}

	// Create UDP syscall socket
	fs := func() {
		udp.fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			panic(err)
		}

		// Bind UDP socket to local port so we can receive pings
		udp.addr = syscall.SockaddrInet4{Port: *port, Addr: [4]byte{}}
		if err := syscall.Bind(udp.fd, &udp.addr); err != nil {
			*port++
			fmt.Println("the", *port-1, "is busy, try next port:", port)
			udp.conn = udp.listen(port)
		}
	}

	// Use syscall or net packet. If USESYSCALL = true than syscall used
	if USESYSCALL {
		fs()
	} else {
		fn()
	}

	// If input faunction paameter port was 0 than get it from connection for
	// future use
	if *port == 0 {
		*port = udp.conn.LocalAddr().(*net.UDPAddr).Port
	}

	return udp.conn
}

// localAddr return string with udp local address
func (udp *udp) localAddr() string {
	if udp.conn != nil {
		return udp.conn.LocalAddr().String()
	}
	var str string
	for i := 0; i < 4; i++ {
		if i > 0 {
			str += "."
		}
		str += strconv.Itoa(int(udp.addr.Addr[i]))
	}
	return str + ":" + strconv.Itoa(udp.addr.Port)
}

// readFrom acts like ReadFrom but returns a UDPAddr.
func (udp *udp) readFrom(b []byte) (int, *net.UDPAddr, error) {
	if udp.conn != nil {
		return udp.conn.ReadFromUDP(b)
	}
	n, addr, err := syscall.Recvfrom(udp.fd, b, 0)
	a := addr.(*syscall.SockaddrInet4)
	return n, &net.UDPAddr{IP: a.Addr[:], Port: a.Port}, err
}

// WriteToUDP acts like WriteTo but takes a UDPAddr.
func (udp *udp) writeTo(b []byte, addr *net.UDPAddr) (int, error) {
	if udp.conn != nil {
		return udp.conn.WriteToUDP(b, addr)
	}
	a := &syscall.SockaddrInet4{Port: addr.Port}
	copy(a.Addr[:], addr.IP)
	err := syscall.Sendto(udp.fd, b, 0, a)
	return len(b), err
}

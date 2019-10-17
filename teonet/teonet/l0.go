// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server module.
//
// L0 server is intended to connect teonet clients to the teonet network. Teonet
// clients should use uses teocli packet which connect it to teonet L0 server.
// The L0 server allow tcp or trudp connection which incuded to teocli packet.
// This module contain data and methods to realise L0 server functions.

package teonet

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

const notL0ServerError = "can't process this command because I'm not L0 server"

// l0Conn is Module data structure
type l0Conn struct {
	teo     *Teonet            // Pointer to Teonet
	stat    *l0Stat            // Statistic
	auth    *l0AuthCom         // Authentication
	param   *paramConf         // Config parameters
	allow   bool               // Allow L0 Server
	wsAllow bool               // Allow L0 websocket server
	wsConn  *wsConn            // Websocket server connector
	wsPort  int                // Websocket TCP port (if 0 - not allowed websocket)
	tcpPort int                // TCP port (if 0 - not allowed TCP)
	conn    net.Listener       // TCP listener connection
	ch      chan *packet       // Packet processing channel
	ma      map[string]*client // Clients address map
	mn      map[string]*client // Clients name map
	mux     sync.Mutex         // Maps mutex
	closed  bool               // Closet flag
}

// packet is Packet processing channels data structure
type packet struct {
	packet []byte
	client *client
}

// conn is an interface to make one parameter for tcp 'conn net.Conn' and
// trudp '*trudp.ChannelData' connection
type conn interface {
	Write([]byte) (int, error)
	Close() error
}

// l0New initialize l0 module
func (teo *Teonet) l0New() *l0Conn {
	l0 := &l0Conn{
		teo:     teo,                 // Pointer to Teonet
		allow:   teo.param.L0allow,   // Allow udp and tcp server(if tcp port > 0)
		tcpPort: teo.param.L0tcpPort, // Allow tcp server(if allow is true)
		wsAllow: teo.param.L0wsAllow, // Allow websocket server(if websocket port > 0)
		wsPort:  teo.param.L0wsPort,  // Allow websocket server(if wsAllow is true)
	}

	if l0.allow || l0.wsAllow {
		l0.stat = l0.statNew()        // Staistic module
		l0.auth = l0.authNew()        // Authenticate module
		l0.param = l0.parametersNew() // Configuration parameters module

		// Start L0 pocessing
		l0.ma = make(map[string]*client)
		l0.mn = make(map[string]*client)
		l0.process()
		// Start udp l0 server
		if l0.allow {
			teolog.Connect(MODULE, "l0 server start listen udp port:", l0.teo.param.Port)
		}
		// Start tcp l0 server
		if l0.tcpPort > 0 {
			l0.tcpServer(&l0.tcpPort)
			teo.param.L0tcpPort = l0.tcpPort
		}
		// Start websocket l0 server
		if l0.wsAllow && l0.wsPort > 0 {
			l0.wsConn = l0.wsServe(l0.wsPort)
		}
	}
	return l0
}

// destroy l0 module
func (l0 *l0Conn) destroy() {
	if l0.allow {
		l0.closeAll()
		teolog.Connect(MODULE, "l0 server stop listen udp port:", l0.teo.param.Port)
		if l0.conn != nil {
			l0.conn.Close()
			l0.conn = nil
		}
		if !l0.closed {
			close(l0.ch)
			l0.closed = true
		}
		if l0.wsConn != nil {
			l0.wsConn.destroy()
		}
	}
}

// network return network type of conn: 'tcp' or' trudp' (in string)
func (l0 *l0Conn) network(client *client) (network string) {
	switch client.conn.(type) {
	case net.Conn:
		network = "tcp"
	case *trudp.ChannelData:
		network = "trudp"
	case *wsHandlerConn:
		network = "ws"
	}
	return
}

// tcpServer TCP L0 server
func (l0 *l0Conn) tcpServer(port *int) {

	const (
		network  = "tcp"
		hostName = ""
	)

	var err error

	// Start tcp server
	for {
		l0.conn, err = net.Listen(network, hostName+":"+strconv.Itoa(*port))
		if err == nil {
			break
		}
		*port++
		fmt.Println("the", *port-1, "is busy, try next port:", *port)
	}
	// If input function parameter port was 0 than get it from connection for
	// future use
	if *port == 0 {
		*port = l0.conn.Addr().(*net.TCPAddr).Port
	}
	teolog.Connect(MODULE, "l0 server start listen tcp port:", *port)

	// Listen for an incoming connection
	go func(port int) {
		for {
			conn, err := l0.conn.Accept()
			if err != nil {
				//teolog.Debug(MODULE, "stop accepting: ", err.Error())
				break
			}
			// Handle connections in a new goroutine.
			go l0.handleConnection(conn)
		}
		teolog.Connect(MODULE, "l0 server stop listen tcp port:", port)
	}(*port)
}

// Handle TCP connection
func (l0 *l0Conn) handleConnection(conn net.Conn) {
	teolog.Connectf(MODULE, "l0 server tcp client %v connected...", conn.RemoteAddr())
	cli, _ := teocli.Init(true)
	b := make([]byte, 2048)
	for {
		n, err := conn.Read(b)
		if err != nil {
			break
		}
		teolog.DebugVvf(MODULE, "got %d bytes data from tcp clien: %v\n",
			n, conn.RemoteAddr().String())
		l0.packetCheck(cli, conn.RemoteAddr().String(), conn, b[:n])
	}
	teolog.Connectf(MODULE, "l0 server tcp client %v disconnected...", conn.RemoteAddr())
	if !l0.closeAddr(conn.RemoteAddr().String()) {
		conn.Close()
	}
}

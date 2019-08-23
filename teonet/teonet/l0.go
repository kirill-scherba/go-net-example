package teonet

import (
	"bufio"
	"fmt"
	"net"
	"strconv"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// Teonet L0 server module

type l0 struct {
	teo   *Teonet      // Pointer to Teonet
	allow bool         // Allow L0 Server
	port  int          // TCP port (if 0 - not allowed TCP)
	conn  net.Listener // TCP listener connection
}

// l0New initialize l0 module
func (teo *Teonet) l0New() *l0 {
	l0 := &l0{teo: teo, allow: teo.param.L0allow, port: teo.param.L0tcpPort}
	if l0.allow {
		teolog.Connect(MODULE, "l0 server start listen udp port:", l0.teo.param.Port)
		//if l0.port > 0 {
		l0.tspServer(&l0.port)
		teo.param.L0tcpPort = l0.port
		//}
	}
	return l0
}

// destroy destroys l0 module
func (l0 *l0) destroy() {
	if l0.allow {
		teolog.Connect(MODULE, "l0 server stop listen udp port:", l0.teo.param.Port)
		if l0.conn != nil {
			l0.conn.Close()
			//teolog.Connect(MODULE, "l0 server stop listen tcp port:", l0.teo.param.L0tcpPort)
		}
	}
}

// tspServer TCP L0 server
func (l0 *l0) tspServer(port *int) {

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
				fmt.Println("Error accepting: ", err.Error())
				//os.Exit(1)
				break
			}
			// Handle connections in a new goroutine.
			go l0.handleConnection(conn)
		}
		teolog.Connect(MODULE, "l0 server stop listen tcp port:", port)
	}(*port)
}

// Handle TCP connection
func (l0 *l0) handleConnection(conn net.Conn) {
	teolog.Connectf(MODULE, "l0 server tcp client %v connected...", conn.RemoteAddr())
	conn.Write([]byte("HTTP/1.1 200 OK\nContent-Type: text/html\n\n<html><body>Hello!</body></html>\n"))
	bufReader := bufio.NewReader(conn)
	buf := make([]byte, 2048)
	for {
		n, err := bufReader.Read(buf)
		if err != nil {
			break
		}
		b := buf[:n]
		teolog.Debugf(MODULE,
			"got %d bytes data from tcp clien, data_len: %d, data: %v\n",
			n, len(b), b)
	}
	teolog.Connectf(MODULE, "l0 server tcp client %v disconnected...", conn.RemoteAddr())
	conn.Close()
}

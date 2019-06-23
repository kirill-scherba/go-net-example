package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	maxBufferSize = 2048
	echoMsg       = "Ping!"
	helloMsg      = "Hello!"
	echoAnswerMsg = "Pong!"
	pingInterval  = 1000 // ms
)

func main() {
	fmt.Println("UDP teset application ver 1.0.0")

	var (
		rhost string
		rport int
		port  int
	)

	flag.IntVar(&rport, "r", 9010, "remote host port (to connect to remote host)")
	flag.StringVar(&rhost, "a", "", "remove host address (to connect to remote host)")
	flag.IntVar(&port, "p", 9000, "this host port (to remote hosts connect to this host)")
	flag.Parse()

	hostName := ""
	network := "udp"
	service := hostName + ":" + strconv.Itoa(port)

	buffer := make([]byte, maxBufferSize)

	// Resolve the UDP address so that we can make use of ListenUDP
	// with an actual IP and port instead of a name (in case a
	// hostname is specified).
	udpAddr, err := net.ResolveUDPAddr(network, service)
	if err != nil {
		panic(err)
	}

	// Start listen UDP port
	conn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("start listenning at", udpAddr)

	// Connect to remote host if it defined:
	// Send hello to remote host
	if rhost != "" {
		service := rhost + ":" + strconv.Itoa(rport)
		rUDPAddr, err := net.ResolveUDPAddr(network, service)
		if err != nil {
			panic(err)
		}
		fmt.Println("connecting to rhost", rUDPAddr)
		conn.WriteToUDP([]byte(helloMsg), rUDPAddr)

		// Keep alive: send Ping
		go func(conn *net.UDPConn) {
			for {
				time.Sleep(pingInterval * time.Millisecond)
				dt, _ := time.Now().MarshalBinary()
				conn.WriteToUDP(append([]byte(echoMsg), dt...), rUDPAddr)
			}
		}(conn)
	}

	// Wait some data received from UDP port
	for {
		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}

		// Process connect message
		if nRead == len(helloMsg) && string(buffer[:len(helloMsg)]) == helloMsg {
			fmt.Println("got", nRead, "bytes 'connect' message from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
			continue
		}

		// Process echo message Ping (send to Pong)
		if nRead > len(echoMsg) && string(buffer[:len(echoMsg)]) == echoMsg {
			fmt.Println("got", nRead, "byte 'ping' command from:", addr, buffer[:nRead])
			conn.WriteToUDP(append([]byte(echoAnswerMsg), buffer[len(echoMsg):nRead]...), addr.(*net.UDPAddr))
			continue
		}

		// Process echo answer message Pong (answer to Ping)
		if nRead > len(echoAnswerMsg) && string(buffer[:len(echoAnswerMsg)]) == echoAnswerMsg {
			var ts time.Time
			ts.UnmarshalBinary(buffer[len(echoAnswerMsg):nRead])
			fmt.Println("got", nRead, "byte 'pong' command from:", addr, "trip time:", time.Since(ts), buffer[:nRead])
			continue
		}

		// Process othe messages
		fmt.Println("got", nRead, "bytes from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
	}
}

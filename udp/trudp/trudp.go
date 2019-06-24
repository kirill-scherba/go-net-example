package trudp

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	maxBufferSize = 2048 // bytes
	pingInterval  = 1000 // ms
	echoMsg       = "ping"
	helloMsg      = "hello"
	echoAnswerMsg = "pong"

	hostName = ""
	network  = "udp"
)

// This packet log function
func log(p ...interface{}) {
	fmt.Println(p...)
}

// TRUDP connection strucure
type TRUDP struct {
	conn   *net.UDPConn
	ticker *time.Ticker
}

// Init start trudp connection
func Init(port int) (retval *TRUDP) {

	// Combine service from host name and port
	service := hostName + ":" + strconv.Itoa(port)

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

	// Start ticker
	ticker := time.NewTicker(pingInterval * time.Millisecond)

	log("start listenning at", udpAddr)

	retval = &TRUDP{
		conn:   conn,
		ticker: ticker,
	}

	retval.ticerCheck()

	return
}

// TRUDP Ticker
func (trudp *TRUDP) ticerCheck() {
	go func() {
		for t := range trudp.ticker.C {
			log("tick at", t)
		}
	}()
}

// Run waits some data received from UDP port and procces it
func (trudp *TRUDP) Run() {

	for {
		buffer := make([]byte, maxBufferSize)

		nRead, addr, err := trudp.conn.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}

		// Process connect message
		if nRead == len(helloMsg) && string(buffer[:len(helloMsg)]) == helloMsg {
			log("got", nRead, "bytes 'connect' message from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
			continue
		}

		// Process echo message Ping (send to Pong)
		if nRead > len(echoMsg) && string(buffer[:len(echoMsg)]) == echoMsg {
			log("got", nRead, "byte 'ping' command from:", addr, buffer[:nRead])
			trudp.conn.WriteToUDP(append([]byte(echoAnswerMsg), buffer[len(echoMsg):nRead]...), addr.(*net.UDPAddr))
			continue
		}

		// Process echo answer message Pong (answer to Ping)
		if nRead > len(echoAnswerMsg) && string(buffer[:len(echoAnswerMsg)]) == echoAnswerMsg {
			var ts time.Time
			ts.UnmarshalBinary(buffer[len(echoAnswerMsg):nRead])
			log("got", nRead, "byte 'pong' command from:", addr, "trip time:", time.Since(ts), buffer[:nRead])
			continue
		}

		// Process other messages
		log("got", nRead, "bytes from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
	}
}

// Connect to remote host
func (trudp *TRUDP) Connect(rhost string, rport int) {

	service := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := net.ResolveUDPAddr(network, service)
	if err != nil {
		panic(err)
	}
	log("connecting to rhost", rUDPAddr)

	// Send hello to remote host
	trudp.conn.WriteToUDP([]byte(helloMsg), rUDPAddr)

	// Keep alive: send Ping
	go func(conn *net.UDPConn) {
		for {
			time.Sleep(pingInterval * time.Millisecond)
			dt, _ := time.Now().MarshalBinary()
			conn.WriteToUDP(append([]byte(echoMsg), dt...), rUDPAddr)
		}
	}(trudp.conn)

}

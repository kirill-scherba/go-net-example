package trudp

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	maxResendAttempt = 50   // (number) max number of resend packet from sendQueue
	maxBufferSize    = 2048 // (bytes) send buffer size in bytes
	pingInterval     = 1000 // (ms) send ping afret in ms
	disconnectAfter  = 3000 // (ms) disconnect afret in ms
	defaultRTT       = 20   // (ms) default retransmit time in ms
	firstPacketID    = 0    // (number) first packet ID and first expectedID number

	helloMsg      = "hello"
	echoMsg       = "ping"
	echoAnswerMsg = "pong"

	network  = "udp"
	hostName = ""
)

// TRUDP connection strucure
type TRUDP struct {
	conn     *net.UDPConn            // connector to send data
	ticker   *time.Ticker            // timer ticler
	logLevel int                     // trudp log level
	logLog   bool                    // show time in trudp log
	tcdmap   map[string]*channelData // channel data map
	packet   *packetType             // packet functions holder
}

// listenUDP Connect to UDP with selected port (the port incremented if busy)
func listenUDP(port int) *net.UDPConn {

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
		port++
		fmt.Println("the", port-1, "is busy, try next port:", port)
		conn = listenUDP(port)
	}

	return conn
}

// TRUDP Ticker
func (trudp *TRUDP) tickerCheck() {
	go func() {
		for t := range trudp.ticker.C {
			trudp.log(DEBUGvv, "tick at", t)
		}
	}()
}

// Init start trudp connection
func Init(port int) (trudp *TRUDP) {

	// Connect to UDP
	conn := listenUDP(port)

	// Start ticker
	ticker := time.NewTicker(pingInterval * time.Millisecond)

	trudp = &TRUDP{
		conn:     conn,
		ticker:   ticker,
		logLevel: CONNECT,
		logLog:   false,
		packet:   &packetType{},
	}
	trudp.tcdmap = make(map[string]*channelData)
	trudp.packet.trudp = trudp

	trudp.log(CONNECT, "start listenning at", conn.LocalAddr())
	trudp.tickerCheck()

	return
}

// Connect to remote host by UDP
func (trudp *TRUDP) Connect(rhost string, rport int) {

	service := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := net.ResolveUDPAddr(network, service)
	if err != nil {
		panic(err)
	}
	trudp.log(CONNECT, "connecting to host", rUDPAddr)

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

// Run waits some data received from UDP port and procces it
func (trudp *TRUDP) Run() {

	for {
		buffer := make([]byte, maxBufferSize)

		nRead, addr, err := trudp.conn.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}
		if nRead == 0 {
			trudp.log(DEBUGv, "empty paket received from:", addr)
			continue
		}

		// Process connect message
		if nRead == len(helloMsg) && string(buffer[:len(helloMsg)]) == helloMsg {
			trudp.log(DEBUG, "got", nRead, "bytes 'connect' message from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
			continue
		}

		// Process echo message Ping (send to Pong)
		if nRead > len(echoMsg) && string(buffer[:len(echoMsg)]) == echoMsg {
			trudp.log(DEBUG, "got", nRead, "byte 'ping' command from:", addr, buffer[:nRead])
			trudp.conn.WriteToUDP(append([]byte(echoAnswerMsg), buffer[len(echoMsg):nRead]...), addr.(*net.UDPAddr))
			continue
		}

		// Process echo answer message Pong (answer to Ping)
		if nRead > len(echoAnswerMsg) && string(buffer[:len(echoAnswerMsg)]) == echoAnswerMsg {
			var ts time.Time
			ts.UnmarshalBinary(buffer[len(echoAnswerMsg):nRead])
			trudp.log(DEBUG, "got", nRead, "byte 'pong' command from:", addr, "trip time:", time.Since(ts), buffer[:nRead])
			continue
		}

		// Check trudp packet
		if trudp.packet.check(buffer[:nRead]) {
			ch := trudp.packet.getChannel(buffer[:nRead])
			id := trudp.packet.getID(buffer[:nRead])
			tp := trudp.packet.getType(buffer[:nRead])
			data := trudp.packet.getData(buffer[:nRead])
			trudp.log(DEBUGvv, "got trudp packet from:", addr, "data:", data, string(data),
				", channel:", ch, "packet id:", id, "type:", tp)

			trudp.packet.process(buffer[:nRead], addr)
			continue
		}

		// Process other messages
		trudp.log(DEBUG, "got", nRead, "bytes from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
	}
}

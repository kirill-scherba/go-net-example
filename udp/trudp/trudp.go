package trudp

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"
)

const (
	maxResendAttempt = 50   // (number) max number of resend packet from sendQueue
	maxBufferSize    = 2048 // (bytes) send buffer size in bytes
	pingInterval     = 1000 // (ms) send ping afret in ms
	disconnectAfter  = 3000 // (ms) disconnect afret in ms
	defaultRTT       = 30   // (ms) default retransmit time in ms
	firstPacketID    = 0    // (number) first packet ID and first expectedID number
	statInterval     = 100  // Interval to show statistic
	DefaultQueueSize = 16   // Default size of send and receive queue
	chWriteSize      = 1024 //2048 // Size of write channele used to send messages from users level

	helloMsg      = "hello"
	echoMsg       = "ping"
	echoAnswerMsg = "pong"

	network  = "udp"
	hostName = ""
)

// TRUDP connection strucure
type TRUDP struct {

	// Listner address
	conn *net.UDPConn // connector to send data

	// Control maps, channels and function holder
	tcdmap map[string]*channelData // channel data map
	chRead chan *eventData         // User level event channel
	packet *packetType             // packet functions holder
	ticker *time.Ticker            // timer ticler

	// Logger configuration
	logLevel int  // trudp log level
	logLogF  bool // show time in trudp log

	// Statustic
	startTime time.Time   // TRUDP start running time
	packets   packetsStat // TRUDP packets statistic

	defaultQueueSize int

	// Control Flags
	showStatF bool // Show statistic
}

// trudpStat structure contain trudp statistic variables
type packetsStat struct {
	send          uint32        // Total packets send
	sendLength    uint64        // Total send in bytes
	ack           uint32        // Total ACK reseived
	receive       uint32        // Total packet reseived
	receiveLength uint64        // Total reseived in bytes
	dropped       uint32        // Total packet droped
	repeat        uint32        // Total packet repeated
	sendRT        realTimeSpeed // Type to calculate send real time speed
	receiveRT     realTimeSpeed // Type to calculate receive real time speed
}

// eventData used as structure in sendEvent function
type eventData struct {
	Tcd   *channelData
	Event int
	Data  []byte
}

// Enumeration of TRUDP events
const (

	/**
	 * Initialize TR-UDP event
	 * @param td Pointer to trudpData
	 */
	INITIALIZE = iota

	/**
	 * Destroy TR-UDP event
	 * @param td Pointer to trudpData
	 */
	DESTROY

	/**
	 * TR-UDP channel disconnected event
	 * @param data NULL
	 * @param data_length 0
	 * @param user_data NULL
	 */
	CONNECTED

	/**
	 * TR-UDP channel disconnected event
	 * @param data Last packet received
	 * @param data_length 0
	 * @param user_data NULL
	 */
	DISCONNECTED

	/**
	 * Got TR-UDP reset packet
	 * @param data NULL
	 * @param data_length 0
	 * @param user_data NULL
	 */
	GOT_RESET

	/**
	 * Send TR-UDP reset packet
	 * @param data Pointer to uint32_t send id or NULL if received id = 0
	 * @param data_length Size of uint32_t or 0
	 * @param user_data NULL
	 */
	SEND_RESET

	/**
	 * Got ACK to reset command
	 * @param data NULL
	 * @param data_length 0
	 * @param user_data NULL
	 */
	GOT_ACK_RESET

	/**
	 * Got ACK to ping command
	 * @param data Pointer to ping data (usually it is a string)
	 * @param data_length Length of data
	 * @param user_data NULL
	 */
	GOT_ACK_PING

	/**
	 * Got PING command
	 * @param data Pointer to ping data (usually it is a string)
	 * @param data_length Length of data
	 * @param user_data NULL
	 */
	GOT_PING

	/**
	 * Got ACK command
	 * @param data Pointer to ACK packet
	 * @param data_length Length of data
	 * @param user_data NULL
	 */
	GOT_ACK

	/**
	 * Got DATA
	 * @param data Pointer to data
	 * @param data_length Length of data
	 * @param user_data NULL
	 */
	GOT_DATA

	/**
	 * Process received data
	 * @param tcd Pointer to trudpData
	 * @param data Pointer to receive buffer
	 * @param data_length Receive buffer length
	 * @param user_data NULL
	 */
	PROCESS_RECEIVE

	/** Process received not TR-UDP data
	 * @param tcd Pointer to trudpData
	 * @param data Pointer to receive buffer
	 * @param data_length Receive buffer length
	 * @param user_data NULL
	 */
	PROCESS_RECEIVE_NO_TRUDP

	/** Process send data
	 * @param data Pointer to send data
	 * @param data_length Length of send
	 * @param user_data NULL
	 */
	SEND_DATA

	RESET_LOCAL
)

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
func (trudp *TRUDP) runStatistic() {
	go func() {
		for range trudp.ticker.C {
			if !trudp.showStatF {
				continue
			}
			idx := 0
			t := time.Now()
			//str := tcd.stat.statHeader(time.Since(trudp.startTime))
			var str string

			// Read trudp channels map keys to slice and sort it
			keys := make([]string, len(trudp.tcdmap))
			for key := range trudp.tcdmap {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			// Get trudp channels statistic string by sorted keys
			for _, key := range keys {
				tcd, ok := trudp.tcdmap[key]
				if ok {
					str += tcd.stat.statBody(tcd, idx, 0)
					idx++
				}
			}

			// Get fotter and print statistic string
			tcs := &channelStat{trudp: trudp} // Empty Methods holder
			str = tcs.statHeader(time.Since(trudp.startTime), time.Since(t)) + str + tcs.statFooter(idx)
			fmt.Print(str)
		}
	}()
}

// Init start trudp connection
func Init(port int) (trudp *TRUDP) {

	// Connect to UDP
	conn := listenUDP(port)

	// Start ticker
	ticker := time.NewTicker(statInterval * time.Millisecond)

	trudp = &TRUDP{
		conn:      conn,
		ticker:    ticker,
		logLevel:  CONNECT,
		logLogF:   false,
		packet:    &packetType{},
		startTime: time.Now(),
	}
	trudp.tcdmap = make(map[string]*channelData)
	trudp.chRead = make(chan *eventData, chWriteSize)
	trudp.packet.trudp = trudp

	trudp.Log(CONNECT, "start listenning at", conn.LocalAddr())
	trudp.runStatistic()

	trudp.sendEvent(nil, INITIALIZE, []byte(conn.LocalAddr().String()))

	return
}

// sendEvent Send event to user level (to event callback or channel)
func (trudp *TRUDP) sendEvent(tcd *channelData, event int, data []byte) {
	trudp.chRead <- &eventData{tcd, event, data}
}

// Connect to remote host by UDP
func (trudp *TRUDP) Connect(rhost string, rport int) {

	service := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := net.ResolveUDPAddr(network, service)
	if err != nil {
		panic(err)
	}
	trudp.Log(CONNECT, "connecting to host", rUDPAddr)

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

		nRead, addr, err := trudp.conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}

		switch {
		// Empty packet
		case nRead == 0:
			trudp.Log(DEBUGv, "empty paket received from:", addr)

		// Check trudp packet
		case trudp.packet.check(buffer[:nRead]):
			//trudp.packet.process(buffer[:nRead], addr)
			packet := &packetType{trudp: trudp, data: buffer[:nRead]}
			packet.process(addr)
			// ch := trudp.packet.getChannel(buffer[:nRead])
			// id := trudp.packet.getID(buffer[:nRead])
			// tp := trudp.packet.getType(buffer[:nRead])
			// data := trudp.packet.getData(buffer[:nRead])
			// trudp.Log(DEBUGvv, "got trudp packet from:", addr, "data:", data, string(data),
			// 	", channel:", ch, "packet id:", id, "type:", tp)

		// Process connect message
		case nRead == len(helloMsg) && string(buffer[:len(helloMsg)]) == helloMsg:
			trudp.Log(DEBUG, "got", nRead, "bytes 'connect' message from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))

		// Process echo message Ping (send to Pong)
		case nRead > len(echoMsg) && string(buffer[:len(echoMsg)]) == echoMsg:
			trudp.Log(DEBUG, "got", nRead, "byte 'ping' command from:", addr, buffer[:nRead])
			trudp.conn.WriteToUDP(append([]byte(echoAnswerMsg), buffer[len(echoMsg):nRead]...), addr /*.(*net.UDPAddr)*/)

		// Process echo answer message Pong (answer to Ping)
		case nRead > len(echoAnswerMsg) && string(buffer[:len(echoAnswerMsg)]) == echoAnswerMsg:
			var ts time.Time
			ts.UnmarshalBinary(buffer[len(echoAnswerMsg):nRead])
			trudp.Log(DEBUG, "got", nRead, "byte 'pong' command from:", addr, "trip time:", time.Since(ts), buffer[:nRead])

		// Process other messages
		default:
			trudp.Log(DEBUG, "got", nRead, "bytes from:", addr, "data: ", buffer[:nRead], string(buffer[:nRead]))
		}
	}
}

// ChRead return channel to read trudp events
func (trudp *TRUDP) ChRead() <-chan *eventData {
	return trudp.chRead
}

// ShowStatistic set showStatF to show trudp statistic window
func (trudp *TRUDP) ShowStatistic(showStatF bool) {
	trudp.showStatF = showStatF
}

// SetDefaultQueueSize set maximum send and receive queues size
func (trudp *TRUDP) SetDefaultQueueSize(defaultQueueSize int) {
	trudp.defaultQueueSize = defaultQueueSize
}

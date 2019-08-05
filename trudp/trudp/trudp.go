package trudp

import (
	"net"
	"strconv"
	"time"
)

const (
	maxResendAttempt = 50   // (number) max number of resend packet from sendQueue
	maxBufferSize    = 2048 // (bytes) send buffer size in bytes
	pingInterval     = 1000 // (ms) send ping afret in ms
	disconnectAfter  = 3000 // (ms) disconnect afret in ms
	defaultRTT       = 30   // (ms) default retransmit time in ms
	maxRTT           = 500  // (ms) default maximum time in ms
	firstPacketID    = 0    // (number) first packet ID and first expectedID number
	chReadSize       = 1024 // Size of read channele used to got data from udp
	chWriteSize      = 96   // Size of write channele used to send data from users level and than send it to remote host
	chEventSize      = 96   // Size or read channel used to send messages to user level

	// DefaultQueueSize is size of send and receive queue
	DefaultQueueSize = 96

	helloMsg      = "hello"
	echoMsg       = "ping"
	echoAnswerMsg = "pong"

	// Network time & local host name
	network  = "udp"
	hostName = ""
)

// TRUDP connection strucure
type TRUDP struct {

	// UDP address and functions
	udp *udp

	// Control maps, channels and function holder
	tcdmap    map[string]*ChannelData // channel data map
	chanEvent chan *EventData         // User level event channel
	packet    *packetType             // packet functions holder
	ticker    *time.Ticker            // timer ticler
	proc      *process

	// Logger configuration
	logLevel int  // trudp log level
	logLogF  bool // show time in trudp log

	// Statistic
	startTime time.Time   // TRUDP start running time
	packets   packetsStat // TRUDP packets statistic

	defaultQueueSize int // Default queues size

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
	sendRT        realTimeSpeed // Send real time speed
	receiveRT     realTimeSpeed // Receive real time speed
	repeatRT      realTimeSpeed // Repiat real time speed
}

// eventData used as structure in sendEvent function
type EventData struct {
	Tcd   *ChannelData
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
	//SEND_DATA

	RESET_LOCAL
)

// Init start trudp connection
func Init(port int) (trudp *TRUDP) {

	trudp = &TRUDP{
		udp:              &udp{},
		packet:           &packetType{},
		logLevel:         CONNECT,
		startTime:        time.Now(),
		tcdmap:           make(map[string]*ChannelData),
		chanEvent:        make(chan *EventData, chEventSize),
		defaultQueueSize: DefaultQueueSize,
	}
	trudp.packet.trudp = trudp

	// Connect to UDP and start UDP workers
	trudp.udp.listen(port)
	trudp.proc = new(process).init(trudp)

	localAddr := trudp.udp.localAddr()
	trudp.Log(CONNECT, "start listenning at", localAddr)
	trudp.sendEvent(nil, INITIALIZE, []byte(localAddr))

	return
}

// sendEvent Send event to user level (to event callback or channel)
func (trudp *TRUDP) sendEvent(tcd *ChannelData, event int, data []byte) {
	trudp.chanEvent <- &EventData{tcd, event, data}
}

// Connect to remote host by UDP
func (trudp *TRUDP) Connect(rhost string, rport int) {

	service := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := trudp.udp.resolveAddr(network, service)
	if err != nil {
		panic(err)
	}
	trudp.Log(CONNECT, "connecting to host", rUDPAddr)

	// Send hello to remote host
	trudp.udp.writeTo([]byte(helloMsg), rUDPAddr)

	// Keep alive: send Ping
	go func() {
		for {
			time.Sleep(pingInterval * time.Millisecond)
			dt, _ := time.Now().MarshalBinary()
			trudp.udp.writeTo(append([]byte(echoMsg), dt...), rUDPAddr)
		}
	}()
}

// Run waits some data received from UDP port and procces it
func (trudp *TRUDP) Run() {

	for {
		buffer := make([]byte, maxBufferSize)

		nRead, addr, err := trudp.udp.readFrom(buffer)
		if err != nil {
			trudp.Log(CONNECT, "stop listenning at", trudp.udp.localAddr())
			close(trudp.proc.chanRead)
			trudp.proc.destroy()
			trudp.proc.wg.Wait()
			trudp.Log(CONNECT, "stopped")
			break
		}

		switch {
		// Empty packet
		case nRead == 0:
			trudp.Log(DEBUGv, "empty paket received from:", addr)

		// Check trudp packet
		case trudp.packet.check(buffer[:nRead]):
			packet := &packetType{trudp: trudp, data: buffer[:nRead]}
			trudp.proc.chanRead <- &readType{addr, packet}
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
			trudp.udp.writeTo(append([]byte(echoAnswerMsg), buffer[len(echoMsg):nRead]...), addr)

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

// Running return true if TRUDP is running now
func (trudp *TRUDP) Running() bool {
	return !trudp.proc.stopRunningF
}

// closeChannels Close all trudp channels
func (trudp *TRUDP) closeChannels() {
	for key, tcd := range trudp.tcdmap {
		tcd.destroy(CONNECT, "close "+key)
	}
}

// Close closes trudp connection and channelRead
func (trudp *TRUDP) Close() {
	if trudp.udp.conn != nil {
		trudp.udp.conn.Close()
	}
}

// ChanEvent return channel to read trudp events
func (trudp *TRUDP) ChanEvent() <-chan *EventData {
	trudp.proc.once.Do(func() {
		trudp.proc.wg.Add(1)
	})
	return trudp.chanEvent
}

// ChanEventClosed signalling that event channel reader routine sucessfully closed
func (trudp *TRUDP) ChanEventClosed() {
	trudp.proc.wg.Done()
}

// ShowStatistic set showStatF to show trudp statistic window
func (trudp *TRUDP) ShowStatistic(showStatF bool) {
	trudp.showStatF = showStatF
}

// SetDefaultQueueSize set maximum send and receive queues size
func (trudp *TRUDP) SetDefaultQueueSize(defaultQueueSize int) {
	trudp.defaultQueueSize = defaultQueueSize
}

// GetAddr return IP and Port of local address
func (trudp *TRUDP) GetAddr() (ip string, port int) {
	ip = string(trudp.udp.conn.LocalAddr().(*net.UDPAddr).IP)
	port = trudp.udp.conn.LocalAddr().(*net.UDPAddr).Port
	return
}

package trudp

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TRUDP channel data structure
type channelData struct {
	trudp *TRUDP // link to trudp

	// Channels remote host address and channel number
	addr *net.UDPAddr // UDP address
	ch   int          // TRUDP channel number
	key  string       // TRUDP channel key (address and channel string representation)

	// Channels current IDs
	id         uint32 // Last send packet ID
	expectedID uint32 // Expected incoming ID

	// Channels packet queues
	sendQueue    []sendQueueData    // send queue
	receiveQueue []receiveQueueData // received queue
	writeQueue   []writeType        // write queue
	maxQueueSize int                // maximum queue size

	// Channel channels and waiting groups
	chProcessCommand chan func()           // channel for worker 'process command'
	chWrite          chan []byte           // channel to write (used to send data from user level)
	chStopWorkers    [workersLen]chan bool // channels to stop wokers
	wgWorkers        sync.WaitGroup        // workers stop wait group

	// Channel flags
	stoppedF     bool // TRUDP channel stopped flag
	sendTestMsgF bool // Send test messages
	//showStatF    bool // Show statistic

	// TRUDP channel statistic
	stat channelStat
}

// Workers indexes costant
const (
	wkProcessCommand = iota // Process command, Process sendQueue and Keepalive
	// ksSomeOherWorker may be added here
	workersLen // Number of workers
)

// reset exequte reset of this cannel
func (tcd *channelData) reset() {
	// Clear user write channel
	tcd.resetChWrite()
	// Clear sendQueue
	tcd.sendQueueReset()
	// Clear receivedQueue
	tcd.receiveQueueReset()
	// Clear writeQueue
	tcd.trudp.proc.writeQueueReset(tcd)
	// Set tcd.id = 0
	tcd.id = firstPacketID
	// Set tcd.expectedID = 1
	tcd.expectedID = firstPacketID
	// \TODO reset trudp channel statistic
	// Send event "RESET was applied" to user level
	tcd.trudp.sendEvent(tcd, RESET_LOCAL, nil)
}

// destroy close and destroy trudp channel
func (tcd *channelData) destroy(msgLevel int, msg string) (err error) {

	// Disable repeatable 'destroy'
	if tcd.stoppedF {
		err = errors.New("can't destroy: the channel " + tcd.key + " already closed")
		return
	}

	tcd.stoppedF = true
	tcd.trudp.Log(msgLevel, msg)

	go func() {

		// Stopping workers and
		for idx := range tcd.chStopWorkers {
			tcd.chStopWorkers[idx] <- true
		}

		// Wait to stop workers
		tcd.wgWorkers.Wait()

		// Close workers stop channel and chSendQueue channel
		for idx := range tcd.chStopWorkers {
			close(tcd.chStopWorkers[idx])
		}
		close(tcd.chProcessCommand)

		// Free and close write channel
		tcd.resetChWrite()
		close(tcd.chWrite)

		// Clear channel queues (the receive queue was cleaned during stop workers)
		tcd.receiveQueueReset()

		// Clear write queue
		tcd.trudp.proc.writeQueueReset(tcd)

		// \TODO clear/correct TRUDP statistics data

		// Remove trudp channel from channels map
		delete(tcd.trudp.tcdmap, tcd.key)
		tcd.trudp.Log(CONNECT, "channel with key", tcd.key, "disconnected")
		tcd.trudp.sendEvent(tcd, DISCONNECTED, []byte(tcd.key))
	}()

	return
}

// getId return new packe id
func (tcd *channelData) getID() (id uint32) {
	id = atomic.AddUint32(&tcd.id, 1) - 1
	return
}

// SendTestMsg set sendTestMsgF flag to send test message by interval
func (tcd *channelData) SendTestMsg(sendTestMsgF bool) {
	tcd.sendTestMsgF = sendTestMsgF
}

// TripTime return current triptime (ms)
func (tcd *channelData) TripTime() float32 {
	return tcd.stat.triptime
}

// WriteTo send data to remote host
func (tcd *channelData) WriteTo(data []byte) (err error) {
	if tcd.stoppedF {
		err = errors.New("can't write to: the channel " + tcd.key + " already closed")
		return
	}
	//tcd.chWrite <- data
	chanAnswer := make(chan bool)
	tcd.trudp.proc.chanWrite <- writeType{tcd, data, chanAnswer}
	<-chanAnswer
	return
}

// makeKey return trudp channel key
func (trudp *TRUDP) makeKey(addr net.Addr, ch int) string {
	return addr.String() + ":" + strconv.Itoa(ch)
}

// newChannelData create new TRUDP ChannelData or select existing
func (trudp *TRUDP) newChannelData(addr *net.UDPAddr, ch int) (tcd *channelData, key string) {

	key = trudp.makeKey(addr, ch)

	// Channel data select
	tcd, ok := trudp.tcdmap[key]
	if ok {
		trudp.Log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}

	now := time.Now()

	// Channel data create
	tcd = &channelData{
		trudp:        trudp,
		addr:         addr,
		ch:           ch,
		key:          key,
		id:           firstPacketID,
		expectedID:   firstPacketID,
		stat:         channelStat{trudp: trudp, timeStarted: now, lastTimeReceived: now},
		sendTestMsgF: false,
		maxQueueSize: trudp.defaultQueueSize,
	}
	tcd.receiveQueue = make([]receiveQueueData, 0)
	tcd.sendQueue = make([]sendQueueData, 0)

	// Add to channels map
	trudp.tcdmap[key] = tcd

	// Channels and sendQueue workers Init
	tcd.sendQueueCommand(nil)

	trudp.Log(CONNECT, "channel with key", key, "connected")
	tcd.trudp.sendEvent(tcd, CONNECTED, []byte(key))

	return
}

// ConnectChannel to remote host by UDP
func (trudp *TRUDP) ConnectChannel(rhost string, rport int, ch int) (tcd *channelData) {
	address := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		panic(err)
	}
	trudp.Log(CONNECT, "connecting to host", rUDPAddr, "at channel", ch)
	tcd, _ = trudp.newChannelData(rUDPAddr, ch)
	return
}

// CloseChannel close trudp channel
func (tcd *channelData) CloseChannel() {
	tcd.destroy(DEBUGv, "destroy this channel: closed by user")
}

// MakeKey return trudp channel key
func (tcd *channelData) MakeKey() string {
	return tcd.key
}

// canWrine return true if writeTo is allowed
func (tcd *channelData) canWrite() bool {
	return len(tcd.sendQueue) < tcd.maxQueueSize && len(tcd.receiveQueue) < tcd.maxQueueSize
}

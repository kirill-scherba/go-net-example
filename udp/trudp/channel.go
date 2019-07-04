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

	// Channels current IDs
	id         uint32 // Last send packet ID
	expectedID uint32 // Expected incoming ID

	// Channels packet queues
	sendQueue    []sendQueueData    // send queue
	receiveQueue []receiveQueueData // received queue

	// Channel channels and waiting groups
	chSendQueue   chan func()           // channel for worker 'trudp process command'
	chStopWorkers [workersLen]chan bool // channels to stop wokers
	wgWorkers     sync.WaitGroup        // workers stop wait group

	// Channel flags
	stoppedF     bool // trudp channel stopped flag
	sendTestMsgF bool // Send test messages

	// Channel statistic
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
	// Clear sendQueue
	tcd.sendQueueReset()
	// Clear receivedQueue
	tcd.receiveQueueReset()
	// Set tcd.id = 0
	tcd.id = firstPacketID
	// Set tcd.expectedID = 1
	tcd.expectedID = firstPacketID
	// Send event "RESET was applied" to user level
	tcd.trudp.sendEvent(tcd, RESET_LOCAL, nil)
}

// destroy close and destroy trudp channel
func (tcd *channelData) destroy(msgLevel int, msg string) {

	// Disable repeatable 'destroy'
	if tcd.stoppedF {
		return
	}
	tcd.stoppedF = true
	tcd.trudp.log(msgLevel, msg)

	go func() {

		// Stopping workers and
		for idx, _ := range tcd.chStopWorkers {
			tcd.chStopWorkers[idx] <- true
		}

		// Wait to stop workers
		tcd.wgWorkers.Wait()

		// Close workers stop channel and chSendQueue channel
		for idx, _ := range tcd.chStopWorkers {
			close(tcd.chStopWorkers[idx])
		}
		close(tcd.chSendQueue)

		// Clear channel queues (the receive queue was cleaned during stop workers)
		tcd.receiveQueueReset()

		// \TODO Clear/Correct TRUDP statistics data

		// Remove trudp channel from channels map
		key := tcd.trudp.makeKey(tcd.addr, tcd.ch)
		delete(tcd.trudp.tcdmap, key)
		tcd.trudp.log(CONNECT, "channel with key", key, "disconnected")
		tcd.trudp.sendEvent(tcd, DISCONNECTED, []byte(key))
	}()
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
func (tcd *channelData) WriteTo(data []byte) error {
	if !tcd.stoppedF {
		tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeTo(tcd)
		return nil
	}
	return errors.New("channel closed")
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
		trudp.log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}

	// Channel data create
	tcd = &channelData{
		trudp:        trudp,
		addr:         addr,
		ch:           ch,
		id:           firstPacketID,
		expectedID:   firstPacketID,
		stat:         channelStat{trudp: trudp, lastTimeReceived: time.Now()},
		sendTestMsgF: false,
	}
	tcd.receiveQueue = make([]receiveQueueData, 0)
	tcd.sendQueue = make([]sendQueueData, 0)

	// Add to channels map
	trudp.tcdmap[key] = tcd

	// Channels and sendQueue workers Init
	tcd.sendQueueCommand(nil)

	trudp.log(CONNECT, "channel with key", key, "connected")
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
	trudp.log(CONNECT, "connecting to host", rUDPAddr, "at channel", ch)
	tcd, _ = trudp.newChannelData(rUDPAddr, ch)
	return
}

// CloseChannel close trudp channel
func (tcd *channelData) CloseChannel() {
	tcd.destroy(DEBUGv, "destroy this channel: closed by user")
}

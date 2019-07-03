package trudp

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type channelData struct {
	trudp *TRUDP // link to trudp

	addr net.Addr // UDP address
	ch   int      // TRUDP channel

	id         uint32 // Last send packet ID
	expectedID uint32 // Expected incoming ID

	sendTestMsg bool

	sendQueue    []sendQueueData    // send queue
	receiveQueue []receiveQueueData // received queue

	chSendQueue chan func()    // channel for worker 'trudp process command'
	stopWorkers [3]chan bool   // channels to stop wokers
	wgWorkers   sync.WaitGroup // workers stop wait group
	stoppedF    bool           // trudp channel stopped flag

	stat channelStat
}

// Workrs index
const (
	wkProcessCommand = iota
	wkResendProcessing
	wkKeepAlive
	//wkStopped
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

		// Stop workers
		tcd.stopWorkers[wkKeepAlive] <- true
		tcd.stopWorkers[wkResendProcessing] <- true
		tcd.stopWorkers[wkProcessCommand] <- true

		// Wait wokers to stopped
		tcd.wgWorkers.Wait()

		// Close workers channels
		for i := 0; i < len(tcd.stopWorkers)-1; i++ {
			close(tcd.stopWorkers[i])
		}
		close(tcd.chSendQueue)

		// Remove trudp channel from channels map
		key := tcd.trudp.makeKey(tcd.addr, tcd.ch)
		delete(tcd.trudp.tcdmap, key)
		tcd.trudp.log(CONNECT, "channel with key", key, "disconnected")
		tcd.trudp.sendEvent(nil, DISCONNECTED, []byte(key))
	}()
}

// getId return new packe id
func (tcd *channelData) getID() (id uint32) {
	id = atomic.AddUint32(&tcd.id, 1) - 1
	return
}

// SendTestMsg set sendTestMsg flag to send test message by interval
func (tcd *channelData) SendTestMsg(sendTestMsg bool) {
	tcd.sendTestMsg = sendTestMsg
}

// TripTime return current triptime (ms)
func (tcd *channelData) TripTime() float32 {
	return tcd.stat.triptime
}

// WriteTo send data to remote host
func (tcd *channelData) WriteTo(data []byte) (err error) {
	if tcd.stoppedF {
		err = errors.New("channel closed")
		return
	}
	tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeTo(tcd)
	return
}

// makeKey return trudp channel key
func (trudp *TRUDP) makeKey(addr net.Addr, ch int) string {
	return addr.String() + ":" + strconv.Itoa(ch)
}

// newChannelData create new TRUDP ChannelData or select existing
func (trudp *TRUDP) newChannelData(addr net.Addr, ch int) (tcd *channelData, key string) {

	key = trudp.makeKey(addr, ch)

	// Channel data select
	tcd, ok := trudp.tcdmap[key]
	if ok {
		trudp.log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}

	// Channel data create
	tcd = &channelData{
		trudp:       trudp,
		addr:        addr,
		ch:          ch,
		id:          firstPacketID,
		expectedID:  firstPacketID,
		stat:        channelStat{trudp: trudp, lastTimeReceived: time.Now()},
		sendTestMsg: false,
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

	// \TODO Just for test: Send hello to remote host
	// for i := 0; i < 3; i++ {
	// 	trudp.packet.dataCreateNew(tcd.getID(), ch, []byte(helloMsg)).writeTo(tcd)
	// }

	return
}

package trudp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// ChannelData is the TRUDP channel data structure
type ChannelData struct {
	trudp *TRUDP // link to trudp

	// Channels remote host address and channel number
	addr *net.UDPAddr // UDP address
	ch   int          // TRUDP channel number
	key  string       // TRUDP channel key (address and channel string representation)

	// Channels current IDs
	id         uint32 // Last send packet ID
	expectedID uint32 // Expected incoming ID

	// Channels packet queues
	sendQueue    *sendQueueType // send queue
	receiveQueue                // received queue
	writeQueue   []*writeType   // write queue
	maxQueueSize int            // maximum queue size

	// Channel flags
	stoppedF     bool // TRUDP channel stopped flag
	sendTestMsgF bool // Send test messages

	// TRUDP channel statistic
	stat channelStat

	connected bool // Channel is connected when it resive data or ack to data
}

// reset exequte reset of this cannel
func (tcd *ChannelData) reset() {
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
	tcd.trudp.sendEvent(tcd, EvResetLocal, nil)
}

// destroy close and destroy trudp channel
func (tcd *ChannelData) destroy(msgLevel int, msg string) (err error) {

	// Disable repeatable 'destroy'
	if tcd.stoppedF {
		err = errors.New("can't destroy: the channel " + tcd.key + " already closed")
		return
	}

	tcd.stoppedF = true
	teolog.Log(msgLevel, MODULE, msg)

	// Clear receive queue
	tcd.sendQueueReset()

	// Clear receive queue
	tcd.receiveQueueReset()

	// Clear write queue
	tcd.trudp.proc.writeQueueReset(tcd)

	// \TODO clear/correct TRUDP statistics data

	// Remove trudp channel from channels map
	delete(tcd.trudp.tcdmap, tcd.key)
	teolog.Log(teolog.CONNECT, MODULE, "channel with key", tcd.key, "disconnected")
	if tcd.connected {
		tcd.trudp.sendEvent(tcd, EvDisconnected, []byte(tcd.key))
	}

	return
}

// incID return current value and increment packet id
func (tcd *ChannelData) incID(id *uint32) (currentID uint32) {
	currentID = *id
	if (*id)++; (*id) == 0 {
		(*id)++
	}
	return
}

// ID return new packe id
func (tcd *ChannelData) ID() (currentID uint32) {
	for {
		currentID = atomic.LoadUint32(&tcd.id)
		// newID := currentID
		// tcd.incID(&new)
		newID := currentID + 1
		if newID == 0 {
			newID++
		}
		if ok := atomic.CompareAndSwapUint32(&tcd.id, currentID, newID); ok {
			break
		}
	}
	return
}

// AllowSendTestMsg set sendTestMsgF flag to send test message by interval
func (tcd *ChannelData) AllowSendTestMsg(sendTestMsgF bool) {
	tcd.sendTestMsgF = sendTestMsgF
}

// TripTime return current triptime (ms)
func (tcd *ChannelData) TripTime() float32 {
	return tcd.stat.triptime
}

// Write send data to remote host
func (tcd *ChannelData) Write(data []byte) (n int, err error) {
	if tcd.stoppedF {
		err = errors.New("can't write to: the channel " + tcd.key + " already closed")
		return
	}
	chanAnswer := make(chan bool)
	tcd.trudp.proc.chanWrite <- &writeType{tcd, data, chanAnswer}
	<-chanAnswer
	n = len(data)
	return
}

// WriteNowait send data to remote host in no wait mode and got result in callback
func (tcd *ChannelData) WriteNowait(data []byte, cb func()) (n int, err error) {
	if tcd.stoppedF {
		err = errors.New("can't write to: the channel " + tcd.key + " already closed")
		return
	}
	go func() {
		chanAnswer := make(chan bool)
		tcd.trudp.proc.chanWrite <- &writeType{tcd, data, chanAnswer}
		<-chanAnswer
		cb()
	}()
	return
}

// WriteUnsafe send data to remote host by UDP
func (tcd *ChannelData) WriteUnsafe(data []byte) (int, error) {
	return tcd.trudp.udp.writeTo(data, tcd.addr)
}

// makeKey return trudp channel key
func (trudp *TRUDP) makeKey(addr net.Addr, ch int) string {
	return addr.String() + ":" + strconv.Itoa(ch)
}

// newChannelData create new TRUDP ChannelData or select existing
func (trudp *TRUDP) newChannelData(addr *net.UDPAddr, ch int, canCreate,
	sendEvConnected bool) (tcd *ChannelData, key string, ok bool) {

	// Send event connected
	sendEventConnected := func() {
		if sendEvConnected {
			tcd.connected = true
			teolog.Log(teolog.CONNECT, MODULE, "channel", key, "connected")
			tcd.trudp.sendEvent(tcd, EvConnected, []byte(key))
		}
	}

	// Channel data select
	key = trudp.makeKey(addr, ch)
	tcd, ok = trudp.tcdmap[key]
	if ok && !tcd.connected {
		sendEventConnected()
	}
	if ok || !ok && !canCreate {
		//teolog.Log(teolog.DEBUGvv, MODULE, "the ChannelData with key", key, "selected")
		return
	}
	ok = true
	now := time.Now()

	// Channel data create
	tcd = &ChannelData{
		trudp:        trudp,
		addr:         addr,
		ch:           ch,
		key:          key,
		id:           firstPacketID,
		expectedID:   firstPacketID,
		stat:         channelStat{trudp: trudp, timeStarted: now, lastTimeReceived: now, triptimeMiddle: maxRTT},
		sendTestMsgF: false,
		maxQueueSize: trudp.defaultQueueSize,
	}
	tcd.sendQueue = sendQueueInit() // list.New()
	tcd.receiveQueue = receiveQueueInit()
	tcd.writeQueue = make([]*writeType, 0)

	// Add to channels map
	trudp.tcdmap[key] = tcd

	sendEventConnected()

	return
}

// ConnectChannel to remote host by UDP
func (trudp *TRUDP) ConnectChannel(rhost string, rport int, ch int) (tcd *ChannelData) {
	address := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := trudp.udp.resolveAddr(network, address)
	if err != nil {
		panic(err)
	}
	teolog.Log(teolog.CONNECT, MODULE, "connecting to host", rUDPAddr, "at channel", ch)
	done := make(chan bool)
	// Create new trudp channel and wait while channel created in kernel level
	go trudp.kernel(func() {
		tcd, _, _ = trudp.newChannelData(rUDPAddr, ch, true, false)
		done <- true
	})
	<-done
	return
}

// Close close trudp channel
func (tcd *ChannelData) Close() (err error) {
	done := make(chan bool)
	go tcd.trudp.kernel(func() {
		tcd.destroy(teolog.DEBUGv,
			fmt.Sprint("destroy channel ", tcd.GetKey(), ": closed by user"),
		)
		done <- true
	})
	<-done
	return
}

// GetCh return trudp channel
func (tcd *ChannelData) GetCh() int {
	return tcd.ch
}

// GetAddr return trudp channel address
func (tcd *ChannelData) GetAddr() *net.UDPAddr {
	return tcd.addr
}

// GetKey return trudp channel key
func (tcd *ChannelData) GetKey() string {
	return tcd.key
}

// GetTriptime return trudp channel triptime
func (tcd *ChannelData) GetTriptime() (float32, float32) {
	return tcd.stat.triptime, tcd.stat.triptimeMiddle
}

// Connected return channel is connected flag
func (tcd *ChannelData) Connected() bool {
	return tcd.connected
}

// canWrine return true if writeTo is allowed
func (tcd *ChannelData) canWrite() bool {
	return tcd.sendQueue.q.Len() < tcd.maxQueueSize /*&& tcd.receiveQueue.Len() < tcd.maxQueueSize*/
}

// keepAlive Send ping if time since tcd.lastTripTimeReceived >= sleepTime
func (tcd *ChannelData) keepAlive() {

	// Send ping after sleep time
	if time.Since(tcd.stat.lastTripTimeReceived) >= sleepTime {
		tcd.trudp.packet.newPing(tcd.ch, []byte(echoMsg)).writeTo(tcd)
		teolog.Log(teolog.DEBUGv, MODULE, "send ping to channel: ", tcd.key)
	}

	// Destroy channel after disconnect time
	if time.Since(tcd.stat.lastTimeReceived) >= disconnectTime {
		tcd.destroy(teolog.DEBUGv,
			fmt.Sprint("destroy channel ", tcd.GetKey(),
				": does not answer long time: ", time.Since(tcd.stat.lastTimeReceived),
			),
		)
	}

	// \TODO send test data - remove it
	if tcd.sendTestMsgF {
		data := []byte(helloMsg + "-" + strconv.Itoa(int(tcd.id)))
		tcd.trudp.packet.newData(tcd.ID(), tcd.ch, data).writeTo(tcd)
	}
}

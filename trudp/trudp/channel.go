package trudp

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
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
	sendQueue    *list.List   // send queue
	receiveQueue *list.List   // received queue
	writeQueue   []*writeType // write queue
	maxQueueSize int          // maximum queue size

	// Channel flags
	stoppedF     bool // TRUDP channel stopped flag
	sendTestMsgF bool // Send test messages

	// TRUDP channel statistic
	stat channelStat
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
	tcd.trudp.sendEvent(tcd, RESET_LOCAL, nil)
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
	tcd.trudp.sendEvent(tcd, DISCONNECTED, []byte(tcd.key))

	return
}

// getId return new packe id
func (tcd *ChannelData) getID() (id uint32) {
	id = atomic.AddUint32(&tcd.id, 1) - 1
	return
}

// SendTestMsg set sendTestMsgF flag to send test message by interval
func (tcd *ChannelData) SendTestMsg(sendTestMsgF bool) {
	tcd.sendTestMsgF = sendTestMsgF
}

// TripTime return current triptime (ms)
func (tcd *ChannelData) TripTime() float32 {
	return tcd.stat.triptime
}

// WriteTo send data to remote host
func (tcd *ChannelData) WriteTo(data []byte) (err error) {
	if tcd.stoppedF {
		err = errors.New("can't write to: the channel " + tcd.key + " already closed")
		return
	}
	chanAnswer := make(chan bool)
	tcd.trudp.proc.chanWrite <- &writeType{tcd, data, chanAnswer}
	<-chanAnswer
	return
}

// makeKey return trudp channel key
func (trudp *TRUDP) makeKey(addr net.Addr, ch int) string {
	return addr.String() + ":" + strconv.Itoa(ch)
}

// newChannelData create new TRUDP ChannelData or select existing
func (trudp *TRUDP) newChannelData(addr *net.UDPAddr, ch int) (tcd *ChannelData, key string) {

	key = trudp.makeKey(addr, ch)

	// Channel data select
	tcd, ok := trudp.tcdmap[key]
	if ok {
		//teolog.Log(teolog.DEBUGvv, MODULE, "the ChannelData with key", key, "selected")
		return
	}

	now := time.Now()

	// Channel data create
	tcd = &ChannelData{
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
	tcd.sendQueue = list.New()
	tcd.receiveQueue = list.New()
	tcd.writeQueue = make([]*writeType, 0)

	// Add to channels map
	trudp.tcdmap[key] = tcd

	teolog.Log(teolog.CONNECT, MODULE, "channel with key", key, "connected")
	tcd.trudp.sendEvent(tcd, CONNECTED, []byte(key))

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
	tcd, _ = trudp.newChannelData(rUDPAddr, ch)
	return
}

// CloseChannel close trudp channel
func (tcd *ChannelData) CloseChannel() {
	tcd.destroy(teolog.DEBUGv,
		fmt.Sprint("destroy channel ", tcd.MakeKey(), ": closed by user"),
	)
}

// MakeKey return trudp channel key
func (tcd *ChannelData) MakeKey() string {
	return tcd.key
}

// canWrine return true if writeTo is allowed
func (tcd *ChannelData) canWrite() bool {
	return tcd.sendQueue.Len() < tcd.maxQueueSize /*&& tcd.receiveQueue.Len() < tcd.maxQueueSize*/
}

// keepAlive Send ping if time since tcd.lastTripTimeReceived >= sleepTime
func (tcd *ChannelData) keepAlive() {

	// Send ping after sleep time
	if time.Since(tcd.stat.lastTripTimeReceived) >= sleepTime {
		tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
		teolog.Log(teolog.DEBUGv, MODULE, "send ping to", tcd.key)
	}

	// Destroy channel after disconnect time
	if time.Since(tcd.stat.lastTimeReceived) >= disconnectTime {
		tcd.destroy(teolog.DEBUGv,
			fmt.Sprint("destroy channel ", tcd.MakeKey(),
				": does not answer long time: ", time.Since(tcd.stat.lastTimeReceived),
			),
		)
	}

	// \TODO send test data - remove it
	if tcd.sendTestMsgF {
		data := []byte(helloMsg + "-" + strconv.Itoa(int(tcd.id)))
		tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeTo(tcd)
	}
}

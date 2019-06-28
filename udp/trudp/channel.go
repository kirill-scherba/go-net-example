package trudp

import (
	"net"
	"strconv"
	"time"
)

type receivedQueueData struct {
	packet []byte
}

type channelData struct {
	trudp *TRUDP // link to trudp

	addr net.Addr // UDP address
	ch   int      // TRUDP channel

	id         uint // Last send packet ID
	expectedID uint // Expected incoming ID

	triptime         float32   // Channels triptime in Millisecond
	triptimeMiddle   float32   // Channels midle triptime in Millisecond
	lastTimeReceived time.Time // Time when last packet was received

	sendQueue     []sendQueueData     // send queue
	receivedQueue []receivedQueueData // received queue

	chSendQueue chan func()
}

const (
	_ANSI_NONE       = "\033[0m"
	_ANSI_RED        = "\033[22;31m"
	_ANSI_LIGHTGREEN = "\033[01;32m"
	_ANSI_LIGHTRED   = "\033[01;31m"
	_ANSI_LIGHTBLUE  = "\033[01;34m"
)

func (tcd *channelData) receivedQueueProcess(packet []byte) {
	id := tcd.trudp.packet.getID(packet)
	switch {
	case id == tcd.expectedID:
		tcd.expectedID++
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTGREEN+"received valid packet id", id, _ANSI_NONE)
	case id == 1:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "need to reset locally"+_ANSI_NONE)
		tcd.reset()
	case tcd.expectedID == 1:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "need to reset remote host"+_ANSI_NONE)
		// \TODO send reset
	case id < tcd.expectedID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
	}
}

// reset this cannel  \TODO
func (tcd *channelData) reset() {
	// Clear sendQueue
	// Clear receivedQueue
	// Set tcd.id = 0
	// Set tcd.expectedID = 1
}

// newChannelData create new TRUDP ChannelData or select existing
func (trudp *TRUDP) newChannelData(addr net.Addr, ch int) (tcd *channelData, key string) {

	key = addr.String() + ":" + strconv.Itoa(ch)

	tcd, ok := trudp.tcdmap[key]
	if ok {
		trudp.log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}

	// Channel data create
	tcd = &channelData{
		trudp:      trudp,
		addr:       addr,
		ch:         ch,
		id:         0,
		expectedID: 1,
	}
	tcd.receivedQueue = make([]receivedQueueData, 0)
	tcd.sendQueue = make([]sendQueueData, 0)
	//tcd.sendQueueProcess(sqpINIT, nil)
	trudp.tcdmap[key] = tcd

	// Keep alive: send Ping
	//tcd.chKeepAlive <- sendQueueProcessCommand{cmd, data}
	go func(conn *net.UDPConn) {
		for {
			time.Sleep(pingInterval * time.Millisecond)
			tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
			tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, []byte(helloMsg)).writeTo(tcd)
		}
	}(trudp.conn)

	trudp.log(DEBUGvv, "new ChannelData with key", key, "created")

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

	// tcd.sendQueueProcess(sqINIT, nil)
	// tcd.sendQueueProcess(sqDESTROY, nil)

	// Send hello to remote host
	for i := 0; i < 3; i++ {
		//trudp.packet.writeTo(tcd, trudp.packet.dataCreateNew(tcd.getID(), ch, []byte(helloMsg))) //, rUDPAddr, true)
		trudp.packet.dataCreateNew(tcd.getID(), ch, []byte(helloMsg)).writeTo(tcd)
	}

	// Keep alive: send Ping
	// go func(conn *net.UDPConn) {
	// 	for {
	// 		time.Sleep(pingInterval * time.Millisecond)
	// 		//trudp.packet.writeTo(tcd, trudp.packet.pingCreateNew(ch, []byte(echoMsg))) //, rUDPAddr, false)
	// 		trudp.packet.pingCreateNew(ch, []byte(echoMsg)).writeTo(tcd)
	// 	}
	// }(trudp.conn)

	return
}

// getId return new packe id
func (tcd *channelData) getID() uint {
	tcd.id++
	return tcd.id
}

// setTriptime save triptime to the ChannelData
func (tcd *channelData) setTriptime(triptime float32) {
	tcd.triptime = triptime
	if tcd.triptimeMiddle == 0 {
		tcd.triptimeMiddle = tcd.triptime
		return
	}
	tcd.triptimeMiddle = (tcd.triptimeMiddle*10 + tcd.triptime) / 11
}

// setLastTimeReceived save last time received from channel to the ChannelData
func (tcd *channelData) setLastTimeReceived() {
	tcd.lastTimeReceived = time.Now()
}

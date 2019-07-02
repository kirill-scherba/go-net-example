package trudp

import (
	"errors"
	"fmt"
)

type receiveQueueData struct {
	packet *packetType
}

const (
	_ANSI_NONE       = "\033[0m"
	_ANSI_RED        = "\033[22;31m"
	_ANSI_LIGHTGREEN = "\033[01;32m"
	_ANSI_LIGHTRED   = "\033[01;31m"
	_ANSI_LIGHTBLUE  = "\033[01;34m"
	_ANSI_YELLOW     = "\033[01;33m"
)

// receivedQueueProcess process received packet, check receivedQueue and
// send received data and events to user level
func (tcd *channelData) receivedQueueProcess(packet []byte) {
	id := tcd.trudp.packet.getID(packet)
	switch {

	// Valid data packet
	case id == tcd.expectedID:
		tcd.expectedID++
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTGREEN+"received valid packet id", id, _ANSI_NONE)
		// Send received data packet to user level
		tcd.trudp.sendEvent(tcd, GOT_DATA, tcd.trudp.packet.getData(packet))
		// Check packets in received queue
		for {
			idx, rqd, err := tcd.receiveQueueFind(tcd.expectedID)
			if err != nil {
				break
			}
			tcd.expectedID++
			tcd.trudp.log(DEBUGv, "find packet in receivedQueue, id:", tcd.trudp.packet.getID(rqd.packet.data))
			// Send received data packet to user level
			tcd.trudp.sendEvent(tcd, GOT_DATA, tcd.trudp.packet.getData(packet))
			tcd.receiveQueueRemove(idx)
		}

	// Invalid packet (with id = 0)
	case id == firstPacketID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "reset locally"+_ANSI_NONE)
		tcd.reset()
		// \TODO Send received data packet to user level

	// Invalid packet (when expectedID = 0)
	case tcd.expectedID == firstPacketID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "send reset remote host"+_ANSI_NONE)
		ch := tcd.trudp.packet.getChannel(packet)
		tcd.trudp.packet.resetCreateNew(ch).writeTo(tcd) // Send reset
		// \TODO Send event "RESET was sent" to user level

	// Already processed packet (id < expectedID)
	case id < tcd.expectedID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
		// \TODO Set statistic REJECTED (already received) packet

	// Packet with id more than expectedID placed to receive queue and wait
	// previouse packets
	case id > tcd.expectedID:
		_, _, err := tcd.receiveQueueFind(id)
		if err == nil {
			tcd.trudp.log(DEBUGv, _ANSI_YELLOW+"move received packet to received queue, id", id, "wait previouse packets"+_ANSI_NONE)
			tcd.receiveQueueAdd(packet)
		} else {
			tcd.trudp.log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
			// \TODO Set statistic REJECTED (already received) packet
		}
	}
}

// receiveQueueAdd add packet to receive queue
func (tcd *channelData) receiveQueueAdd(data []byte) {
	packet := &packetType{trudp: tcd.trudp, data: data}
	tcd.receiveQueue = append(tcd.receiveQueue, receiveQueueData{packet: packet})
	tcd.trudp.log(DEBUGv, "add to send queue, id", packet.getID(packet.data))
}

// receiveQueueFind find packet with selected id in receiveQueue
func (tcd *channelData) receiveQueueFind(id uint) (idx int, rqd receiveQueueData, err error) {
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	for idx, rqd = range tcd.receiveQueue {
		if tcd.trudp.packet.getID(rqd.packet.data) == id {
			err = nil
			break
		}
	}
	return
}

// receiveQueueRemove remove previousely found element from receive queue by index
func (tcd *channelData) receiveQueueRemove(idx int) {
	tcd.receiveQueue = append(tcd.receiveQueue[:idx], tcd.receiveQueue[idx+1:]...)
	tcd.trudp.log(DEBUGv, "remove from receive queue, index", idx)
}

// receiveQueueReset resets (clear) send queue
func (tcd *channelData) receiveQueueReset() {
	// for _, sqd := range tcd.receiveQueue {
	// 	sqd.packet.destroy()
	// }
	tcd.receiveQueue = tcd.receiveQueue[:0]
}

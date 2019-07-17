package trudp

import (
	"errors"
	"fmt"
)

// receiveQueueData receive queue data structure
type receiveQueueData struct {
	packet *packetType
}

// receiveQueueAdd add packet to receive queue
func (tcd *channelData) receiveQueueAdd(packet *packetType) {
	//packet := &packetType{trudp: tcd.trudp, data: data}
	tcd.receiveQueue = append(tcd.receiveQueue, receiveQueueData{packet: packet})
	tcd.trudp.Log(DEBUGv, "add to send queue, id", packet.getID())
}

// receiveQueueFind find packet with selected id in receiveQueue
func (tcd *channelData) receiveQueueFind(id uint32) (idx int, rqd receiveQueueData, err error) {
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	for idx, rqd = range tcd.receiveQueue {
		if rqd.packet.getID() == id {
			err = nil
			break
		}
	}
	return
}

// receiveQueueRemove remove previousely found element from receive queue by index
func (tcd *channelData) receiveQueueRemove(idx int) {
	tcd.receiveQueue = append(tcd.receiveQueue[:idx], tcd.receiveQueue[idx+1:]...)
	tcd.trudp.Log(DEBUGv, "remove from receive queue, index", idx)
}

// receiveQueueReset resets (clear) send queue
func (tcd *channelData) receiveQueueReset() {
	tcd.receiveQueue = tcd.receiveQueue[:0]
}

// receiveQueueProcess find packets in received queue sendEvent and remove packet
func (tcd *channelData) receiveQueueProcess(sendEvent func(data []byte)) {
	for {
		idx, rqd, err := tcd.receiveQueueFind(tcd.expectedID)
		if err != nil {
			break
		}
		tcd.expectedID++
		tcd.trudp.Log(DEBUGv, "find packet in receivedQueue, id:", rqd.packet.getID())
		// Send received data packet to user level
		sendEvent(rqd.packet.getData())
		tcd.receiveQueueRemove(idx)
	}
}

package trudp

import (
	"container/list"
	"errors"
	"fmt"
)

// receiveQueueData receive queue data structure
type receiveQueueData struct {
	packet *packetType
}

// receiveQueueAdd add packet to receive queue
func (tcd *ChannelData) receiveQueueAdd(packet *packetType) {
	tcd.receiveQueue.PushBack(&receiveQueueData{packet: packet})
	tcd.trudp.Log(DEBUGv, "add to send queue, id", packet.getID())
}

// receiveQueueFind find packet with selected id in receiveQueue
func (tcd *ChannelData) receiveQueueFind(id uint32) (e *list.Element, rqd *receiveQueueData, err error) {
	for e = tcd.receiveQueue.Front(); e != nil; e = e.Next() {
		rqd = e.Value.(*receiveQueueData)
		if rqd.packet.getID() == id {
			return
		}
	}
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	return
}

// receiveQueueRemove remove previousely found element from receive queue by index
func (tcd *ChannelData) receiveQueueRemove(e *list.Element) {
	tcd.receiveQueue.Remove(e)
	tcd.trudp.Log(DEBUGv, "remove from receive queue, e", e.Value.(*receiveQueueData).packet.getID())
}

// receiveQueueReset resets (clear) send queue
func (tcd *ChannelData) receiveQueueReset() {
	tcd.receiveQueue.Init()
}

// receiveQueueProcess find packets in received queue sendEvent and remove packet
func (tcd *ChannelData) receiveQueueProcess(sendEvent func(data []byte)) {
	for {
		e, rqd, err := tcd.receiveQueueFind(tcd.expectedID)
		if err != nil {
			break
		}
		tcd.expectedID++
		tcd.trudp.Log(DEBUGv, "find packet in receivedQueue, id:", rqd.packet.getID())
		sendEvent(rqd.packet.getData())
		tcd.receiveQueueRemove(e)
	}
}

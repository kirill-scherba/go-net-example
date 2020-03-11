package trudp

import (
	"errors"
	"strconv"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// receiveQueueType is the receive queue type definition
type receiveQueueType map[uint32]*receiveQueueData

// receiveQueueData receive queue data structure
type receiveQueueData struct {
	packet *packetType
}

// receiveQueueInit create new receive
func receiveQueueInit() receiveQueueType {
	return make(map[uint32]*receiveQueueData)
}

// receiveQueueAdd add packet to receive queue
func (tcd *ChannelData) receiveQueueAdd(packet *packetType) {
	id := packet.ID()
	tcd.receiveQueue[id] = &receiveQueueData{packet: packet}
	teolog.Log(teolog.DEBUGvv, MODULE, "add to receive queue, id:", id)
}

// receiveQueueFind find packet with selected id in receiveQueue
func (tcd *ChannelData) receiveQueueFind(id uint32) (rqd *receiveQueueData, ok bool) {
	rqd, ok = tcd.receiveQueue[id]
	return
}

// receiveQueueRemove remove element from receive queue by id
func (tcd *ChannelData) receiveQueueRemove(id uint32) {
	delete(tcd.receiveQueue, id)
	teolog.Logf(teolog.DEBUGvv, MODULE, "remove id %d from receive queue", id)
}

// receiveQueueReset resets (clear) send queue
func (tcd *ChannelData) receiveQueueReset() {
	tcd.receiveQueue = receiveQueueInit()
}

// receiveQueueProcess find packets in received queue sendEvent and remove packet
func (tcd *ChannelData) receiveQueueProcess(sendEvent func(data []byte)) (err error) {
	for {
		id := tcd.expectedID
		rqd, ok := tcd.receiveQueueFind(id)
		if !ok {
			break
		}
		// \TODO: this a critical place where we have packet in received queue
		// but has not place in event queue and can't read new packet because
		// afraid deadlock
		if !tcd.trudp.sendEventAvailable() {
			teolog.Error(MODULE, "ebzdik-2:"+strconv.Itoa(len(tcd.trudp.chanEvent)))
			err = errors.New("can't process all receive queue")
			break
		}
		tcd.incID(&tcd.expectedID)
		teolog.Log(teolog.DEBUGvv, MODULE, "find packet in receivedQueue, id:", id)
		sendEvent(rqd.packet.Data())
		tcd.receiveQueueRemove(id)
	}
	return
}

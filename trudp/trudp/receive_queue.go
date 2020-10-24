package trudp

import (
	"errors"
	"strconv"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// receiveQueue is the receive queue type definition
type receiveQueue map[uint32]*receiveQueueData

// receiveQueueData receive queue data structure
type receiveQueueData struct {
	packet *packetType
}

// receiveQueueInit create new receive
func receiveQueueInit() receiveQueue {
	return make(map[uint32]*receiveQueueData)
}

// receiveQueueReset resets (clear) send queue
func (tcd *ChannelData) receiveQueueReset() {
	tcd.receiveQueue = receiveQueueInit()
}

// receiveQueueProcess find packets in received queue sendEvent and remove packet
func (tcd *ChannelData) receiveQueueProcess(sendEvent func(data []byte)) (err error) {
	for {
		id := tcd.expectedID
		rqd, ok := tcd.receiveQueue.Find(id)
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
		tcd.receiveQueue.Remove(id)
	}
	return
}

// receiveQueueAdd add packet to receive queue
func (r receiveQueue) Add(packet *packetType) {
	id := packet.ID()
	r[id] = &receiveQueueData{packet: packet}
	teolog.Log(teolog.DEBUGvv, MODULE, "add to receive queue, id:", id)
}

// receiveQueueFind find packet with selected id in receiveQueue
func (r receiveQueue) Find(id uint32) (rqd *receiveQueueData, ok bool) {
	rqd, ok = r[id]
	return
}

// receiveQueueRemove remove element from receive queue by id
func (r receiveQueue) Remove(id uint32) {
	delete(r, id)
	teolog.Logf(teolog.DEBUGvv, MODULE, "remove id %d from receive queue", id)
}

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
	tcd.trudp.log(DEBUGv, "add to send queue, id", packet.getID())
}

// receiveQueueFind find packet with selected id in receiveQueue
func (tcd *channelData) receiveQueueFind(id uint) (idx int, rqd receiveQueueData, err error) {
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
	tcd.trudp.log(DEBUGv, "remove from receive queue, index", idx)
}

// receiveQueueReset resets (clear) send queue
func (tcd *channelData) receiveQueueReset() {
	// for _, sqd := range tcd.receiveQueue {
	// 	sqd.packet.destroy()
	// }
	tcd.receiveQueue = tcd.receiveQueue[:0]
}

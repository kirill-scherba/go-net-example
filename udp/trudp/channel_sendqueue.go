package trudp

import (
	"errors"
	"fmt"
	"time"
)

type sendQueueData struct {
	packet   *packetData
	sendtime time.Time
}

type sendQueueProcessCommand struct {
	commandType int
	data        interface{}
}

// func (tcd *channelData) sendQueueProcess(cmd quecommand) {
func (tcd *channelData) sendQueueProcess(fnc func()) {
	if tcd.chSendQueue == nil {
		tcd.trudp.log(DEBUGv, "sendQueue channel created")
		tcd.chSendQueue = make(chan func())

		// Send queue process command worker
		go func() {
			for {
				if tcd.chSendQueue != nil {
					(<-tcd.chSendQueue)()
				} else {
					break
				}
			}
		}()

		go func() {
			for {
				var t time.Duration = pingInterval
				time.Sleep(t * time.Millisecond)

				tcd.sendQueueProcess(func() {
					//t = tcd.sendQueueProcessQueue()
				})
			}
		}()

	}

	tcd.chSendQueue <- fnc
}

// sendQueueAdd add packet to send queue
func (tcd *channelData) sendQueueAdd(packet *packetData) {
	tcd.sendQueue = append(tcd.sendQueue, sendQueueData{packet: packet, sendtime: time.Now()})
	tcd.trudp.log(DEBUGv, "add to send queue, id", tcd.trudp.packet.getId(packet.data))
}

// sendQueueFind find packet in sendQueue
func (tcd *channelData) sendQueueFind(packet []byte) (idx int, sqd sendQueueData, id uint, err error) {
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	id = tcd.trudp.packet.getId(packet)
	for idx, sqd = range tcd.sendQueue {
		if tcd.trudp.packet.getId(sqd.packet.data) == id {
			err = nil
			break
		}
	}
	return
}

// sendQueueUpdate update record in send queue
func (tcd *channelData) sendQueueUpdate(packet []byte) {
	_, sqd, id, err := tcd.sendQueueFind(packet)
	if err == nil {
		sqd.sendtime = time.Now()
	}
	tcd.trudp.log(DEBUGv, "updated record in send queue, id", id)
}

// sendQueueRemove remove packet from send queue
func (tcd *channelData) sendQueueRemove(packet []byte) {
	idx, sqd, id, err := tcd.sendQueueFind(packet)
	if err == nil {
		sqd.packet.destroy()
		tcd.sendQueue = append(tcd.sendQueue[:idx], tcd.sendQueue[idx+1:]...)
		tcd.trudp.log(DEBUGv, "remove from send queue, id", id)
	}
}

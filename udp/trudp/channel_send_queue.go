package trudp

import (
	"errors"
	"fmt"
	"time"
)

type sendQueueData struct {
	packet      *packetData
	sendTime    time.Time
	arrivalTime time.Time
}

type sendQueueProcessCommand struct {
	commandType int
	data        interface{}
}

// sendQueueProcess receive messageas from channel and exequte it
// in 'Send queue process command' worker
func (tcd *channelData) sendQueueProcess(fnc func()) {

	if tcd.chSendQueue == nil {
		tcd.trudp.log(DEBUGv, "sendQueue channel created")
		tcd.chSendQueue = make(chan func())

		// Send queue 'process command' worker
		go func() {
			for {
				if tcd.chSendQueue == nil {
					break
				}
				(<-tcd.chSendQueue)()
			}
		}()

		// Send queue 'resend processing' worker
		go func() {
			var t time.Duration = defaultRTT * time.Millisecond
			for {
				if tcd.chSendQueue == nil {
					break
				}
				//tcd.trudp.log(DEBUG, "proces, time:", int(t))
				time.Sleep(t)
				tcd.sendQueueProcess(func() { t = tcd.sendQueueResendProcess() })
			}
		}()
	}

	tcd.chSendQueue <- fnc
}

// sendQueueResendProcess resend packet from send queue if it does not got
// ACK during selected time
func (tcd *channelData) sendQueueResendProcess() (rtt time.Duration) {
	rtt = defaultRTT * time.Millisecond
	now := time.Now()
	for _, sqd := range tcd.sendQueue {
		var t time.Duration
		if !now.After(sqd.arrivalTime) {
			t = time.Until(sqd.arrivalTime)
		} else {
			// Resend recort with arrivalTime less than Windows
			t = time.Duration(defaultRTT+tcd.triptimeMiddle) * time.Millisecond
			sqd.sendTime = now
			sqd.arrivalTime = now.Add(t)
			tcd.trudp.conn.WriteTo(sqd.packet.data, tcd.addr)

			tcd.trudp.log(DEBUG, "resend sendQueue packet with",
				"id:", fmt.Sprintf("%4d", sqd.packet.getID(sqd.packet.data)),
				"rtt:", t)
		}
		// Next time to run sendQueueResendProcess
		if t < rtt {
			rtt = t
		}
	}
	return
}

// sendQueueAdd add packet to send queue
func (tcd *channelData) sendQueueAdd(packet *packetData) {
	now := time.Now()
	var rttTime time.Duration = defaultRTT
	tcd.sendQueue = append(tcd.sendQueue, sendQueueData{
		packet:      packet,
		sendTime:    now,
		arrivalTime: now.Add(rttTime * time.Millisecond)})

	tcd.trudp.log(DEBUGv, "add to send queue, id", tcd.trudp.packet.getID(packet.data))
}

// sendQueueFind find packet in sendQueue
func (tcd *channelData) sendQueueFind(packet []byte) (idx int, sqd sendQueueData, id uint, err error) {
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	id = tcd.trudp.packet.getID(packet)
	for idx, sqd = range tcd.sendQueue {
		if tcd.trudp.packet.getID(sqd.packet.data) == id {
			err = nil
			break
		}
	}
	return
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

// sendQueueReset \TODO reset (clear) send queue
func (tcd *channelData) sendQueueReset() {
	for _, sqd := range tcd.sendQueue {
		sqd.packet.destroy()
	}
	tcd.sendQueue = nil
}

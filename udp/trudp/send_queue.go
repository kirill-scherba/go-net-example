package trudp

import (
	"errors"
	"fmt"
	"time"
)

type sendQueueData struct {
	packet        *packetType
	sendTime      time.Time
	arrivalTime   time.Time
	resendAttempt int
}

// sendQueueResendProcess Resend packet from send queue if it does not got
// ACK during selected time. Destroy channel if too much resends happens =
// maxResendAttempt constant
// \TODO check this resend and calculate new resend time algorithm
func (tcd *channelData) sendQueueResendProcess() (rtt time.Duration) {
	rtt = (defaultRTT + time.Duration(tcd.stat.triptimeMiddle)) * time.Millisecond
	now := time.Now()
	for _, sqd := range tcd.sendQueue {
		var t time.Duration
		if !now.After(sqd.arrivalTime) {
			//t = time.Until(sqd.arrivalTime)
			break
		} else {
			// Destroy this trudp channel if resendAttemp more than maxResendAttemp
			if sqd.resendAttempt >= maxResendAttempt {
				// Destroy this trudp channel
				tcd.destroy(DEBUGv, fmt.Sprint("destroy this channel: too much resends happens: ", sqd.resendAttempt))
				break
			}
			// Resend record with arrivalTime less than Windows
			t = time.Duration(defaultRTT+tcd.stat.triptimeMiddle) * time.Millisecond
			sqd.packet.writeTo(tcd)
			// Statistic
			tcd.stat.repeat()

			tcd.trudp.Log(DEBUG, "resend sendQueue packet with",
				"id:", sqd.packet.getID(),
				"attempt:", sqd.resendAttempt,
				"rtt:", t)
		}
	}
	// Next time to run sendQueueResendProcess
	// if len(tcd.sendQueue) > 0 {
	// 	rtt = tcd.sendQueue[0].arrivalTime.Sub(now)
	// }
	return
}

// sendQueueAdd add or update send queue packet
func (tcd *channelData) sendQueueAdd(packet *packetType) {
	now := time.Now()
	var rttTime time.Duration = defaultRTT + time.Duration(tcd.stat.triptimeMiddle)
	arrivalTime := now.Add(rttTime * time.Millisecond)

	idx, _, _, err := tcd.sendQueueFind(packet)
	if err != nil {
		tcd.sendQueue = append(tcd.sendQueue, sendQueueData{
			packet:      packet,
			sendTime:    now,
			arrivalTime: arrivalTime,
		})
		tcd.trudp.Log(DEBUGv, "add to send queue, id", packet.getID())
	} else {
		tcd.sendQueue[idx].arrivalTime = arrivalTime
		tcd.sendQueue[idx].resendAttempt++
		tcd.trudp.Log(DEBUGv, "update in send queue, id", packet.getID())
	}
}

// sendQueueFind find packet in sendQueue
func (tcd *channelData) sendQueueFind(packet *packetType) (idx int, sqd sendQueueData, id uint32, err error) {
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	id = packet.getID()
	for idx, sqd = range tcd.sendQueue {
		if sqd.packet.getID() == id {
			err = nil
			break
		}
	}
	return
}

// sendQueueRemove remove packet from send queue
func (tcd *channelData) sendQueueRemove(packet *packetType) {
	idx, sqd, id, err := tcd.sendQueueFind(packet)
	if err == nil {
		sqd.packet.destroy()
		tcd.sendQueue = append(tcd.sendQueue[:idx], tcd.sendQueue[idx+1:]...)
		tcd.trudp.Log(DEBUGv, "remove from send queue, id", id)
	}
}

// sendQueueReset resets (clear) send queue
func (tcd *channelData) sendQueueReset() {
	for _, sqd := range tcd.sendQueue {
		sqd.packet.destroy()
	}
	tcd.sendQueue = tcd.sendQueue[:0]
}

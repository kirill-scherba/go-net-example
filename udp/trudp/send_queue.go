package trudp

import (
	"container/list"
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
	for e := tcd.sendQueue.Front(); e != nil; e = e.Next() {
		sqd := e.Value.(*sendQueueData)
		// Do while packets ready to resend
		if !now.After(sqd.arrivalTime) {
			break
		}
		// Destroy this trudp channel if resendAttemp more than maxResendAttemp
		if sqd.resendAttempt >= maxResendAttempt {
			tcd.destroy(DEBUGv, fmt.Sprint("destroy this channel: too much resends happens: ", sqd.resendAttempt))
			break
		}
		// Resend packet, save resend to statistic and show message
		sqd.packet.writeTo(tcd)
		tcd.stat.repeat()
		tcd.trudp.Log(DEBUG, "resend sendQueue packet with",
			"id:", sqd.packet.getID(),
			"attempt:", sqd.resendAttempt)
	}
	// Next time to run sendQueueResendProcess
	if tcd.sendQueue.Len() > 0 {
		rtt = tcd.sendQueue.Front().Value.(*sendQueueData).arrivalTime.Sub(now)
	}
	return
}

// sendQueueAdd add or update send queue packet
func (tcd *channelData) sendQueueAdd(packet *packetType) {
	now := time.Now()
	var triptimeMiddle time.Duration
	if tcd.stat.triptimeMiddle > maxRTT {
		triptimeMiddle = maxRTT
	} else {
		triptimeMiddle = time.Duration(tcd.stat.triptimeMiddle)
	}
	var rttTime time.Duration = defaultRTT + triptimeMiddle
	arrivalTime := now.Add(rttTime * time.Millisecond)

	_, sqd, _, err := tcd.sendQueueFind(packet)
	if err != nil {
		tcd.sendQueue.PushBack(&sendQueueData{
			packet:      packet,
			sendTime:    now,
			arrivalTime: arrivalTime,
		})
		tcd.trudp.Log(DEBUGv, "add to send queue, id", packet.getID())
	} else {
		sqd.arrivalTime = arrivalTime
		sqd.resendAttempt++
		tcd.trudp.Log(DEBUGv, "update in send queue, id", packet.getID())
	}
}

// sendQueueFind find packet in sendQueue
func (tcd *channelData) sendQueueFind(packet *packetType) (e *list.Element, sqd *sendQueueData, id uint32, err error) {
	id = packet.getID()
	for e = tcd.sendQueue.Front(); e != nil; e = e.Next() {
		sqd = e.Value.(*sendQueueData)
		if sqd.packet.getID() == id {
			return
		}
	}
	err = errors.New(fmt.Sprint("not found, packet id: ", id))
	return
}

// sendQueueRemove remove packet from send queue
func (tcd *channelData) sendQueueRemove(packet *packetType) {
	e, sqd, id, err := tcd.sendQueueFind(packet)
	if err == nil {
		sqd.packet.destroy()
		tcd.sendQueue.Remove(e)
		tcd.trudp.Log(DEBUGv, "remove from send queue, id", id)
	}
}

// sendQueueReset resets (clear) send queue
func (tcd *channelData) sendQueueReset() {
	for e := tcd.sendQueue.Front(); e != nil; e = e.Next() {
		e.Value.(*sendQueueData).packet.destroy()
	}
	tcd.sendQueue.Init()
}

package trudp

import (
	"container/list"
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// sendQueueType is the send queue type definition
type sendQueueType struct {
	q   *list.List       // send queue list
	idx sendQueueIdxType // send queue id index
}

type sendQueueIdxType map[uint32]*list.Element

type sendQueueData struct {
	packet        *packetType // packet
	sendTime      time.Time   // time when packet was send
	arrivalTime   time.Time   // time when packet need resend
	resendAttempt int         // number of resend was done
}

// receiveQueueInit create new send queue
func sendQueueInit() *sendQueueType {
	return &sendQueueType{list.New(), sendQueueIdxInit()}
}

// sendQueueIdxInit create new send queue id index
func sendQueueIdxInit() sendQueueIdxType {
	return make(map[uint32]*list.Element)
}

// sendQueueRttTime return send queue rtt time
func (tcd *ChannelData) sendQueueRttTime() (triptimeMiddle time.Duration) {
	if tcd.stat.triptimeMiddle > maxRTT {
		triptimeMiddle = maxRTT
	} else {
		triptimeMiddle = time.Duration(tcd.stat.triptimeMiddle)
	}
	triptimeMiddle += defaultRTT
	return
}

// sendQueueAdd add or update send queue packet
func (tcd *ChannelData) sendQueueAdd(packet *packetType) {
	id := packet.ID()
	now := time.Now()
	arrivalTime := now.Add(tcd.sendQueueRttTime() * time.Millisecond)

	_, sqd, ok := tcd.sendQueueFind(id)
	if !ok {
		tcd.sendQueue.idx[id] = tcd.sendQueue.q.PushBack(&sendQueueData{
			packet:      packet,
			sendTime:    now,
			arrivalTime: arrivalTime,
		})
		teolog.Log(teolog.DEBUGvv, MODULE, "add to send queue, id:", id)
	} else {
		sqd.resendAttempt++
		sqd.arrivalTime = arrivalTime
		teolog.Log(teolog.DEBUGvv, MODULE, "update in send queue, id:", id)
	}
}

// sendQueueFind find packet in sendQueue
func (tcd *ChannelData) sendQueueFind(id uint32) (e *list.Element,
	sqd *sendQueueData, ok bool) {
	e, ok = tcd.sendQueue.idx[id]
	if ok {
		sqd = e.Value.(*sendQueueData)
	}
	return
}

// sendQueueRemove remove packet from send queue
func (tcd *ChannelData) sendQueueRemove(id uint32) {
	e, sqd, ok := tcd.sendQueueFind(id)
	if ok {
		sqd.packet.destroy()
		tcd.sendQueue.q.Remove(e)
		delete(tcd.sendQueue.idx, id)
		teolog.Log(teolog.DEBUGvv, MODULE, "remove from send queue, id:", id)
	}
}

// sendQueueCalculateLength calculate send queue length
func (tcd *ChannelData) sendQueueCalculateLength() {
	// Calculate new send queue length if send packets speed more than 30 pac/sec
	if tcd.stat.packets.sendRT.SpeedPacSec > 30 {
		//currentLen := tcd.sendQueue.Len()
		lessMaxSize := tcd.maxQueueSize < 2048 //1024
		queueIsFull := tcd.sendQueue.q.Len() >= tcd.maxQueueSize
		moreDefaultSize := tcd.maxQueueSize > 4 //tcd.trudp.defaultQueueSize
		//  if queue capacity less max capacity size
		if lessMaxSize {
			// if repeat speed is nil (0 repeat packets during second) and
			// queue is full
			if tcd.stat.packets.repeatRT.SpeedPacSec == 0 && queueIsFull {
				tcd.maxQueueSize += 32
			}
		}
		// if queue capacity more default(minimal) capacity size
		if moreDefaultSize {
			// if repeat speed more than 20 packets per second or
			// if repeat speed more than 10 packets per second and queue is full
			if tcd.stat.packets.repeatRT.SpeedPacSec > 20 ||
				tcd.stat.packets.repeatRT.SpeedPacSec > 10 && queueIsFull {
				tcd.maxQueueSize -= 4
			}
		}
	}
}

// sendQueueReset resets (clear) send queue
func (tcd *ChannelData) sendQueueReset() {
	for e := tcd.sendQueue.q.Front(); e != nil; e = e.Next() {
		e.Value.(*sendQueueData).packet.destroy()
	}
	tcd.sendQueue.q.Init()
	tcd.sendQueue.idx = sendQueueIdxInit()
}

// sendQueueResendProcess resend packet from send queue if it does not got
// ACK during selected time. Destroy channel if too much resends happens =
// maxResendAttempt constant
// \TODO check this resend and calculate new resend time algorithm
func (tcd *ChannelData) sendQueueResendProcess() (rtt time.Duration) {
	rtt = (defaultRTT + time.Duration(tcd.stat.triptimeMiddle)) * time.Millisecond
	now := time.Now()
	for e := tcd.sendQueue.q.Front(); e != nil; e = e.Next() {
		sqd := e.Value.(*sendQueueData)
		// Do while packets ready to resend
		if !now.After(sqd.arrivalTime) {
			tcd.stat.repeat(false)
			break
		}
		// Destroy this trudp channel if resendAttemp more than maxResendAttemp
		if sqd.resendAttempt >= maxResendAttempt {
			tcd.destroy(teolog.DEBUGv, fmt.Sprint("destroy channel ",
				tcd.GetKey(), ": too much resends happens: ",
				sqd.resendAttempt))
			break
		}
		// Resend packet, save resend to statistic and show message
		// sqd.packet.updateTimestamp().writeTo(tcd)
		p := sqd.packet
		p.destoryF = false
		p.data = append([]byte(nil), sqd.packet.data...)
		p.updateTimestamp().writeTo(tcd)
		//
		tcd.stat.repeat(true)
		teolog.Log(teolog.DEBUGvv, MODULE, "resend sendQueue packet ",
			"id:", sqd.packet.ID(),
			"attempt:", sqd.resendAttempt)
	}
	// Next time to run sendQueueResendProcess
	if tcd.sendQueue.q.Len() > 0 {
		rtt = tcd.sendQueue.q.Front().Value.(*sendQueueData).arrivalTime.Sub(now)
	}
	return
}

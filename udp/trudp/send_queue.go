package trudp

import (
	"errors"
	"fmt"
	"time"
)

type sendQueueData struct {
	packet        *packetData
	sendTime      time.Time
	arrivalTime   time.Time
	resendAttempt int
}

type sendQueueProcessCommand struct {
	commandType int
	data        interface{}
}

// sendQueueProcess receive messageas from channel and exequte it
// in 'Send queue process command' worker
func (tcd *channelData) sendQueueProcess(fnc func()) {

	// Start trudp channel and sendQueue workers
	if tcd.chSendQueue == nil {
		tcd.trudp.log(DEBUGv, "sendQueue channel created")
		tcd.chSendQueue = make(chan func())
		var stopWorkers = false

		// Send queue 'process command' worker
		go func() {
			for {
				// Wait message from channel
				f, ok := <-tcd.chSendQueue
				// Exit from worker if channel closed
				if !ok {
					tcd.trudp.log(DEBUGv, "sendQueue channel 'process command' worker stopped")
					stopWorkers = true
					break
				}
				// Exequte commands
				switch f {
				case nil:
					continue // Init (nil)
				default:
					f() // All other commands
				}
			}
		}()

		// Send queue 'resend processing' worker
		go func() {
			var t time.Duration = defaultRTT * time.Millisecond
			for {
				//tcd.trudp.log(DEBUG, "proces, time:", int(t))
				time.Sleep(t)
				if stopWorkers {
					tcd.trudp.log(DEBUGv, "sendQueue channel 'resend processing' worker stopped")
					break
				}
				tcd.sendQueueProcess(func() { t = tcd.sendQueueResendProcess() })
			}
		}()

		// Channel 'keep alive (send Ping)' worker
		go func() {
			for {
				time.Sleep(pingInterval * time.Millisecond)
				if stopWorkers {
					tcd.trudp.log(DEBUGv, "sendQueue channel 'keep alive (send Ping)' worker stopped")
					break
				}
				// \TODO Send ping only if tcd.lastTimeReceived < pingInterval
				//if tcd.lastTimeReceived
				//tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
				tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, []byte(helloMsg)).writeTo(tcd)
			}
		}()

	}

	// Send message to sendQueue 'process command' worker
	tcd.chSendQueue <- fnc
}

// sendQueueResendProcess resend packet from send queue if it does not got
// ACK during selected time
func (tcd *channelData) sendQueueResendProcess() (rtt time.Duration) {
	rtt = defaultRTT * time.Millisecond
	now := time.Now()
	for i, sqd := range tcd.sendQueue {
		var t time.Duration
		if !now.After(sqd.arrivalTime) {
			t = time.Until(sqd.arrivalTime)
		} else {
			// Destroy this trudp channel if resendAttemp more than maxResendAttemp
			if sqd.resendAttempt >= maxResendAttempt {
				// \TODO destroy this trudp channel
				tcd.trudp.log(DEBUGv, "destroy this channel: too much resends happens", sqd.resendAttempt)
				tcd.destroy()
				break
			}
			// Resend recort with arrivalTime less than Windows
			t = time.Duration(defaultRTT+tcd.triptimeMiddle) * time.Millisecond
			tcd.sendQueue[i].sendTime = now
			tcd.sendQueue[i].resendAttempt++
			tcd.sendQueue[i].arrivalTime = now.Add(t)
			tcd.trudp.conn.WriteTo(sqd.packet.data, tcd.addr)

			tcd.trudp.log(DEBUG, "resend sendQueue packet with",
				"id:", fmt.Sprintf("%4d", sqd.packet.getID(sqd.packet.data)),
				"attempt:", sqd.resendAttempt,
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

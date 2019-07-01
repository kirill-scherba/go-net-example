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

// sendQueueProcess receive messageas from channel and exequte it
// in 'Send queue process command' worker
func (tcd *channelData) sendQueueProcess(fnc func()) {

	// Start trudp channel and sendQueue workers
	if tcd.chSendQueue == nil {

		tcd.trudp.log(DEBUGv, "sendQueue channel created")

		// Initialize channels
		tcd.chSendQueue = make(chan func())
		for idx, _ := range tcd.stopWorkers {
			tcd.stopWorkers[idx] = make(chan bool)
		}

		// Send queue 'process command' worker. Exequte all concurent sendQueue
		// commands.
		go func() {
			tcd.trudp.log(DEBUGv, "worker 'trudp process command' started")
		for_l:
			for {
				// Wait message from chSendQueue or stopWorkers channels
				select {
				case <-tcd.stopWorkers[wkProcessCommand]:
					break for_l
				case fun := <-tcd.chSendQueue:
					// Exequte commands(functions) but skip nil, nil sends on Init
					switch fun {
					case nil:
					default:
						fun()
					}
				}
			}
			tcd.trudp.log(DEBUGv, "worker 'trudp process command' stopped")
			tcd.stopWorkers[wkStopped] <- true
		}()

		// Send queue 'resend processing' worker
		go func() {
			var t time.Duration = defaultRTT * time.Millisecond
			tcd.trudp.log(DEBUGv, "worker 'send queue resend processing' started")
		for_l:
			for {
				select {
				case <-tcd.stopWorkers[wkResendProcessing]:
					break for_l
				default:
					time.Sleep(t)
					tcd.sendQueueProcess(func() { t = tcd.sendQueueResendProcess() })
				}
			}
			tcd.sendQueueReset()
			tcd.trudp.log(DEBUGv, "worker 'send queue resend processing' stopped")
			tcd.stopWorkers[wkStopped] <- true
		}()

		// Channel 'keep alive (send ping)' worker. Sleep during pingInterval
		// constant and send ping if nothing received in sleep period. Destroy
		// channel if peer does not answer long time = disconnectAfterTime constant
		go func() {
			slepTime := pingInterval * time.Millisecond
			disconnectAfterTime := disconnectAfter * time.Millisecond
			tcd.trudp.log(DEBUGv, "worker 'trudp keep alive (send ping)' started")
		for_l:
			for {
				select {
				case <-tcd.stopWorkers[wkKeepAlive]:
					break for_l
				default:
					time.Sleep(slepTime)
					// Send ping if time since tcd.lastTimeReceived >= pingInterval
					switch {
					case time.Since(tcd.lastTimeReceived) >= disconnectAfterTime:
						tcd.trudp.log(DEBUGv, "destroy this channel: does not answer long time", time.Since(tcd.lastTimeReceived))
						tcd.destroy()
						break
					case time.Since(tcd.lastTimeReceived) >= slepTime:
						tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
						tcd.trudp.log(DEBUGv, "send ping to", tcd.trudp.makeKey(tcd.addr, tcd.ch))
					}
					// \TODO send test data - remove it
					if tcd.sendTestMsg {
						tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, []byte(helloMsg)).writeTo(tcd)
					}
				}
			}
			tcd.trudp.log(DEBUGv, "worker 'trudp keep alive (send ping)' stopped")
			tcd.stopWorkers[wkStopped] <- true
		}()

	}

	// Send message to sendQueue 'process command' worker
	tcd.chSendQueue <- fnc
}

// sendQueueResendProcess resend packet from send queue if it does not got
// ACK during selected time. Destroy channel if too much resends happens =
// maxResendAttempt constant
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
				// Destroy this trudp channel
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
	//tcd.sendQueue = nil
	tcd.sendQueue = tcd.sendQueue[:0]
}

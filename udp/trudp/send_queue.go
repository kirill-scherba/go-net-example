package trudp

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type sendQueueData struct {
	packet        *packetType
	sendTime      time.Time
	arrivalTime   time.Time
	resendAttempt int
}

// sendQueueCommand receive messageas from channel and exequte it
// in 'Send queue process command' worker
func (tcd *channelData) sendQueueCommand(fnc func()) {

	// Start trudp channel and sendQueue workers
	if tcd.chSendQueue == nil {

		tcd.trudp.log(DEBUGv, "sendQueue channel created")

		// Initialize channels
		tcd.chSendQueue = make(chan func())
		for idx, _ := range tcd.stopWorkers {
			tcd.stopWorkers[idx] = make(chan bool)
		}

		// Initialize workers stop wait group
		tcd.wgWorkers.Add(len(tcd.stopWorkers))

		// Workers star stop messages
		startMsg := func(name string) { tcd.trudp.log(DEBUGv, "worker "+name+" started") }
		stopMsg := func(name string) { tcd.trudp.log(DEBUGv, "worker "+name+" stopped"); tcd.wgWorkers.Done() }

		// Send queue 'process command' worker. Exequte all concurent sendQueue
		// commands.
		go func() {
			worker := "'trudp process command'"
			startMsg(worker)
			defer func() { stopMsg(worker) }()
			for {
				// Wait message from chSendQueue or stopWorkers channels
				select {
				case <-tcd.stopWorkers[wkProcessCommand]:
					return
				case fun := <-tcd.chSendQueue:
					// Exequte commands(functions) but skip nil, nil sends on Init
					if fun != nil {
						fun()
					}
				}
			}
		}()

		// Send queue 'resend processing' worker
		go func() {
			worker := "'send queue resend processing'"
			startMsg(worker)
			var t time.Duration = defaultRTT * time.Millisecond
			defer func() { tcd.sendQueueReset(); stopMsg(worker) }()
			for {
				select {
				case <-tcd.stopWorkers[wkResendProcessing]:
					return
				default:
					time.Sleep(t)
					tcd.sendQueueCommand(func() { t = tcd.sendQueueResendProcess() })
				}
			}
		}()

		// Channel 'keep alive (send ping)' worker. Sleep during pingInterval
		// constant and send ping if nothing received in sleep period. Destroy
		// channel if peer does not answer long time = disconnectAfterTime constant
		go func() {
			worker := "'trudp keep alive (send ping)'"
			slepTime := pingInterval * time.Millisecond
			disconnectAfterTime := disconnectAfter * time.Millisecond
			startMsg(worker)
			defer func() { stopMsg(worker) }()
			for {
				select {
				case <-tcd.stopWorkers[wkKeepAlive]:
					return
				default:
					time.Sleep(slepTime)
					// Send ping if time since tcd.lastTripTimeReceived >= pingInterval
					switch {
					case time.Since(tcd.stat.lastTimeReceived) >= disconnectAfterTime:
						tcd.destroy(DEBUGv, fmt.Sprint("destroy this channel: does not answer long time, since: ", time.Since(tcd.stat.lastTimeReceived)))
					case time.Since(tcd.stat.lastTripTimeReceived) >= slepTime:
						tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
						tcd.trudp.log(DEBUGv, "send ping to", tcd.trudp.makeKey(tcd.addr, tcd.ch))
					}
					// \TODO send test data - remove it
					if tcd.sendTestMsg {
						data := []byte(helloMsg + "-" + strconv.Itoa(int(tcd.id)))
						tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeTo(tcd)
						tcd.trudp.sendEvent(tcd, SEND_DATA, data)
					}
				}
			}
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
				//tcd.trudp.log(DEBUGv, "destroy this channel: too much resends happens", sqd.resendAttempt)
				tcd.destroy(DEBUGv, fmt.Sprint("destroy this channel: too much resends happens: ", sqd.resendAttempt))
				break
			}
			// Resend recort with arrivalTime less than Windows
			t = time.Duration(defaultRTT+tcd.stat.triptimeMiddle) * time.Millisecond
			tcd.sendQueue[i].sendTime = now
			tcd.sendQueue[i].resendAttempt++
			tcd.sendQueue[i].arrivalTime = now.Add(t)
			tcd.trudp.conn.WriteTo(sqd.packet.data, tcd.addr)
			tcd.trudp.sendEvent(tcd, SEND_DATA, sqd.packet.getData())

			tcd.trudp.log(DEBUG, "resend sendQueue packet with",
				"id:", fmt.Sprintf("%4d", sqd.packet.getID()),
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
func (tcd *channelData) sendQueueAdd(packet *packetType) {
	now := time.Now()
	var rttTime time.Duration = defaultRTT
	tcd.sendQueue = append(tcd.sendQueue, sendQueueData{
		packet:      packet,
		sendTime:    now,
		arrivalTime: now.Add(rttTime * time.Millisecond)})

	tcd.trudp.log(DEBUGv, "add to send queue, id", packet.getID())
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
		tcd.trudp.log(DEBUGv, "remove from send queue, id", id)
	}
}

// sendQueueReset resets (clear) send queue
func (tcd *channelData) sendQueueReset() {
	for _, sqd := range tcd.sendQueue {
		sqd.packet.destroy()
	}
	tcd.sendQueue = tcd.sendQueue[:0]
}

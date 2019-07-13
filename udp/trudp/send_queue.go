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
func (tcd *channelData) sendQueueCommand(fnc func()) (err error) {

	// Return error if trudp channel already closed
	if tcd.stoppedF {
		err = errors.New("can't process command: the channel " + tcd.key + " already closed")
		return
	}

	// Start trudp channel and sendQueue workers
	if tcd.chProcessCommand == nil {

		tcd.trudp.Log(DEBUGv, "sendQueue channel created")

		// Initialize channels
		tcd.chWrite = make(chan []byte /* *packetType*/, chWriteSize)
		tcd.chProcessCommand = make(chan func())
		for idx := range tcd.chStopWorkers {
			tcd.chStopWorkers[idx] = make(chan bool)
		}

		// Initialize workers stop wait group
		tcd.wgWorkers.Add(len(tcd.chStopWorkers))

		// Workers star stop functions
		start := func(name string) { tcd.trudp.Log(DEBUGv, "worker "+name+" started") }
		stop := func(name string) { tcd.trudp.Log(DEBUGv, "worker "+name+" stopped"); tcd.wgWorkers.Done() }

		// Send queue 'process command' worker. Exequte all concurent sendQueue
		// commands.
		go func() {
			worker := "'process command'"
			start(worker)
			resendTime := defaultRTT * time.Millisecond
			sleepTime := pingInterval * time.Millisecond
			disconnectTime := disconnectAfter * time.Millisecond
			timerResend := time.After(resendTime)
			timerKeep := time.NewTicker(sleepTime)

			defer func() { timerKeep.Stop(); tcd.sendQueueReset(); stop(worker) }()
			for {
				// Wait message from chSendQueue or stopWorkers channels
				select {

				// Process Stop worker channel
				case <-tcd.chStopWorkers[wkProcessCommand]:
					return

				// task 1: Execute coomands(functions) received from chSendQueue channel
				case fun := <-tcd.chProcessCommand:
					if fun != nil {
						fun()
					}

				// task 2: Process sendQueue (resend packets from sendQueue)
				case <-timerResend:
					// resendTime = tcd.sendQueueResendProcess()
					// timerResend = time.After(resendTime)

				// task 3: Keepalive: Send ping if time since tcd.lastTripTimeReceived >= pingInterval
				case <-timerKeep.C:
					switch {
					case time.Since(tcd.stat.lastTimeReceived) >= disconnectTime:
						tcd.destroy(DEBUGv, fmt.Sprint("destroy this channel: does not answer long time: ", time.Since(tcd.stat.lastTimeReceived)))
					case time.Since(tcd.stat.lastTripTimeReceived) >= sleepTime:
						tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
						tcd.trudp.Log(DEBUGv, "send ping to", tcd.key)
					}
					// \TODO send test data - remove it
					if tcd.sendTestMsgF {
						data := []byte(helloMsg + "-" + strconv.Itoa(int(tcd.id)))
						tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeToUnsafe(tcd)
					}

				// task 4: Got packet from chWrite (from user level) and write it to teonet channel
				case data := <-tcd.checkChWrite():
					tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeToUnsafe(tcd)
				}
			}
		}()
	}

	// Send message to sendQueue 'process command' worker
	tcd.chProcessCommand <- fnc
	return
}

// checkChWrite got chWrite or nil channel depend of sendQueue length
func (tcd *channelData) checkChWrite() chan []byte {
	if len(tcd.sendQueue) < tcd.maxQueueSize && len(tcd.receiveQueue) < tcd.maxQueueSize {
		return tcd.chWrite
	}
	return nil
}

// resetChWrite reset user write channel
func (tcd *channelData) resetChWrite() {
	for len(tcd.chWrite) > 0 {
		select {
		case _, ok := <-tcd.chWrite:
			if ok {
				//packet.destroy()
			}
		}
	}
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
			sqd.packet.writeToUnsafe(tcd)
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

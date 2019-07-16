package trudp

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// This module process all trudp internal events:
// - read (received from udp),
// - write (received from user level, need write to udp)
// - keep alive timer
// - resend packet from send queue timer
// - show statistic timer

// process data structure
type process struct {
	trudp       *TRUDP           // link to trudp
	chanWrite   chan writeType   // channel to write (used to send data from user level)
	chanRead    chan readType    // channel to read (used to process packets received from udp)
	timerKeep   *time.Ticker     // keep alive timer
	timerResend <-chan time.Time // resend packet from send queue timer
}

// read channel data structure
type readType struct {
	addr   *net.UDPAddr
	packet *packetType
}

// read channel data structure
type writeType struct {
	tcd        *channelData
	data       []byte
	chanAnswer chan bool
}

// init
func (proc *process) init(trudp *TRUDP) *process {

	proc.trudp = trudp

	// Set time variables
	disconnectTime := disconnectAfter * time.Millisecond
	sleepTime := pingInterval * time.Millisecond
	resendTime := defaultRTT * time.Millisecond
	pingTime := pingInterval * time.Millisecond

	// Init channels and timers
	proc.chanRead = make(chan readType)
	proc.chanWrite = make(chan writeType)
	//
	proc.timerResend = time.After(resendTime)
	proc.timerKeep = time.NewTicker(pingTime)

	// Module worker
	go func() {

		// Do it on return
		defer func() { proc.timerKeep.Stop() }()

		for {
			select {

			// Process read packet (received from udp)
			case readPac := <-proc.chanRead:
				readPac.packet.process(readPac.addr)

			// Process write packet (received from user level, need write to udp)
			case writePac := <-proc.chanWrite:
				tcd := writePac.tcd
				if tcd.canWrite() {
					proc.writeTo(writePac)
				} else {
					proc.writeQueueAdd(tcd, writePac)
				}

			// Keepalive: Send ping if time since tcd.lastTripTimeReceived >= pingInterval
			case <-proc.timerKeep.C:
				for _, tcd := range proc.trudp.tcdmap {
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
				}

			// Process send queue (resend packets from send queue)
			case <-proc.timerResend:
				resendTime = defaultRTT * time.Millisecond
				for _, tcd := range proc.trudp.tcdmap {
					rt := tcd.sendQueueResendProcess()
					if rt < resendTime {
						resendTime = rt
					}
				}
				proc.timerResend = time.After(resendTime) // Set new timer value
			}
		}
	}()

	return proc
}

// wrieTo write packet to trudp channel and send true to Answer channel
func (proc *process) writeTo(writePac writeType) {
	tcd := writePac.tcd
	proc.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, writePac.data).writeToUnsafe(tcd)
	writePac.chanAnswer <- true
}

// writeQueueAdd add write packet to write queue
func (proc *process) writeQueueAdd(tcd *channelData, writePac writeType) {
	tcd.writeQueue = append(tcd.writeQueue, writePac)
}

// writeQueueWriteTo get packet from writeQueue and send it to trudp channel
func (proc *process) writeQueueWriteTo(tcd *channelData) {
	for len(tcd.writeQueue) > 0 && tcd.canWrite() {
		writePac := tcd.writeQueue[0]
		proc.writeTo(writePac)
		tcd.writeQueue = tcd.writeQueue[1:]
	}
}

func (proc *process) writeQueueReset(tcd *channelData) {
	for _, writePac := range tcd.writeQueue {
		writePac.chanAnswer <- false
	}
	tcd.writeQueue = tcd.writeQueue[:0]
}

// destroy
func (proc *process) destroy() {

}

package trudp

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
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
	trudp          *TRUDP           // link to trudp
	chanRead       chan *readType   // channel to read (used to process packets received from udp)
	chanWrite      chan *writeType  // channel to write (used to send data from user level)
	chanWriter     chan *writerType // channel to write (used to write data to udp)
	timerKeep      *time.Ticker     // keep alive timer
	timerResend    <-chan time.Time // resend packet from send queue timer
	timerStatistic *time.Ticker     // statistic show ticker

	stopRunningF bool           // Stop running flag
	once         sync.Once      // Once to sync trudp event channel stop
	wg           sync.WaitGroup // Wait group
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

type writerType struct {
	packet *packetType
	addr   *net.UDPAddr
}

// init
func (proc *process) init(trudp *TRUDP) *process {

	proc.trudp = trudp

	// Set time variables
	disconnectTime := disconnectAfter * time.Millisecond
	sleepTime := pingInterval * time.Millisecond
	resendTime := defaultRTT * time.Millisecond
	pingTime := pingInterval * time.Millisecond
	statTime := statInterval * time.Millisecond

	// Init channels and timers
	proc.chanWriter = make(chan *writerType, chReadSize)
	proc.chanWrite = make(chan *writeType, chWriteSize)
	proc.chanRead = make(chan *readType, chReadSize)
	//
	proc.timerStatistic = time.NewTicker(statTime)
	proc.timerResend = time.After(resendTime)
	proc.timerKeep = time.NewTicker(pingTime)

	// Module worker
	go func() {

		trudp.Log(CONNECT, "process worker started")
		proc.wg.Add(1)

		// Do it on return
		defer func() {
			proc.timerKeep.Stop()
			proc.timerStatistic.Stop()
			close(proc.chanWriter)
			trudp.Log(CONNECT, "process worker stopped")
			proc.wg.Done()
		}()

		for !proc.stopRunningF {
			select {

			// Process read packet (received from udp)
			case readPac, ok := <-proc.chanRead:
				if ok {
					readPac.packet.process(readPac.addr)
				}

				// Process write packet (received from user level, need write to udp)
			case writePac, ok := <-proc.chanWrite:
				if ok {
					tcd := writePac.tcd
					if tcd.canWrite() {
						proc.writeTo(writePac)
					} else {
						proc.writeQueueAdd(tcd, writePac)
					}
				}

			// Keepalive: Send ping if time since tcd.lastTripTimeReceived >= pingInterval
			case <-proc.timerKeep.C:
				for _, tcd := range proc.trudp.tcdmap {
					switch {
					case time.Since(tcd.stat.lastTimeReceived) >= disconnectTime:
						tcd.destroy(DEBUGv,
							fmt.Sprint("destroy this channel: does not answer long time: ",
								time.Since(tcd.stat.lastTimeReceived)))
					case time.Since(tcd.stat.lastTripTimeReceived) >= sleepTime:
						tcd.trudp.packet.pingCreateNew(tcd.ch, []byte(echoMsg)).writeTo(tcd)
						tcd.trudp.Log(DEBUGv, "send ping to", tcd.key)
					}
					// \TODO send test data - remove it
					if tcd.sendTestMsgF {
						data := []byte(helloMsg + "-" + strconv.Itoa(int(tcd.id)))
						tcd.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, data).writeTo(tcd)
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

			// Process statistic show
			case <-proc.timerStatistic.C:
				proc.showStatistic()
			}
		}
	}()

	// Write worker
	go func() {
		trudp.Log(CONNECT, "writer worker started")
		proc.wg.Add(1)
		defer func() { trudp.Log(CONNECT, "writer worker stopped"); proc.wg.Done() }()
		for w := range proc.chanWriter {
			proc.trudp.udp.writeTo(w.packet.data, w.addr)
			if !w.packet.sendQueueF {
				w.packet.destroy()
			}
		}
	}()

	return proc
}

// wrieTo write packet to trudp channel and send true to Answer channel
func (proc *process) writeTo(writePac *writeType) {
	if proc.trudp.proc.stopRunningF {
		return
	}
	tcd := writePac.tcd
	writePac.chanAnswer <- true
	proc.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, writePac.data).writeTo(tcd)
}

// writeQueueAdd add write packet to write queue
func (proc *process) writeQueueAdd(tcd *channelData, writePac *writeType) {
	tcd.writeQueue = append(tcd.writeQueue, writePac)
}

// writeQueueWriteTo get packet from writeQueue and send it to trudp channel
func (proc *process) writeQueueWriteTo(tcd *channelData) {
	for len(tcd.writeQueue) > 0 && tcd.canWrite() {
		writePac := tcd.writeQueue[0]
		tcd.writeQueue = tcd.writeQueue[1:]
		proc.writeTo(writePac)
	}
}

func (proc *process) writeQueueReset(tcd *channelData) {
	for _, writePac := range tcd.writeQueue {
		writePac.chanAnswer <- false
	}
	tcd.writeQueue = tcd.writeQueue[:0]
}

func (proc *process) showStatistic() {
	trudp := proc.trudp
	if !trudp.showStatF {
		return
	}
	idx := 0
	t := time.Now()
	var str string

	// Read trudp channels map keys to slice and sort it
	keys := make([]string, len(trudp.tcdmap))
	for key := range trudp.tcdmap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Get trudp channels statistic string by sorted keys
	for _, key := range keys {
		tcd, ok := trudp.tcdmap[key]
		if ok {
			str += tcd.stat.statBody(tcd, idx, 0)
			idx++
		}
	}

	// Get fotter and print statistic string
	tcs := &channelStat{trudp: trudp} // Empty Methods holder
	str = tcs.statHeader(time.Since(trudp.startTime), time.Since(t)) + str + tcs.statFooter(idx)
	fmt.Print(str)
}

// destroy
func (proc *process) destroy() {
	proc.stopRunningF = true
}

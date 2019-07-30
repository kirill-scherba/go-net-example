package trudp

import (
	"fmt"
	"net"
	"sort"
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
	trudp       *TRUDP           // link to trudp
	chanRead    chan *readType   // channel to read (used to process packets received from udp)
	chanWrite   chan *writeType  // channel to write (used to send data from user level)
	chanWriter  chan *writerType // channel to write (used to write data to udp)
	timerResend <-chan time.Time // resend packet from send queue timer

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

const disconnectTime = disconnectAfter * time.Millisecond
const sleepTime = pingInterval * time.Millisecond

// init
func (proc *process) init(trudp *TRUDP) *process {

	proc.trudp = trudp

	// Set time variables
	resendTime := defaultRTT * time.Millisecond

	// Init channels and timers
	proc.chanWriter = make(chan *writerType, chReadSize)
	proc.chanWrite = make(chan *writeType, chWriteSize)
	proc.chanRead = make(chan *readType, chReadSize)
	//
	proc.timerResend = time.After(resendTime)

	// Module worker
	go func() {

		trudp.Log(CONNECT, "process worker started")
		proc.wg.Add(1)

		// Do it on return
		defer func() {
			close(proc.chanWriter)
			trudp.Log(CONNECT, "process worker stopped")

			// Close trudp channels, send DESTROY event and close event channel
			trudp.closeChannels()
			trudp.sendEvent(nil, DESTROY, []byte(trudp.udp.localAddr()))
			close(trudp.chanEvent)

			proc.wg.Done()
		}()

		chanWriteClosedF := false
		i := 0
		for {
			select {

			// Process read packet (received from udp)
			case readPac, ok := <-proc.chanRead:
				if !ok {
					if !chanWriteClosedF {
						chanWriteClosedF = true
						close(trudp.proc.chanWrite)
					}
					break
				}
				readPac.packet.process(readPac.addr)

			// Process write packet (received from user level, need write to udp)
			case writePac, ok := <-proc.chanWrite:
				if !ok {
					return
				}
				proc.writeTo(writePac)

			// Process send queue (resend packets from send queue), check Keep alive
			// and show statistic (check after 30 ms)
			case <-proc.timerResend:
				// Loop trudp channels map and check Resend send queue and/or send keep alive signal (ping)
				for _, tcd := range proc.trudp.tcdmap {
					if i%100 == 0 {
						tcd.keepAlive()
					}
					tcd.sendQueueResendProcess()
				}
				// Show statistic window
				if i%3 == 0 {
					proc.showStatistic()
				}
				proc.timerResend = time.After(resendTime) // Set new timer value
				i++
			}
		}
	}()

	// Write worker
	go func() {
		proc.wg.Add(1)
		trudp.Log(CONNECT, "writer worker started")
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

// writeTo  write packet to trudp channel or write packet to write queue
func (proc *process) writeTo(writePac *writeType) {
	tcd := writePac.tcd
	if tcd.canWrite() {
		proc.writeToDirect(writePac)
	} else {
		proc.writeToQueue(tcd, writePac)
	}
}

// writeToDirect write packet to trudp channel and send true to Answer channel
func (proc *process) writeToDirect(writePac *writeType) {
	tcd := writePac.tcd
	writePac.chanAnswer <- true
	proc.trudp.packet.dataCreateNew(tcd.getID(), tcd.ch, writePac.data).writeTo(tcd)
}

// writeToQueue add write packet to write queue
func (proc *process) writeToQueue(tcd *channelData, writePac *writeType) {
	tcd.writeQueue = append(tcd.writeQueue, writePac)
}

// writeQueueWriteTo get packet from writeQueue and send it to trudp channel
func (proc *process) writeQueueWriteTo(tcd *channelData) {
	for len(tcd.writeQueue) > 0 && tcd.canWrite() {
		writePac := tcd.writeQueue[0]
		tcd.writeQueue = tcd.writeQueue[1:]
		proc.writeToDirect(writePac)
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

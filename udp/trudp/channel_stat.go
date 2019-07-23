package trudp

import (
	"fmt"
	"time"
)

// channelStat structure contain channel statistic variables
type channelStat struct {
	trudp                *TRUDP      // Pointer to TRUDP structure
	packets              packetsStat // Packets statistic
	timeStarted          time.Time   // Time when channel created
	triptime             float32     // Channels triptime in Millisecond
	triptimeMiddle       float32     // Channels midle triptime in Millisecond
	lastTimeReceived     time.Time   // Time when last packet was received
	lastTripTimeReceived time.Time   // Time when last packet with triptime was received
}

// realTimeSpeed type to calculate real time speed
type realTimeSpeed struct {
	secArr      [10][2]int // Secondes array
	lastIDX     int        // Last secundes array index
	speedPacSec int        // Speed in pacets/second
	speedMbSec  float32    // Speed in mb/second
}

// calculate function calculate real time packets speed in pac/sec and mb/sec
func (realTime *realTimeSpeed) calculate(length int) {
	now := time.Now()
	nsec := now.UnixNano()
	millis := nsec / 1000000
	currentIdx := int((millis / 100) % 10)
	if realTime.lastIDX != currentIdx {
		realTime.lastIDX = currentIdx
		realTime.secArr[currentIdx][0] = 0
		realTime.secArr[currentIdx][1] = 0
	}
	realTime.secArr[currentIdx][0]++
	realTime.secArr[currentIdx][1] += length
	realTime.speedPacSec, realTime.speedMbSec = func() (speedPacSec int, speedMbSec float32) {
		for _, v := range realTime.secArr {
			speedPacSec += v[0]
			speedMbSec += float32(v[1])
		}
		speedMbSec = speedMbSec / float32(1024*1024)
		return
	}()
}

// setTriptime save triptime to the ChannelData
func (tcs *channelStat) setTriptime(triptime float32) {
	tcs.triptime = triptime
	if tcs.triptimeMiddle == 0 {
		tcs.triptimeMiddle = tcs.triptime
		return
	}
	tcs.triptimeMiddle = (tcs.triptimeMiddle*10 + tcs.triptime) / 11
	tcs.lastTripTimeReceived = time.Now()
}

// setLastTimeReceived save last time received from channel to the ChannelData
func (tcs *channelStat) setLastTimeReceived() {
	tcs.lastTimeReceived = time.Now()
}

// received adds data packets received to statistic
func (tcs *channelStat) received(length int) {
	tcs.trudp.packets.receive++                 // Total data packets received
	tcs.packets.receive++                       // Channel data packets received
	tcs.packets.receiveLength += uint64(length) // Length of packet
	tcs.packets.receiveRT.calculate(length)     // Calculate received real time speed
}

// ackReceived adds ack packets received to statistic
func (tcs *channelStat) ackReceived() {
	tcs.trudp.packets.ack++ // Total ack ackets received
	tcs.packets.ack++       // Channel ack ackets received
}

// dropped adds 'packet received and dropped' to statistic
func (tcs *channelStat) dropped() {
	tcs.trudp.packets.dropped++ // Total received and dropped
	tcs.packets.dropped++       // Channel received and dropped
}

// send adds data packets send to statistic
func (tcs *channelStat) send(length int) {
	tcs.trudp.packets.send++                 // Total packets send
	tcs.packets.send++                       // Channel packets send
	tcs.packets.sendLength += uint64(length) // Length of packet
	tcs.packets.sendRT.calculate(length)     // Calculate send real time speed
}

// repeat adds data packets repeat to statistic
func (tcs *channelStat) repeat() {
	tcs.trudp.packets.repeat++ // Total packets repeat
	tcs.packets.repeat++       // Channel packets repeat
}

// statHeader return statistic header string
func (tcs *channelStat) statHeader(runningTime, executionTime time.Duration) string {
	addr := tcs.trudp.udp.localAddr()
	return fmt.Sprintf(
		/*_ANSI_CLS+*/
		"\0337"+ // Save cursor
			"\033[0;0H"+ // Set cursor to the top
			"TR-UDP statistics, addr: %s, running time: %20v, show statistic time: %v                       \n"+
			"List of channels:                                                                                                                                                                           \n"+
			"--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"  # Key                          Send   Pac/sec   Total(mb) Trip time /  Wait(ms) |  Recv   Pac/sec   Total(mb)     ACK |     Repeat         Drop |   SQ     WQ     RQ    UrQ    UwQ     EQ \n"+
			"--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n",
		addr,
		runningTime,
		executionTime)
}

// statFooter return statistic futer string
func (tcs *channelStat) statFooter(length int) string {
	return fmt.Sprintf(
		"--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"                                                                                                                                                                                            \n"+
			"\033[%d;r"+ // Setscroll mode
			"\0338", // Restore cursor
		8+length)
}

// statBody return one channel statistic body string
func (tcs *channelStat) statBody(tcd *channelData, idx, page int) (retstr string) {

	//timeSinceStart := float64(time.Since(tcs.timeStarted).Seconds())
	// Return repeat packets in %
	repeatP := func() (retval uint32) {
		retval = 0
		if tcs.packets.send > 0 {
			retval = 100 * tcs.packets.repeat / tcs.packets.send
		}
		return
	}
	// Return dropped packets in %
	droppedP := func() (retval uint32) {
		retval = 0
		if tcs.packets.receive > 0 {
			retval = 100 * tcs.packets.dropped / tcs.packets.receive
		}
		return
	}

	retstr = fmt.Sprintf(
		"%3d "+_ANSI_BROWN+"%-24.*s"+_ANSI_NONE+" %8d  %8d %10.3f  %9.3f /%9.3f %8d  %8d %10.3f %8d %8d(%d%%) %8d(%d%%) %6d %6d %6d %6d %6d %6d \n"+
			"",

		idx+1,                 // trudp channel number (in statistic screen)
		len(tcd.key), tcd.key, // key len and key
		tcs.packets.send,                               // packets send
		tcs.packets.sendRT.speedPacSec,                 // float64(tcs.packets.send)/timeSinceStart, // send speed in packets/sec
		float64(tcs.packets.sendLength)/(1024*1024),    // send total in mb
		tcs.triptime,                                   // trip time
		tcs.triptimeMiddle,                             // trip time middle
		tcs.packets.receive,                            // packets receive
		tcs.packets.receiveRT.speedPacSec,              // float64(tcs.packets.receive)/timeSinceStart,    // receive speed in packets/sec
		float64(tcs.packets.receiveLength)/(1024*1024), // receive total in mb
		tcs.packets.ack,                                // packets ack receive
		tcs.packets.repeat,                             // packets repeat
		repeatP(),                                      // packets repeat in %
		tcs.packets.dropped,                            // packets drop
		droppedP(),                                     // packets drop in %
		tcd.sendQueue.Len(),                            // sendQueueSize,
		len(tcd.writeQueue),                            // writeQueueSize,
		tcd.receiveQueue.Len(),                         // receiveQueueSize
		len(tcd.trudp.proc.chanRead),                   // channel read udp Size
		len(tcd.trudp.proc.chanWriter),                 // channel write udp Size
		len(tcd.trudp.chanEvent),                       // eventsQueueSize
	)
	return
}

package trudp

import (
	"fmt"
	"strings"
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

var line = "\033[2K" + strings.Repeat("-", 188) + "\n"

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
	if length > 0 {
		realTime.secArr[currentIdx][0]++
	}
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
	tcs.packets.receive++                             // Channel data packets received
	tcs.trudp.packets.receive++                       // Total data packets received
	tcs.packets.receiveLength += uint64(length)       // Length of packet
	tcs.trudp.packets.receiveLength += uint64(length) // Total length of packet
	tcs.packets.receiveRT.calculate(length)           // Calculate received real time speed
	tcs.trudp.packets.receiveRT.calculate(length)     // Calculate total received real time speed
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
	tcs.packets.send++                             // Channel packets send
	tcs.trudp.packets.send++                       // Total packets send
	tcs.packets.sendLength += uint64(length)       // Length of packet
	tcs.trudp.packets.sendLength += uint64(length) // Total length of packet
	tcs.packets.sendRT.calculate(length)           // Calculate send real time speed
	tcs.trudp.packets.sendRT.calculate(length)     // Calculate total send real time speed
}

// repeat adds data packets repeat to statistic
func (tcs *channelStat) repeat(r bool) {
	if r {
		tcs.trudp.packets.repeat++        // Total packets repeat
		tcs.packets.repeat++              // Channel packets repeat
		tcs.packets.repeatRT.calculate(1) // Calculate repeat speed
	} else {
		tcs.packets.repeatRT.calculate(0) // Calculate repeat speed
	}
}

// statHeader return statistic header string
func (tcs *channelStat) statHeader(runningTime, executionTime time.Duration) string {
	addr := tcs.trudp.udp.localAddr()
	return fmt.Sprintf(
		/*_ANSI_CLS+*/
		"\0337"+ // Save cursor
			"\033[0;0H"+ // Set cursor to the top
			"\033[?7l"+ // Does not wrap
			"\033[2K"+ // Clear line
			"TR-UDP statistics, addr: %s, running time: %20v, show statistic time: %v\n"+
			"\033[2K"+
			"List of channels:\n"+
			line+
			"\033[2K"+
			"  # Key                          Send   Pac/sec   Total(mb) Ping(ms) / Wait(ms) |"+
			"   Recv   Pac/sec   Total(mb)     ACK |     Repeat         Drop |"+
			" SQ l/max   WQ   RQ    UrQ    UwQ     EQ \n"+
			line,
		addr,
		runningTime,
		executionTime)
}

// statFooter return statistic futer string
func (tcs *channelStat) statFooter(length int) (str string) {
	lenadd := 9
	str = "\033[2K" // Clear line
	if length > 0 {
		if length > 1 {
			str += fmt.Sprintf(""+
				"%3d %-24.*s %8d  %8d %10.3f        -  /       -  %8d  %8d %10.3f %8d %8d(%d%%) %8d(%d%%)      -      -      - ",
				length, // Number of channels
				0, "",  // Empty
				tcs.trudp.packets.send,                               // Total send packet
				tcs.trudp.packets.sendRT.speedPacSec,                 // Total send packet/sec
				float64(tcs.trudp.packets.sendLength)/(1024*1024),    // Total send in mb
				tcs.trudp.packets.receive,                            // Total received packet
				tcs.trudp.packets.receiveRT.speedPacSec,              // Total received packet/sec
				float64(tcs.trudp.packets.receiveLength)/(1024*1024), // Total received in mb
				tcs.trudp.packets.ack,                                // packets ack received
				tcs.trudp.packets.repeat,                             // packets repeat
				repeatP(&tcs.trudp.packets),                          // packets repeat in %
				tcs.trudp.packets.dropped,                            // packets dropped
				droppedP(&tcs.packets),                               // packets dropped in %
			)
		} else {
			str += strings.Repeat(" ", 166)
		}

		str += fmt.Sprintf("%6d %6d %6d ",
			len(tcs.trudp.proc.chanRead),   // channel read udp Size
			len(tcs.trudp.proc.chanWriter), // channel write udp Size
			len(tcs.trudp.chanEvent),       // eventsQueueSize
		)
	}
	str = line + str + fmt.Sprintf(
		"\n"+
			"\033[2K\n"+ // Clear line
			"\033[%d;r"+ // Setscroll mode
			"\0338", // Restore cursor
		length+lenadd,
	)
	return
}

// repeatP Return repeat packets in %
func repeatP(packets *packetsStat) (retval uint32) {
	retval = 0
	if packets.send > 0 {
		retval = 100 * packets.repeat / packets.send
	}
	return
}

// droppedP Return dropped packets in %
func droppedP(packets *packetsStat) (retval uint32) {
	retval = 0
	if packets.receive > 0 {
		retval = 100 * packets.dropped / packets.receive
	}
	return
}

// statBody return one channel statistic body string
func (tcs *channelStat) statBody(tcd *ChannelData, idx, page int) (retstr string) {

	retstr = fmt.Sprintf("\033[2K"+
		"%3d "+_ANSI_BROWN+"%-24.*s"+_ANSI_NONE+" %8d  %8d %10.3f%9.3f  /%8.3f  %8d  %8d %10.3f %8d %13s %8d(%d%%) %9s %4d %4d      -      -      - \n",

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
		tcs.packets.ack,                                // packets ack received
		fmt.Sprintf("%d/%d(%d%%)",
			tcs.packets.repeat,               // packets repeat
			tcs.packets.repeatRT.speedPacSec, // packets repeat per sec
			repeatP(&tcs.packets)),           // packets repeat in %
		tcs.packets.dropped,    // packets dropped
		droppedP(&tcs.packets), // packets dropped in %

		fmt.Sprintf("%d/%d",
			tcd.sendQueue.Len(), // sendQueueSize,
			tcd.maxQueueSize),   // size of send queue
		len(tcd.writeQueue),    // writeQueueSize,
		tcd.receiveQueue.Len(), // receiveQueueSize
	)
	return
}

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
}

func (tcs *channelStat) repeat() {
	tcs.trudp.packets.repeat++ // Total packets repeat
	tcs.packets.repeat++       // Channel packets repeat
}

func (tcs *channelStat) statHeader(runningTime, executionTime time.Duration) string {
	return fmt.Sprintf(
		/*_ANSI_CLS+*/
		"\0337"+ // Save cursor
			"\033[0;0H"+ // Set cursor to the top
			"TR-UDP statistics, port 8030, running time: %v, show statistic time: %v                       \n"+
			"List of channels:                                                                                                                                                                 \n"+
			"----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"  # Key                          Send   Speed(p/s)  Total(mb) Trip time /  Wait(ms) |  Recv   Speed(p/s)  Total(mb)     ACK |     Repeat         Drop |   SQ     WQ     RQ     EQ \n"+
			"----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n",
		runningTime,
		executionTime)
}

func (tcs *channelStat) statFooter(length int) string {
	return fmt.Sprintf(
		"----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"                                                                                                                                                                                  \n"+
			"\033[%d;r"+ // Setscroll mode
			"\0338", // Restore cursor
		8+length)
}

func (tcs *channelStat) statBody(tcd *channelData, idx, page int) (retstr string) {
	timeSinceStart := float64(time.Since(tcs.timeStarted).Seconds())
	// Repeat in %
	repeatP := func() (retval uint32) {
		retval = 0
		if tcs.packets.send > 0 {
			retval = 100 * tcs.packets.repeat / tcs.packets.send
		}
		return
	}
	// Dropped in %
	droppedP := func() (retval uint32) {
		retval = 0
		if tcs.packets.receive > 0 {
			retval = 100 * tcs.packets.dropped / tcs.packets.receive
		}
		return
	}

	retstr = fmt.Sprintf(
		"%3d "+_ANSI_BROWN+"%-24.*s"+_ANSI_NONE+" %8d %11.3f %10.3f  %9.3f /%9.3f %8d %11.3f %10.3f %8d %8d(%d%%) %8d(%d%%) %6d %6d %6d %6d \n"+
			"",

		idx+1,                 // trudp channel number (in statistic screen)
		len(tcd.key), tcd.key, // key len and key
		tcs.packets.send, // packets send
		float64(tcs.packets.send)/timeSinceStart,    // send speed in packets/sec
		float64(tcs.packets.sendLength)/(1024*1024), // send total in mb
		tcs.triptime,        // trip time
		tcs.triptimeMiddle,  // trip time middle
		tcs.packets.receive, // packets receive
		float64(tcs.packets.receive)/timeSinceStart,    //  receive speed in packets/sec
		float64(tcs.packets.receiveLength)/(1024*1024), // receive total in mb
		tcs.packets.ack,       // packets ack receive
		tcs.packets.repeat,    // packets repeat
		repeatP(),             // packets repeat in %
		tcs.packets.dropped,   // packets drop
		droppedP(),            // packets drop in %
		len(tcd.sendQueue),    // sendQueueSize,
		len(tcd.chWrite),      // writeQueueSize,
		len(tcd.receiveQueue), // receiveQueueSize
		len(tcd.trudp.chRead), // eventsQueueSize
	)
	return
}

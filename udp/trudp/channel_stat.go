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
func (tcs *channelStat) received() {
	tcs.trudp.packets.receive++ // Total data packets received
	tcs.packets.receive++       // Channel data packets received
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
func (tcs *channelStat) send() {
	tcs.trudp.packets.send++ // Total packets send
	tcs.packets.send++       // Channel packets send
}

func (tcs *channelStat) sprintln(tcd *channelData, idx, page int) (retstr string) {
	timeSinceStart := float64(time.Since(tcs.timeStarted) * time.Second)
	retstr = fmt.Sprintf(
		//"%s%3d "+_ANSI_BROWN+"%-24.*s"+_ANSI_NONE+" %8d %11.3f %10.3f  %9.3f /%9.3f %8d %11.3f %10.3f %8d %8d(%d%%) %8d(%d%%) %6d %6d %6d\n",
		/*_ANSI_CLS+*/

		"\0337"+ // Save cursor
			"\033[0;0H"+
			"  List of channels:                                                                                                                                                        \n"+
			"---------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"  # Key                          Send  Speed(p/s)   Total(mb) Trip time /  Wait(ms) |  Recv   Speed(p/s)  Total(mb)     ACK |     Repeat         Drop |   SQ     WQ     RQ \n"+
			"---------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n"+
			"%3d "+_ANSI_BROWN+"%-24.*s"+_ANSI_NONE+" %8d %11.3f %10.3f  %9.3f /%9.3f %8d %11.3f %10.3f %8d %8d(%d%%) %8d(%d%%) %6d %6d %6d\n"+
			"                                                                                                                                                                           \n"+
			"\033[7;r"+
			"\0338",

		idx+1,
		len(tcd.key), tcd.key, // key_len, key,
		tcs.packets.send, //tcd->stat.packets_send,
		//(double)(1.0 * tcd->stat.send_speed / 1024.0),
		float64(tcs.packets.send)/timeSinceStart, //(double)tcd->stat.packets_send / ((tsf - tcd->stat.started) / 1000000.0),
		float64(tcs.packets.send),                // tcd->stat.send_total,
		tcs.triptime,                             //  tcd->stat.triptime_last / 1000.0,
		tcs.triptimeMiddle,                       // tcd->stat.wait,
		tcs.packets.receive,                      // tcd->stat.packets_receive,
		//(double)(1.0 * tcd->stat.receive_speed / 1024.0),
		float64(tcs.packets.receive)/timeSinceStart, //  (double)tcd->stat.packets_receive / ((tsf - tcd->stat.started) / 1000000.0),
		float64(tcs.packets.receive),                //	tcd->stat.receive_total,
		tcs.packets.ack,                             // tcd->stat.ack_receive,
		0,                                           // tcd->stat.packets_attempt,
		0,                                           // tcd->stat.packets_send ? 100 * tcd->stat.packets_attempt / tcd->stat.packets_send : 0,
		tcs.packets.dropped,                         //tcd->stat.packets_receive_dropped,
		0,                                           // tcd->stat.packets_receive ? 100 * tcd->stat.packets_receive_dropped / tcd->stat.packets_receive : 0,
		len(tcd.sendQueue),                          // sendQueueSize,
		len(tcd.chWrite),                            // writeQueueSize,
		len(tcd.receiveQueue),                       //receiveQueueSize)
	)
	return
}

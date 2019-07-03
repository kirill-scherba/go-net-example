package trudp

import "time"

// channelStat structure contain channel statistic variables
type channelStat struct {
	trudp                *TRUDP      // Pointer to TRUDP structure
	packets              packetsStat // Packets statistic
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

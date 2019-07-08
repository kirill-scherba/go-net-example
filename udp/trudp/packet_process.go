package trudp

import (
	"fmt"
	"net"
)

// Packet type
const (
	DATA     = iota //(0x0)
	ACK             //(0x1)
	RESET           //(0x2)
	ACKReset        //(0x3)
	PING            //(0x4)
	ACKPing         //(0x5)
)

// process received packet
func (pac *packetType) process(addr *net.UDPAddr) (processed bool) {
	processed = false

	ch := pac.getChannel()
	tcd, key := pac.trudp.newChannelData(addr, ch)
	tcd.stat.setLastTimeReceived()

	packetType := pac.getType()
	switch packetType {

	// DATA packet received
	case DATA:
		// Create ACK packet and send it back to sender
		pac.ackCreateNew().writeTo(tcd)
		tcd.stat.received(len(pac.data))
		// Show Log
		pac.trudp.Log(DEBUGv, "DATA      packet received, key:", key,
			"id:", pac.getID(),
			"expected id:", tcd.expectedID,
			"data length:", len(pac.data),
			"data:", pac.getData())
		// Process received queue
		pac.packetDataProcess(tcd)

	// ACK-to-data packet received
	case ACK:
		// Set trip time to ChannelData
		tcd.stat.setTriptime(pac.getTriptime())
		tcd.stat.ackReceived()
		// Show Log
		pac.trudp.Log(DEBUGv, "ACK       packet received, key:", key,
			"id:", pac.getID(),
			"trip time:", fmt.Sprintf("%.3f", tcd.stat.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.stat.triptimeMiddle), "ms")
		// Remove packet from send queue
		go tcd.sendQueueCommand(func() { tcd.sendQueueRemove(pac) })

	// RESET packet received
	case RESET:
		pac.trudp.Log(DEBUGv, "RESET     packet received, key:", key)
		pac.ackToResetCreateNew().writeTo(tcd)
		tcd.sendQueueCommand(func() { tcd.reset() })

	// ACK-to-reset packet received
	case ACKReset:
		pac.trudp.Log(DEBUGv, "ACK_RESET packet received, key:", key)
		tcd.sendQueueCommand(func() { tcd.reset() })

	// PING packet received
	case PING:
		// Create ACK to ping packet and send it back to sender
		pac.ackToPingCreateNew().writeTo(tcd)
		// Show Log
		pac.trudp.Log(DEBUGv, "PING      packet received, key:", key,
			"id:", pac.getID(),
			"expected id:", tcd.expectedID,
			"data:", pac.getData(), string(pac.getData()))

	// ACK-to-PING packet received
	case ACKPing:
		// Set trip time to ChannelData
		tcd.stat.setTriptime(pac.getTriptime())
		pac.trudp.Log(DEBUGv, "ACK_PING  packet received, key:", key,
			"id:", pac.getID(),
			"trip time:", fmt.Sprintf("%.3f", tcd.stat.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.stat.triptimeMiddle), "ms")

	// UNKNOWN packet received
	default:
		pac.trudp.Log(DEBUGv, "UNKNOWN   packet received, key:", key, ", type:", packetType)
	}

	return
}

const (
	_ANSI_NONE = "\033[0m"
	_ANSI_CLS  = "\033[2J"

	_ANSI_RED        = "\033[22;31m"
	_ANSI_LIGHTGREEN = "\033[01;32m"
	_ANSI_LIGHTRED   = "\033[01;31m"
	_ANSI_LIGHTBLUE  = "\033[01;34m"
	_ANSI_YELLOW     = "\033[01;33m"
	_ANSI_BROWN      = "\033[22;33m"
)

// packetDataProcess process received data packet, check receivedQueue and
// send received data and events to user level
func (pac *packetType) packetDataProcess(tcd *channelData) {
	id := pac.getID()
	switch {

	// Valid data packet
	case id == tcd.expectedID:
		tcd.expectedID++
		tcd.trudp.Log(DEBUGv, _ANSI_LIGHTGREEN+"received valid packet id", id, _ANSI_NONE)
		// Send received packet data to user level
		tcd.trudp.sendEvent(tcd, GOT_DATA, pac.getData())
		// Check packets in received queue and send it data to user level
		tcd.receiveQueueProcess(func(data []byte) { tcd.trudp.sendEvent(tcd, GOT_DATA, data) })

	// Invalid packet (with id = 0)
	case id == firstPacketID:
		tcd.trudp.Log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "reset locally"+_ANSI_NONE)
		wait := make(chan bool)
		tcd.sendQueueCommand(func() { tcd.reset(); pac.packetDataProcess(tcd); wait <- true })
		<-wait

	// Invalid packet (when expectedID = 0)
	case tcd.expectedID == firstPacketID:
		tcd.trudp.Log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "send reset remote host"+_ANSI_NONE)
		pac.resetCreateNew().writeTo(tcd) // Send reset
		// Send event "RESET was sent" to user level
		tcd.trudp.sendEvent(tcd, SEND_RESET, nil)

	// Already processed packet (id < expectedID)
	case id < tcd.expectedID:
		tcd.trudp.Log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
		// Set statistic REJECTED (already received) packet
		tcd.stat.dropped()

	// Packet with id more than expectedID placed to receive queue and wait
	// previouse packets
	case id > tcd.expectedID:
		_, _, err := tcd.receiveQueueFind(id)
		if err != nil {
			tcd.trudp.Log(DEBUGv, _ANSI_YELLOW+"move received packet to received queue, id", id, "wait previouse packets"+_ANSI_NONE)
			tcd.receiveQueueAdd(pac)
		} else {
			tcd.trudp.Log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already in receive queue"+_ANSI_NONE)
			// Set statistic REJECTED (already received) packet
			tcd.stat.dropped()
		}
	}
}

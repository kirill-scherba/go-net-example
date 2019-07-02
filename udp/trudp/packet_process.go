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
func (pac *packetType) process(addr net.Addr) (processed bool) {
	processed = false

	ch := pac.getChannel()
	tcd, key := pac.trudp.newChannelData(addr, ch)
	tcd.setLastTimeReceived()

	packetType := pac.getType()
	switch packetType {

	case DATA: // DATA packet received
		// Create ACK packet and send it back to sender
		pac.ackCreateNew().writeTo(tcd)
		// Show log
		pac.trudp.log(DEBUGv, "DATA      packet received, key:", key,
			"id:", fmt.Sprintf("%4d", pac.getID()),
			"expected id:", tcd.expectedID,
			"data length:", len(pac.data),
			"data:", pac.getData())
		// Process received queue
		pac.packetDataProcess(tcd)

	case ACK: // ACK-to-data packet received
		// Set trip time to ChannelData
		tcd.setTriptime(pac.getTriptime())
		// Show log
		pac.trudp.log(DEBUGv, "ACK       packet received, key:", key,
			"id:", fmt.Sprintf("%4d", pac.getID()),
			"trip time:", fmt.Sprintf("%.3f", tcd.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.triptimeMiddle), "ms")
		// Remove packet from send queue
		tcd.sendQueueProcess(func() { tcd.sendQueueRemove(pac) })

	case RESET: // RESET packet received
		pac.trudp.log(DEBUGv, "RESET     packet received, key:", key)
		pac.ackToResetCreateNew().writeTo(tcd)
		tcd.reset()

	case ACKReset: // ACK-to-reset packet received
		pac.trudp.log(DEBUGv, "ACK_RESET packet received, key:", key)
		tcd.reset()

	case PING: // PING packet received
		// Create ACK to ping packet and send it back to sender
		pac.ackToPingCreateNew().writeTo(tcd)
		// Show log
		pac.trudp.log(DEBUGv, "PING      packet received, key:", key,
			"id:", fmt.Sprintf("%4d", pac.getID()),
			"expected id:", tcd.expectedID,
			"data:", pac.getData(), string(pac.getData()))

	case ACKPing: // ACK-to-PING packet received
		// Set trip time to ChannelData
		tcd.setTriptime(pac.getTriptime())
		pac.trudp.log(DEBUGv, "ACK_PING  packet received, key:", key,
			"id:", fmt.Sprintf("%4d", pac.getID()),
			"trip time:", fmt.Sprintf("%.3f", tcd.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.triptimeMiddle), "ms")

	default: // UNKNOWN packet received
		pac.trudp.log(DEBUGv, "UNKNOWN   packet received, key:", key, ", type:", packetType)
	}

	return
}

const (
	_ANSI_NONE       = "\033[0m"
	_ANSI_RED        = "\033[22;31m"
	_ANSI_LIGHTGREEN = "\033[01;32m"
	_ANSI_LIGHTRED   = "\033[01;31m"
	_ANSI_LIGHTBLUE  = "\033[01;34m"
	_ANSI_YELLOW     = "\033[01;33m"
)

// packetDataProcess process received data packet, check receivedQueue and
// send received data and events to user level
func (packet *packetType) packetDataProcess(tcd *channelData) {
	id := packet.getID()
	switch {

	// Valid data packet
	case id == tcd.expectedID:
		tcd.expectedID++
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTGREEN+"received valid packet id", id, _ANSI_NONE)
		// Send received data packet to user level
		tcd.trudp.sendEvent(tcd, GOT_DATA, packet.getData())
		// Check packets in received queue
		for {
			idx, rqd, err := tcd.receiveQueueFind(tcd.expectedID)
			if err != nil {
				break
			}
			tcd.expectedID++
			tcd.trudp.log(DEBUGv, "find packet in receivedQueue, id:", rqd.packet.getID())
			// Send received data packet to user level
			tcd.trudp.sendEvent(tcd, GOT_DATA, rqd.packet.getData())
			tcd.receiveQueueRemove(idx)
		}

	// Invalid packet (with id = 0)
	case id == firstPacketID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "reset locally"+_ANSI_NONE)
		tcd.reset()
		// \TODO Send received data packet to user level

	// Invalid packet (when expectedID = 0)
	case tcd.expectedID == firstPacketID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTRED+"received invalid packet id", id, "send reset remote host"+_ANSI_NONE)
		packet.resetCreateNew().writeTo(tcd) // Send reset
		// \TODO Send event "RESET was sent" to user level

	// Already processed packet (id < expectedID)
	case id < tcd.expectedID:
		tcd.trudp.log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
		// \TODO Set statistic REJECTED (already received) packet

	// Packet with id more than expectedID placed to receive queue and wait
	// previouse packets
	case id > tcd.expectedID:
		_, _, err := tcd.receiveQueueFind(id)
		if err == nil {
			tcd.trudp.log(DEBUGv, _ANSI_YELLOW+"move received packet to received queue, id", id, "wait previouse packets"+_ANSI_NONE)
			tcd.receiveQueueAdd(packet)
		} else {
			tcd.trudp.log(DEBUGv, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
			// \TODO Set statistic REJECTED (already received) packet
		}
	}
}

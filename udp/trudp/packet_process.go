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
		tcd.receivedQueueProcess(pac)

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

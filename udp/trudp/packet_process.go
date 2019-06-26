package trudp

import "net"

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
func (pac *packetType) process(packet []byte, addr net.Addr) (processed bool) {
	processed = false

	ch := pac.getChannel(packet)
	tcd, key := pac.trudp.newChannelData(addr, ch)
	tcd.setLastTimeReceived()

	packetType := pac.getType(packet)
	switch packetType {

	case DATA: // DATA packet received
		pac.trudp.log(DEBUGv, "DATA packet received, key:", key,
			"id:", pac.getId(packet), "expected id:", tcd.expectedID,
			"data: ", pac.getData(packet))
    // Create ACK packet and send it back to sender
		pac.writeTo(tcd, pac.ackCreateNew(packet), addr, false)
    tcd.receivedQueueProcess(packet) // Process received queue

	case ACK: // ACK-to-data packet received
    // Set trip time to ChannelData
		tcd.setTriptime(pac.getTriptime(packet))
		pac.trudp.log(DEBUGv, "ACK packet received, key:", key,
			"trip time:", tcd.triptime, "ms")
		tcd.sendQueueRemove(packet) // Remove packet from send queue

	case RESET: // RESET packet received
		pac.trudp.log(DEBUGv, "RESET packet received, key:", key)

	case ACKReset: // ACK-to-reset packet received
		pac.trudp.log(DEBUGv, "ACK_RESET packet received, key:", key)

	case PING: // PING packet received
		pac.trudp.log(DEBUGv, "PING packet received, key:", key,
			"data:", pac.getData(packet), string(pac.getData(packet)))
		// Create ACK to ping packet and send it back to sender
		pac.writeTo(tcd, pac.ackToPingCreateNew(packet), addr, false)

	case ACKPing: // ACK-to-PING packet received
		// Set trip time to ChannelData
		tcd.setTriptime(pac.getTriptime(packet))
		pac.trudp.log(DEBUGv, "ACK_PING packet received, key:", key,
			"trip time:", tcd.triptime, "ms")

	default: // UNKNOWN packet received
		pac.trudp.log(DEBUGv, "UNKNOWN packet received, key:", key, ", type:", packetType)
	}

	return
}

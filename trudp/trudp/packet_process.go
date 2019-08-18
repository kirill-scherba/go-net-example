package trudp

import (
	"fmt"
	"net"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
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
	tcd, key, ok := pac.trudp.newChannelData(addr, ch, pac.getType() == DATA)
	if !ok {
		return
	}

	tcd.stat.setLastTimeReceived()

	packetType := pac.getType()
	switch packetType {

	// DATA packet received
	case DATA:

		// \TODO: drop this packet if EQ len >= MaxValue
		// if len(pac.trudp.chanEvent) > 16 {
		// 	break
		// }

		// Show Log
		teolog.DebugVf(MODULE, "got DATA packet id: %d, channel: %s, "+
			"expected id: %d, data_len: %d",
			pac.getID(), key, tcd.expectedID, len(pac.data),
		)

		// Create ACK packet and send it back to sender
		pac.ackCreateNew().writeTo(tcd)
		tcd.stat.received(len(pac.data))

		// Process received queue
		pac.packetDataProcess(tcd)

	// ACK-to-data packet received
	case ACK:

		// Show Log
		teolog.DebugVf(MODULE, "got ACK packet id: %d, channel: %s, "+
			"triptime: %.3f ms\n",
			pac.getID(), key, tcd.stat.triptime,
		)

		// Set trip time to ChannelData
		tcd.stat.setTriptime(pac.getTriptime())
		tcd.stat.ackReceived()

		// Remove packet from send queue
		tcd.sendQueueRemove(pac)
		tcd.trudp.proc.writeQueueWriteTo(tcd)

	// RESET packet received
	case RESET:

		teolog.DebugV(MODULE, "got RESET packet, channel:", key)
		pac.ackToResetCreateNew().writeTo(tcd)
		tcd.reset()

	// ACK-to-reset packet received
	case ACKReset:

		teolog.DebugV(MODULE, "got ACK_RESET packet, channel:", key)
		tcd.reset()

	// PING packet received
	case PING:

		// Show Log
		teolog.DebugVf(MODULE, "got PING packet id: %d, channel: %s, data: %s\n",
			pac.getID(), key, string(pac.getData()),
		)
		// Create ACK to ping packet and send it back to sender
		pac.ackToPingCreateNew().writeTo(tcd)

	// ACK-to-PING packet received
	case ACKPing:

		// Show Log
		teolog.DebugVf(MODULE, "got ACK_PING packet id: %d, channel: %s, "+
			"triptime: %.3f ms\n",
			pac.getID(), key, tcd.stat.triptime,
		)

		// Set trip time to ChannelData
		triptime := pac.getTriptime()
		tcd.stat.setTriptime(triptime)

		// Send event to user level
		if tcd.trudp.allowEvents > 0 { // \TODO use GOT_ACK_PING to check allow this event
			tcd.trudp.sendEvent(tcd, GOT_ACK_PING, nil) // []byte(fmt.Sprintf("%.3f", triptime)))
		}

	// UNKNOWN packet received
	default:
		teolog.DebugV(MODULE, "UNKNOWN packet received, channel:", key,
			", type:", packetType,
		)
	}

	return
}

// packetDataProcess process received data packet, check receivedQueue and
// send received data and events to user level
func (pac *packetType) packetDataProcess(tcd *ChannelData) {
	id := pac.getID()
	switch {

	// Valid data packet
	case id == tcd.expectedID:
		tcd.expectedID++
		teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSILightGreen,
			fmt.Sprintf("received valid packet id: %d, channel: %s",
				int(id), tcd.GetKey())))
		// Send received packet data to user level
		tcd.trudp.sendEvent(tcd, GOT_DATA, pac.getData())
		// Check valid packets in received queue and send it data to user level
		tcd.receiveQueueProcess(func(data []byte) {
			tcd.trudp.sendEvent(tcd, GOT_DATA, data)
		})

	// Invalid packet (with id = 0)
	case id == firstPacketID:
		teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSILightRed,
			fmt.Sprintf("received invalid packet id: %d (expected id: %d), channel: %s, "+
				"reset locally", id, tcd.expectedID, tcd.GetKey())))
		tcd.reset()                // Reset local
		pac.packetDataProcess(tcd) // Process packet with id 0

	// Invalid packet (when expectedID = 0)
	case tcd.expectedID == firstPacketID:
		teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSILightRed,
			fmt.Sprintf("received invalid packet id: %d (expected id: %d), channel: %s, "+
				"send reset to remote host", id, tcd.expectedID, tcd.GetKey())))
		pac.resetCreateNew().writeTo(tcd) // Send reset
		// Send event "RESET was sent" to user level
		tcd.trudp.sendEvent(tcd, SEND_RESET, nil)

	// Already processed packet (id < expectedID)
	case id < tcd.expectedID:
		teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSILightBlue,
			fmt.Sprintf("skip received packet id: %d, channel: %s, "+
				"already processed", id, tcd.GetKey())))
		// Set statistic REJECTED (already received) packet
		tcd.stat.dropped()

	// Packet with id more than expectedID placed to receive queue and wait
	// previouse packets
	case id > tcd.expectedID:
		_, _, err := tcd.receiveQueueFind(id)
		if err != nil {
			teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSIYellow,
				fmt.Sprintf("put packet id: %d, channel: %s to received queue, "+
					"wait previouse packets", id, tcd.GetKey())))
			tcd.receiveQueueAdd(pac)
		} else {
			teolog.DebugV(MODULE, teokeys.Color(teokeys.ANSILightBlue,
				fmt.Sprintf("skip received packet id: %d, channel: %s, "+
					"already in receive queue", id, tcd.GetKey())))
			// Set statistic REJECTED (already received) packet
			tcd.stat.dropped()
		}
	}
}

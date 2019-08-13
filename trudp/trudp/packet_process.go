package trudp

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

//const MODULE = "trudp_packet"

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

		// \TODO: drop this packet if EQ len >= MaxValue
		// if len(pac.trudp.chanEvent) > 16 {
		// 	break
		// }
		// Create ACK packet and send it back to sender
		pac.ackCreateNew().writeTo(tcd)
		tcd.stat.received(len(pac.data))
		// Show Log
		teolog.Log(teolog.DEBUGv, MODULE, "DATA packet received, key:", key,
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
		teolog.Log(teolog.DEBUGv, MODULE, "ACK packet received, key:", key,
			"id:", pac.getID(),
			"trip time:", fmt.Sprintf("%.3f", tcd.stat.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.stat.triptimeMiddle), "ms")
		// Remove packet from send queue
		tcd.sendQueueRemove(pac)
		tcd.trudp.proc.writeQueueWriteTo(tcd)

	// RESET packet received
	case RESET:
		teolog.Log(teolog.DEBUGv, MODULE, "RESET packet received, key:", key)
		pac.ackToResetCreateNew().writeTo(tcd)
		tcd.reset()

	// ACK-to-reset packet received
	case ACKReset:
		teolog.Log(teolog.DEBUGv, MODULE, "ACK_RESET packet received, key:", key)
		tcd.reset()

	// PING packet received
	case PING:
		// Create ACK to ping packet and send it back to sender
		pac.ackToPingCreateNew().writeTo(tcd)
		// Show Log
		teolog.Log(teolog.DEBUGv, MODULE, "PING packet received, key:", key,
			"id:", pac.getID(),
			"expected id:", tcd.expectedID,
			"data:", pac.getData(), string(pac.getData()))

	// ACK-to-PING packet received
	case ACKPing:
		// Set trip time to ChannelData
		triptime := pac.getTriptime()
		tcd.stat.setTriptime(triptime)
		teolog.Log(teolog.DEBUGv, MODULE, "ACK_PING packet received, key:", key,
			"id:", pac.getID(),
			"trip time:", fmt.Sprintf("%.3f", tcd.stat.triptime), "ms",
			"trip time midle:", fmt.Sprintf("%.3f", tcd.stat.triptimeMiddle), "ms")
		if tcd.trudp.allowEvents > 0 { // \TODO use GOT_ACK_PING to check allow this event
			tcd.trudp.sendEvent(tcd, GOT_ACK_PING, nil) // []byte(fmt.Sprintf("%.3f", triptime)))
		}

	// UNKNOWN packet received
	default:
		teolog.Log(teolog.DEBUGv, MODULE, "UNKNOWN packet received, key:", key, ", type:", packetType)
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
func (pac *packetType) packetDataProcess(tcd *ChannelData) {
	id := pac.getID()
	switch {

	// Valid data packet
	case id == tcd.expectedID:
		tcd.expectedID++
		teolog.Log(teolog.DEBUGv, MODULE, teokeys.Color(teokeys.ANSILightGreen, "received valid packet id "+strconv.Itoa(int(id))))
		// Send received packet data to user level
		tcd.trudp.sendEvent(tcd, GOT_DATA, pac.getData())
		// Check packets in received queue and send it data to user level
		tcd.receiveQueueProcess(func(data []byte) { tcd.trudp.sendEvent(tcd, GOT_DATA, data) })

	// Invalid packet (with id = 0)
	case id == firstPacketID:
		teolog.Log(teolog.DEBUGv, MODULE, _ANSI_LIGHTRED+"received invalid packet id", id, "reset locally"+_ANSI_NONE)
		tcd.reset()
		pac.packetDataProcess(tcd)

	// Invalid packet (when expectedID = 0)
	case tcd.expectedID == firstPacketID:
		teolog.Log(teolog.DEBUGv, MODULE, _ANSI_LIGHTRED+"received invalid packet id", id, "send reset remote host"+_ANSI_NONE)
		pac.resetCreateNew().writeTo(tcd) // Send reset
		// Send event "RESET was sent" to user level
		tcd.trudp.sendEvent(tcd, SEND_RESET, nil)

	// Already processed packet (id < expectedID)
	case id < tcd.expectedID:
		teolog.Log(teolog.DEBUGv, MODULE, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already processed"+_ANSI_NONE)
		// Set statistic REJECTED (already received) packet
		tcd.stat.dropped()

	// Packet with id more than expectedID placed to receive queue and wait
	// previouse packets
	case id > tcd.expectedID:
		_, _, err := tcd.receiveQueueFind(id)
		if err != nil {
			teolog.Log(teolog.DEBUGv, MODULE, _ANSI_YELLOW+"move received packet to received queue, id", id, "wait previouse packets"+_ANSI_NONE)
			tcd.receiveQueueAdd(pac)
		} else {
			teolog.Log(teolog.DEBUGv, MODULE, _ANSI_LIGHTBLUE+"skipping received packet id", id, "already in receive queue"+_ANSI_NONE)
			// Set statistic REJECTED (already received) packet
			tcd.stat.dropped()
		}
	}
}

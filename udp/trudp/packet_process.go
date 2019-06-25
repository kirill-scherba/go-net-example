package trudp

import "net"

const (
	DATA      = iota //(0x0)
	ACK              //(0x1)
	RESET            //(0x2)
	ACK_RESET        //(0x3)
	PING             //(0x4)
	ACK_PING         //(0x5)
)

func (pac *packetType) process(packet []byte, addr net.Addr) (processed bool) {
	processed = false

	packetType := pac.getType(packet)
	switch packetType {

	// DATA packet received
	case DATA:
		pac.trudp.log(DEBUG_V, "DATA packet received")
		// Create ACK packet and send it back to sender
		packetACK, destroy := pac.ackCreateNew(packet)
		defer destroy()
		pac.trudp.conn.WriteTo(packetACK, addr)

	case ACK:
		pac.trudp.log(DEBUG_V, "ACK packet received")

	case RESET:
		pac.trudp.log(DEBUG_V, "RESET packet received")

	case ACK_RESET:
		pac.trudp.log(DEBUG_V, "ACK_RESET packet received")

		// PING packet received
	case PING:
		pac.trudp.log(DEBUG_V, "PING packet received")
		// Create ACK to ping packet and send it back to sender
		packetACKping, destroy := pac.ackToPingCreateNew(packet)
		defer destroy()
		pac.trudp.conn.WriteTo(packetACKping, addr)

	case ACK_PING:
		tripTime := float64(pac.trudp.getTimestamp()-pac.getTimestamp(packet)) / 1000.0
		pac.trudp.log(DEBUG_V, "ACK_PING packet received, trip time:", tripTime, "ms")

	default:
		pac.trudp.log(DEBUG_V, "UNKNOWN packet received, type:", packetType)
	}

	return
}

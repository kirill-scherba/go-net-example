package trudp

// Data packet type
type packetData struct {
	packetType
}

// Service packet type
type packetService struct {
	packetType
}

// destriy packet
func (packet packetType) destroy() {
	packet.trudp.packet.freeCreated(packet.data)
}

// writeTo send packetData to trudp channel
func (packet packetData) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	tcd.sendQueueProcess(func() { tcd.sendQueueAdd(&packet) })
}

// writeTo send packetService to trudp channel
func (packet packetService) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	packet.destroy()
}

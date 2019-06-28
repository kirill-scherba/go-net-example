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

// reWriteTo resend packetData from send queue and update record in send queue
// \TODO NOT USED
// func (packet packetData) reWriteTo(tcd *channelData) {
// 	data := packet.data
// 	packet.trudp.conn.WriteTo(data, tcd.addr)
// 	tcd.sendQueueProcess(func() { tcd.sendQueueUpdate(data) })
// }

// writeTo send packetService to trudp channel
func (packet packetService) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	packet.destroy()
}

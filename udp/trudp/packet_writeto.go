package trudp

type packetBase struct {
	data  []byte
	trudp *TRUDP
}

// Data packet type
type packetData struct {
	packetBase
}

// Service packet type
type packetService struct {
	packetBase
}

// destriy packet
func (packet packetBase) destroy() {
	packet.trudp.packet.freeCreated(packet.data)
}

// writeTo send packetData to trudp channel
func (packet packetData) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	// f := func() {
	// 	tcd.sendQueueAdd(&packet)
	// }
	tcd.sendQueueProcess(func() { tcd.sendQueueAdd(&packet) })
}

// reWriteTo resend packetData from send queue and update record in send queue
// \TODO NOT USED
func (packet packetData) reWriteTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	tcd.sendQueueProcess(func() { tcd.sendQueueUpdate(data) }) //&packet)
}

// writeTo send packetService to trudp channel
func (packet packetService) writeTo(tcd *channelData) {
	data := packet.data
	packet.trudp.conn.WriteTo(data, tcd.addr)
	packet.destroy()
}

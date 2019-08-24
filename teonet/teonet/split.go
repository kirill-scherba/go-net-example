package teonet

import "fmt"

// Module to Split / Combine Teonet packages

// splitPacket split module data structure
type splitPacket struct {
	teo *Teonet
	buf []byte
}

// splitNew create splitPacket receiver
func (teo *Teonet) splitNew() *splitPacket {
	return &splitPacket{teo: teo}
}

// combine got received packet and return combined teonet packet or nil if not
// combined yet or error
func (split *splitPacket) combine(rec *receiveData) (packet []byte, err error) {
	return
}

// cmdSplit CMD_SPLIT command processing
func (split *splitPacket) cmdSplit(rec *receiveData) (packet []byte, status bool) {
	split.teo.com.log(rec.rd, "CMD_SPLIT command")
	fmt.Printf("data_len: %d\n", rec.rd.DataLen())
	split.combine(rec)
	return
}

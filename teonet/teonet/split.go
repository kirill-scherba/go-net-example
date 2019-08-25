package teonet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"
)

// Module to Split / Combine Teonet packages

// splitPacket split module data structure
type splitPacket struct {
	teo *Teonet
	m   map[string]*receiveData
}

const (
	maxDataLen     = 448
	maxPacketLen   = 0x7FFFF * 2
	lastPacketFlag = 0x8000
)

// splitNew create splitPacket receiver
func (teo *Teonet) splitNew() *splitPacket {
	return &splitPacket{teo: teo, m: make(map[string]*receiveData)}
}

// combine got received packet and return combined teonet packet or nil if not
// combined yet or error
func (split *splitPacket) combine(rec *receiveData) (packet []byte, cmd int, err error) {

	// Parse command
	buf := bytes.NewReader(rec.rd.Data())
	le := binary.LittleEndian
	var packetNum, subpacketNum uint16
	binary.Read(buf, le, &packetNum)
	binary.Read(buf, le, &subpacketNum)
	lastPacket := subpacketNum&lastPacketFlag != 0
	subpacketNum = subpacketNum & (lastPacketFlag - 1)

	// makeKey create string to use it as key in map
	// Map key structure:
	//
	// uint8_t  from_length
	// char[]   from
	// uint16   packet_num
	// uint16   sub_packet_num
	//
	makeKey := func(from string, packetNum, subpacketNum uint16) string {
		return fmt.Sprintf("%s:%d:%d", from, packetNum, subpacketNum)
	}

	// addToMap add packet to map
	addToMap := func(key string, rec *receiveData) {
		split.m[key] = rec
	}

	// Save packet to map
	addToMap(makeKey(rec.rd.From(), packetNum, subpacketNum), rec)
	if !lastPacket {
		return
	}

	// Combine packet
	const ptr = int(unsafe.Sizeof(packetNum)) * 2
	for i := 0; i <= int(subpacketNum); i++ {
		key := makeKey(rec.rd.From(), packetNum, uint16(i))
		rec, ok := split.m[key]
		if !ok {
			err = errors.New("the subpacket has not received or added to the map")
			packet = nil
			return
		}
		data := rec.rd.Data()[ptr:]
		if i == 0 {
			l := len(data) - 1
			cmd = int(data[l])
			data = data[:l]
		}
		packet = append(packet, data...)
		delete(split.m, key)
	}

	return
}

// cmdSplit CMD_SPLIT command processing
func (split *splitPacket) cmdSplit(rec *receiveData) (processed bool, err error) {
	split.teo.com.log(rec.rd, "CMD_SPLIT command")
	data, cmd, err := split.combine(rec)
	if err != nil {
		err = errors.New("can't combine packets")
		return
	}
	if data == nil {
		processed = true
		return
	}
	rd, err := split.teo.packetCreateNew(cmd, rec.rd.From(), data).Parse()
	if err != nil {
		err = errors.New("can't parse combined packet")
		return
	}
	processed = split.teo.com.process(&receiveData{rd, rec.tcd})
	return
}

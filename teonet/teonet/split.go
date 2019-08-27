package teonet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

// Module to Split / Combine Teonet packages

// splitPacket split module data structure
type splitPacket struct {
	packetNum uint16
	teo       *Teonet
	m         map[string]*receiveData
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

// split spits data to subpackets and return number subpacket. For each
// subpacket the 'f func(data []byte)' callback function calls. If data len
// less than maxPacketLen num = 0 and callback function does not calls
func (split *splitPacket) split(cmd byte, data []byte, f func(cmd byte, data []byte)) (num int, err error) {

	// Send unsplit packet
	if len(data) < maxDataLen {
		f(cmd, data)
		return
	}

	split.packetNum++
	var subpacketNum uint16

	// callback Add command to first packet execute callback function and
	// increment number of subpacket couter
	callback := func(num *int, data []byte, lastSubpacket bool) {
		subpacketNum = uint16(*num)
		if lastSubpacket {
			subpacketNum = subpacketNum | lastPacketFlag
		}
		buf := new(bytes.Buffer)
		le := binary.LittleEndian
		binary.Write(buf, le, split.packetNum)
		binary.Write(buf, le, subpacketNum)
		binary.Write(buf, le, data)
		if subpacketNum == 0 {
			binary.Write(buf, le, byte(cmd))
		}
		f(CmdSplit, buf.Bytes())
		*num++
	}

	// Split data to subpackets and execute callback function
	for len(data) > 0 {
		l := len(data)
		last := l <= maxDataLen
		if !last {
			l = maxDataLen
		}
		callback(&num, data[:l], last)
		data = data[l:]
	}
	return
}

// combine got received packet and return combined teonet packet or nil if not
// combined yet or error
func (split *splitPacket) combine(rec *receiveData) (packet []byte, cmd byte, err error) {

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
			cmd = data[l]
			data = data[:l]
		}
		packet = append(packet, data...)
		delete(split.m, key)
	}

	return
}

// removeClient remove disconnected user from packets map
// \TODO: use this function when client disconneted
func (split *splitPacket) removeClient(client string) {
	for key, _ := range split.m {
		if strings.HasPrefix(key, client+":") {
			delete(split.m, key)
		}
	}
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
	rd, err := split.teo.packetCreateNew(rec.rd.From(), cmd, data).Parse()
	if err != nil {
		err = errors.New("can't parse combined packet")
		return
	}
	processed = split.teo.com.process(&receiveData{rd, rec.tcd})
	return
}

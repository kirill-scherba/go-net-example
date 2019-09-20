// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teoroom (teo-room: teonet room controller service) package
//
// Teoroom unites users to room and send commands between it

package teoroom

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
)

// Room controller commands
const (
	// ComRoomRequest [in] #129 Room request
	// [input] command for room controller
	ComRoomRequest = 129

	// ComRoomRequest [out] #129 Room request answer
	// [output] command from room controller
	ComRoomRequestAnswer = 129

	// ComRoomData [in,out] #130 Data transfer
	// [input or output] command for room controller
	ComRoomData = 130
)

type Teoroom struct {
	m map[string]*Client
}

type Client struct {
	name string
}

// Init initialize room controller
func Init() (troom *Teoroom, err error) {
	troom = &Teoroom{}
	troom.m = make(map[string]*Client)
	return
}

// Destroy close room controller
func (troom *Teoroom) Destroy() {

}

// Connect connects client to room
func Connect() (err error) {
	return
}

// Disconnec disconnects client from room
func Disconnec() (err error) {
	return
}

// Clients commands (commands executet in client) -----------------------------

// TeoConnector is teonet connector interface. It may be server (*Teonet) or
// client (*TeoLNull) connector
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
}

// RoomRequest send room request from client
func RoomRequest(con TeoConnector, peer string, i interface{}) {
	switch data := i.(type) {
	case nil:
		con.SendTo(peer, ComRoomRequest, nil)
	case []byte:
		con.SendTo(peer, ComRoomRequest, data)
	default:
		err := errors.New(fmt.Sprintf("Invalid type %T in SendTo function", data))
		panic(err)
	}
}

// SendData exchange data beatvean room member
func SendData(con TeoConnector, peer string, data ...interface{}) (err error) {
	buf := new(bytes.Buffer)
	for _, i := range data {
		switch what := i.(type) {
		case nil:
			err = binary.Write(buf, binary.LittleEndian, "nil")
		case []byte:
			err = binary.Write(buf, binary.LittleEndian, what)
		case encoding.BinaryMarshaler:
			d, err := what.MarshalBinary()
			if err != nil {
				panic(err)
			}
			err = binary.Write(buf, binary.LittleEndian, d)
		case int:
			err = binary.Write(buf, binary.LittleEndian, uint64(what))
		default:
			err = errors.New(fmt.Sprintf("Invalid type %T in SendTo function", data))
		}
		if err != nil {
			return
		}
	}
	d := buf.Bytes()
	con.SendTo(peer, ComRoomData, d)

	return
}

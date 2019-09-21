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

	// ComDisconnect [in] #131 Disconnect from room controller (from room)
	// [input] command for room controller
	ComDisconnect = 131
)

// Teoroom teonet room controller data
type Teoroom struct {
	m map[string]*Client
}

// Client data
type Client struct {
	name string
}

// Init initialize room controller
func Init() (tr *Teoroom, err error) {
	tr = &Teoroom{}
	tr.m = make(map[string]*Client)
	return
}

// Destroy close room controller
func (tr *Teoroom) Destroy() {

}

// RoomRequest request connect client to room
func (tr *Teoroom) RoomRequest(client string) (err error) {
	if _, ok := tr.m[client]; ok {
		err = errors.New("client already in room")
		return
	}
	tr.m[client] = &Client{}
	return
}

// GotData receive data from client
func (tr *Teoroom) GotData(client string, data []byte, f func(l0, client string, data []byte)) {
	for key := range tr.m {
		if key != client {
			f("", key, nil)
		}
	}
}

// Disconnec disconnects client from room
func (tr *Teoroom) Disconnect(client string) (err error) {
	if _, ok := tr.m[client]; !ok {
		err = errors.New("client not in room")
		return
	}
	delete(tr.m, client)
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
		err := fmt.Errorf("Invalid type %T in SendTo function", data)
		panic(err)
	}
}

// Disconnect send disconnect command
func Disconnect(con TeoConnector, peer string, i interface{}) {
	switch data := i.(type) {
	case nil:
		con.SendTo(peer, ComDisconnect, nil)
	case []byte:
		con.SendTo(peer, ComDisconnect, data)
	default:
		err := fmt.Errorf("Invalid type %T in SendTo function", data)
		panic(err)
	}
}

// SendData send data from client
func SendData(con TeoConnector, peer string, data ...interface{}) (err error) {
	buf := new(bytes.Buffer)
	for _, i := range data {
		switch what := i.(type) {
		case nil:
			err = binary.Write(buf, binary.LittleEndian, "nil")
		case []byte:
			err = binary.Write(buf, binary.LittleEndian, what)
		case encoding.BinaryMarshaler:
			var d []byte
			d, err = what.MarshalBinary()
			if err != nil {
				return
			}
			err = binary.Write(buf, binary.LittleEndian, d)
		case int:
			err = binary.Write(buf, binary.LittleEndian, uint64(what))
		default:
			err = fmt.Errorf("Invalid type %T in SendTo function", data)
		}
		if err != nil {
			return
		}
	}
	d := buf.Bytes()
	con.SendTo(peer, ComRoomData, d)

	return
}

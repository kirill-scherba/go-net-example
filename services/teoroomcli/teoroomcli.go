// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoroomcli is  teonet room controller service client package.
package teoroomcli

import (
	"bytes"
	"encoding"
	"encoding/binary"
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

	// ComStart [in] #132 Room started (got from room controller)
	// [input] command for room controller
	ComStart = 132
)

// Clients commands (commands executet in client) -----------------------------

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo method.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
}

// RoomRequest sends room request command
func RoomRequest(con TeoConnector, peer string, data interface{}) {
	switch d := data.(type) {
	case nil:
		con.SendTo(peer, ComRoomRequest, nil)
	case []byte:
		con.SendTo(peer, ComRoomRequest, d)
	default:
		err := fmt.Errorf("Invalid type %T in SendTo function", d)
		panic(err)
	}
}

// Disconnect sends disconnect from room (leave room) command
func Disconnect(con TeoConnector, peer string, data interface{}) {
	switch d := data.(type) {
	case nil:
		con.SendTo(peer, ComDisconnect, nil)
	case []byte:
		con.SendTo(peer, ComDisconnect, d)
	case byte:
		con.SendTo(peer, ComDisconnect, append([]byte{}, d))
	default:
		err := fmt.Errorf("Invalid type %T in SendTo function", d)
		panic(err)
	}
}

// Data sends data command
func Data(con TeoConnector, peer string, data ...interface{}) (num int, err error) {
	buf := new(bytes.Buffer)
	for _, i := range data {
		switch d := i.(type) {
		case nil:
			err = binary.Write(buf, binary.LittleEndian, "nil")
		case encoding.BinaryMarshaler:
			var dd []byte
			if dd, err = d.MarshalBinary(); err == nil {
				err = binary.Write(buf, binary.LittleEndian, dd)
			}
		case int:
			err = binary.Write(buf, binary.LittleEndian, uint64(d))
		default:
			err = binary.Write(buf, binary.LittleEndian, d)
		}
		if err != nil {
			return
		}
	}
	return con.SendTo(peer, ComRoomData, buf.Bytes())
}

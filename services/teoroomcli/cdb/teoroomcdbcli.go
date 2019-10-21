// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cdb is teonet teoroom cdb service client package.
package cdb

import (
	"bytes"
	"encoding/binary"
)

// RoomCreateRequest used in ComRoomCreated command
type RoomCreateRequest struct {
	RoomNum uint32
}

// MarshalBinary encodes RoomCreateRequest data into binary buffer.
func (rec *RoomCreateRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, rec.RoomNum)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomCreateRequest receiver data.
func (rec *RoomCreateRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &rec.RoomNum)
	return
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cdb is teonet teoroom cdb service client package.
package cdb

import (
	"bytes"
	"encoding/binary"

	"github.com/gocql/gocql"
)

// RoomCreateRequest used in ComRoomCreated command as request
type RoomCreateRequest struct {
	RoomID  gocql.UUID
	RoomNum uint32
}

// MarshalBinary encodes RoomCreateRequest data into binary buffer.
func (req *RoomCreateRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.RoomNum)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomCreateRequest receiver data.
func (req *RoomCreateRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &req.RoomID)
	if err != nil {
		return
	}
	err = binary.Read(buf, le, &req.RoomNum)
	return
}

// RoomCreateResponce used in ComRoomCreated command as responce
type RoomCreateResponce struct {
	RoomID gocql.UUID
}

// MarshalBinary encodes RoomCreateResponce data into binary buffer.
func (res *RoomCreateResponce) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, res.RoomID)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomCreateResponce receiver data.
func (res *RoomCreateResponce) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &res.RoomID)
	return
}

// RoomStatusRequest used in ComRoomStatus command as request
type RoomStatusRequest struct {
	RoomID gocql.UUID
	Status byte
}

// MarshalBinary encodes RoomCreateResponce data into binary buffer.
func (req *RoomStatusRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.Status)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomCreateResponce receiver data.
func (req *RoomStatusRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &req.RoomID)
	if err != nil {
		return
	}
	err = binary.Read(buf, le, &req.Status)
	return
}

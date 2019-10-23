// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package stats is teonet teoroom statistic (which writing to cdb) service
// client package.
package stats

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/gocql/gocql"
)

// Teoroom cdb commands
const (
	CmdRoomCreated  = iota + 134
	CmdRoomStatus   // 135
	CmdClientStatus // 136
)

// TeoCdb is Teonet teo-cdb peer name
var TeoCdb = "teo-cdb"

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo method.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
	//SendAnswer(pac *teonet.Packet, cmd byte, data []byte) (int, error)
	SendAnswer(pac interface{}, cmd byte, data []byte) (int, error)
	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2
	// second.
	// WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	// 	Data []byte
	// 	Err  error
	// }
}

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

// MarshalBinary encodes RoomStatusRequest data into binary buffer.
func (req *RoomStatusRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.Status)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomStatusRequest receiver data.
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

// SendRoomCreate sends RoomCreate to cdb
func SendRoomCreate(teo TeoConnector, roomID gocql.UUID, roomNum uint32) {
	req := &RoomCreateRequest{RoomID: roomID, RoomNum: roomNum}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdRoomCreated, data)
}

// SendRoomStatus sends RoomStatus to cdb
func SendRoomStatus(teo TeoConnector, roomID gocql.UUID, status byte) {
	req := &RoomStatusRequest{RoomID: roomID, Status: status}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdRoomStatus, data)
}

// Status of client status request
const (
	ClientAdded = iota
	ClientLeave
	ClientGameStat
)

// ClientStatusRequest used in ComClientStatus command as request
type ClientStatusRequest struct {
	Status   byte // 0 - Added; 1 - Leave; 2 - GameStat;
	RoomID   gocql.UUID
	ID       gocql.UUID
	GameStat []byte
}

// MarshalBinary encodes ClientStatusRequest data into binary buffer.
func (req *ClientStatusRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.Status)
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.ID)
	binary.Write(buf, le, req.GameStat)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into ClientStatusRequest receiver data.
func (req *ClientStatusRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &req.Status)
	err = binary.Read(buf, le, &req.RoomID)
	err = binary.Read(buf, le, &req.ID)
	if l := len(data) - int(unsafe.Sizeof(req.Status)+unsafe.Sizeof(req.RoomID)+
		unsafe.Sizeof(req.ID)); l > 0 {
		req.GameStat = make([]byte, l)
		binary.Read(buf, le, &req.GameStat)
	}
	return
}

// SendClientStatus sends ClientStatus to cdb
func SendClientStatus(teo TeoConnector, status byte, roomID gocql.UUID, id gocql.UUID, statAr ...[]byte) {
	var stat []byte
	if len(statAr) > 0 {
		stat = statAr[0]
	}
	req := &ClientStatusRequest{Status: status, RoomID: roomID, ID: id, GameStat: stat}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdClientStatus, data)
}

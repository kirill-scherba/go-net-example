// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package stats is teonet teoroom statistic (which writing to cdb) service
// client package.
package stats

import (
	"bytes"
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/gocql/gocql"
)

// Teoroom cdb commands
const (
	CmdSetRoomCreated = iota + 134 // 134 Set room created state
	CmdSetRoomState                // 135 Set room state changed
	CmdSetClientState              // 136 Set client state changed
	CmdRoomsByCreated              // 137 Get rooms by created time
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
	WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
		Data []byte
		Err  error
	}
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

// RoomStateRequest used in ComRoomStatus command as request
type RoomStateRequest struct {
	RoomID gocql.UUID
	Status byte
}

// MarshalBinary encodes RoomStatusRequest data into binary buffer.
func (req *RoomStateRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.Status)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomStatusRequest receiver data.
func (req *RoomStateRequest) UnmarshalBinary(data []byte) (err error) {
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
	teo.SendTo(TeoCdb, CmdSetRoomCreated, data)
}

// SendRoomState sends RoomStatus to cdb
func SendRoomState(teo TeoConnector, roomID gocql.UUID, status byte) {
	req := &RoomStateRequest{RoomID: roomID, Status: status}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdSetRoomState, data)
}

// State of client state request
const (
	ClientAdded = iota
	ClientLoadded
	ClientStarted
	ClientLeave
	ClientDisconnected
	ClientGameStat
)

// ClientStateRequest used in ComClientState command as request
type ClientStateRequest struct {
	State    byte // 0 - Added; 1 - Leave; 2 - GameStat;
	RoomID   gocql.UUID
	ID       gocql.UUID
	GameStat []byte
}

// MarshalBinary encodes ClientStateRequest data into binary buffer.
func (req *ClientStateRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.State)
	binary.Write(buf, le, req.RoomID)
	binary.Write(buf, le, req.ID)
	binary.Write(buf, le, req.GameStat)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into ClientStateRequest receiver data.
func (req *ClientStateRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	err = binary.Read(buf, le, &req.State)
	err = binary.Read(buf, le, &req.RoomID)
	err = binary.Read(buf, le, &req.ID)
	if l := len(data) - int(unsafe.Sizeof(req.State)+unsafe.Sizeof(req.RoomID)+
		unsafe.Sizeof(req.ID)); l > 0 {
		req.GameStat = make([]byte, l)
		binary.Read(buf, le, &req.GameStat)
	}
	return
}

// SendClientState sends ClientState to cdb
func SendClientState(teo TeoConnector, state byte, roomID gocql.UUID,
	id gocql.UUID, statAr ...[]byte) {
	var stat []byte
	if len(statAr) > 0 {
		stat = statAr[0]
	}
	req := &ClientStateRequest{State: state, RoomID: roomID, ID: id,
		GameStat: stat}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdSetClientState, data)
}

// RoomByCreatedRequest request room by created field
type RoomByCreatedRequest struct {
	ReqID uint32    // Request id
	From  time.Time // Time when room created
	To    time.Time // Time when room created
	Limit uint32    // Number of records to read
}

// MarshalBinary encodes RoomCreatedRequest data into binary buffer.
func (req *RoomByCreatedRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.ReqID)
	from, _ := req.From.MarshalBinary()
	binary.Write(buf, le, from)
	to, _ := req.To.MarshalBinary()
	binary.Write(buf, le, to)
	binary.Write(buf, le, req.Limit)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomCreatedRequest receiver data.
func (req *RoomByCreatedRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian

	// Create time buffer
	var t time.Time
	tdata, _ := t.MarshalBinary()
	tlen := len(tdata)
	tbuf := make([]byte, tlen)

	err = binary.Read(buf, le, &req.ReqID)
	if err != nil {
		return
	}

	err = binary.Read(buf, le, &tbuf)
	if err != nil {
		return
	}
	req.From.UnmarshalBinary(tbuf)

	err = binary.Read(buf, le, &tbuf)
	if err != nil {
		return
	}
	req.To.UnmarshalBinary(tbuf)

	err = binary.Read(buf, le, &req.Limit)
	return
}

// Room data structure
type Room struct {
	ID      gocql.UUID // Room ID
	RoomNum uint32     // Room number
	Created time.Time  // Time when room created
	Started time.Time  // Time when room started
	Closed  time.Time  // Time when room closed to add players
	Stopped time.Time  // Time when room stopped
	State   uint8      // Current rooms state
}

// RoomByCreatedResponce responce to room request
type RoomByCreatedResponce struct {
	ReqID uint32 // Request id
	Rooms []Room
}

// MarshalBinary encodes RoomByCreatedResponce data into binary buffer.
func (res *RoomByCreatedResponce) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, res.ReqID)
	for _, v := range res.Rooms {
		binary.Write(buf, le, v.ID)
		binary.Write(buf, le, v.RoomNum)
		created, _ := v.Created.MarshalBinary()
		// fmt.Println("Created:", created, len(created))
		started, _ := v.Started.MarshalBinary()
		closed, _ := v.Closed.MarshalBinary()
		stopped, _ := v.Stopped.MarshalBinary()
		binary.Write(buf, le, created)
		binary.Write(buf, le, started)
		binary.Write(buf, le, closed)
		binary.Write(buf, le, stopped)
		binary.Write(buf, le, v.State)
	}
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into RoomByCreatedResponce receiver data.
func (res *RoomByCreatedResponce) UnmarshalBinary(data []byte) (err error) {
	res.Rooms = nil
	var t time.Time
	le := binary.LittleEndian
	buf := bytes.NewReader(data)
	tdata, _ := t.MarshalBinary()
	tlen := len(tdata)

	err = binary.Read(buf, le, &res.ReqID)
	for i := 0; ; i++ {
		var id gocql.UUID
		err = binary.Read(buf, le, &id)
		if err != nil {
			err = nil
			break
		}

		res.Rooms = append(res.Rooms, Room{ID: id})
		err = binary.Read(buf, le, &res.Rooms[i].RoomNum)
		if err != nil {
			return
		}

		t := make([]byte, tlen)
		err = binary.Read(buf, le, &t)
		if err != nil {
			return
		}
		res.Rooms[i].Created.UnmarshalBinary(t)
		//
		err = binary.Read(buf, le, &t)
		if err != nil {
			return
		}
		res.Rooms[i].Started.UnmarshalBinary(t)
		//
		err = binary.Read(buf, le, &t)
		if err != nil {
			return
		}
		res.Rooms[i].Closed.UnmarshalBinary(t)
		//
		err = binary.Read(buf, le, &t)
		if err != nil {
			return
		}
		res.Rooms[i].Stopped.UnmarshalBinary(t)
		//
		err = binary.Read(buf, le, &res.Rooms[i].State)
		if err != nil {
			return
		}
	}
	return
}

// SendRoomByCreated sends RoomByCreated Request to cdb
func SendRoomByCreated(teo TeoConnector, from, to time.Time, limit uint32) (
	res RoomByCreatedResponce, err error) {
	req := &RoomByCreatedRequest{ReqID: limit, From: from, To: to, Limit: limit}
	data, _ := req.MarshalBinary()
	teo.SendTo(TeoCdb, CmdRoomsByCreated, data)
	if r := <-teo.WaitFrom(TeoCdb, CmdRoomsByCreated, func(data []byte) (rv bool) {
		//fmt.Println("check function body data:", data)
		if err = res.UnmarshalBinary(data); err == nil {
			//fmt.Println("check function unmarshalled res:", res)
			rv = res.ReqID == req.ReqID
		}
		return
	}); r.Err != nil {
		err = r.Err
	}
	return
}

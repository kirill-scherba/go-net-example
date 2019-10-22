// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroomcdb command processing module.

package cdb

import (
	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/cdb"
)

// Process receiver to process teousers commands
type Process struct{ *Rooms }

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo method.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
	//SendAnswer(pac *teonet.Packet, cmd byte, data []byte) (int, error)
	SendAnswer(pac interface{}, cmd byte, data []byte) (int, error)
	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2 second.
	// WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	// 	Data []byte
	// 	Err  error
	// }
}

// TeoPacket is teonet packet interface
type TeoPacket interface {
	Cmd() byte
	Data() []byte
}

// ComRoomCreated process setCreating, return room uuid.
//
// Input data (binary): room_num uint32.
//
// Output data (byte):  id gocql.uuid
func (p *Process) ComRoomCreated(pac TeoPacket) (err error) {
	req := &cdb.RoomCreateRequest{}
	req.UnmarshalBinary(pac.Data())
	err = p.setCreating(req.RoomID, req.RoomNum)
	if err != nil {
		return
	}
	res := &cdb.RoomCreateResponce{RoomID: req.RoomID}
	d, err := res.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// ComRoomStateChanged process state changed command
func (p *Process) ComRoomStateChanged(pac TeoPacket) (err error) {
	req := &cdb.RoomStatusRequest{}
	req.UnmarshalBinary(pac.Data())

	switch req.Status {
	case teoroom.RoomRunning:
		err = p.setRunning(req.RoomID)
	}

	// if err != nil {
	// 	return
	// }
	return
}

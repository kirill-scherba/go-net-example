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

	case teoroom.RoomClosed:
		err = p.setClosed(req.RoomID)

	case teoroom.RoomStopped:
		err = p.setStopped(req.RoomID)
	}
	return
}

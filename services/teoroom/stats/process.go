// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom stats command processing module.

package stats

import (
	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
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
	req := &stats.RoomCreateRequest{}
	req.UnmarshalBinary(pac.Data())
	err = p.setCreating(req.RoomID, req.RoomNum)
	if err != nil {
		return
	}
	res := &stats.RoomCreateResponce{RoomID: req.RoomID}
	d, err := res.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// ComRoomStateChanged process state changed command
func (p *Process) ComRoomStateChanged(pac TeoPacket) (err error) {
	req := &stats.RoomStateRequest{}
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

// ComClientStatus process rooms client state changed command
func (p *Process) ComClientStatus(pac TeoPacket) (err error) {
	req := &stats.ClientStateRequest{}
	req.UnmarshalBinary(pac.Data())
	switch req.State {

	case stats.ClientAdded:
		err = p.setAdded(req.RoomID, req.ID)

	case stats.ClientLoadded:
		err = p.setLoadded(req.RoomID, req.ID)

	case stats.ClientStarted:
		err = p.setStarted(req.RoomID, req.ID)

	case stats.ClientLeave:
		err = p.setLeave(req.RoomID, req.ID)

	case stats.ClientDisconnected:
		err = p.setDisconnected(req.RoomID, req.ID)
	}
	return
}

// ComGetRoomsByCreated get rooms request by Created, read data from database
// and return answer to request
func (p *Process) ComGetRoomsByCreated(pac TeoPacket) (rooms []stats.Room,
	err error) {
	req := &stats.RoomByCreatedRequest{}
	req.UnmarshalBinary(pac.Data())
	rooms, err = p.getByCreated(req.From, req.To, req.Limit)
	if err != nil {
		return
	}
	res := &stats.RoomByCreatedResponce{ReqID: req.ReqID, Rooms: rooms}
	d, err := res.MarshalBinary()

	teolog.Debugf(MODULE, "Got %d bytes RoomByCreated request, read data and "+
		"send %d bytes snswer\n", len(pac.Data()), len(d))

	// Sent answer
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	if err != nil {
		teolog.Errorf(MODULE, "Err SendAnswer: %s\n", err)
		return
	}
	err = res.UnmarshalBinary(d)
	if err != nil {
		teolog.Errorf(MODULE, "Err SendAnswer responce UnmarshalBinary: %s\n",
			err)
		return
	}
	return
}

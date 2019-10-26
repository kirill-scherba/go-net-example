// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoroom (teo-room) is the Teonet room controller service package
//
// Room controller used to connect users to rooms and send commands between it.
// This package used in server and client applications.
//
package teoroom

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Teorooms errors code
const (
	GetError = iota
	ConfigCdbDoesNotExists
	UnmarshalJSON
)

// Teoroom is room controller data
type Teoroom struct {
	teo      *teonet.Teonet // Pointer to teonet
	roomID   uint32         // Next room id
	creating []uint32       // Creating rooms slice with creating rooms id
	mroom    *mroomType     // Rooms map contained created rooms and map mutex
	mcli     *mcliType      // Clients map contain clients connected to room controller and map mutes
	*Process
}

// errorTeoroom is Teoroom errors data structure
type errorTeoroom struct {
	code int
	prob string
}

// Error returns error in string format.
func (e *errorTeoroom) Error() string {
	return fmt.Sprintf("%d - %s", e.code, e.prob)
}

// New creates teonet room controller.
func New(teo *teonet.Teonet) (tr *Teoroom, err error) {
	tr = &Teoroom{teo: teo, mcli: newMcli(), mroom: newMroom()}
	tr.Process = &Process{tr}
	return
}

// Destroy the room controller.
func (tr *Teoroom) Destroy() {}

// sendExistingData sends saved data of all connected and loaded clients to
// this new clientg
func (tr *Teoroom) sendExistingData(client string, f func(
	l0 *teonet.L0PacketData, client string, cmd byte, data []byte) error) (err error) {

	c, err := tr.mcli.get(client)
	if err != nil {
		return
	}
	roomID, cliID, err := c.roomClientID()
	if err != nil {
		return
	}
	r, err := tr.mroom.get(roomID)
	if err != nil {
		return
	}
	for id, cli := range r.client {
		if id != cliID && cli != nil {
			for _, d := range cli.data {
				f(c.L0PacketData, client, teoroomcli.ComRoomData, d)
			}
		}
	}
	return
}

// resendData resend client data to all clients connected to this room with him.
// Process data for first clients command: Save client data and send him data
// saved by existing and ready clients.
func (tr *Teoroom) resendData(client string, cmd byte, data []byte, f func(
	l0 *teonet.L0PacketData, client string, cmd byte, data []byte) error) (err error) {

	// If client does not exists in map - skip this request
	cli, err := tr.mcli.get(client)
	if err != nil {
		return
	}

	// Find client in room
	roomID, cliID, err := cli.roomClientID()
	if err != nil {
		return
	}

	// If client send first data than it looaded and ready to play - send him
	// data saved by existing and ready clients
	if cli.data == nil {
		fmt.Printf(
			"Client %s loaded and ready to play, roomID: %d, client id: %d\n",
			client, roomID, cliID)
		tr.sendExistingData(client, f)
		var r *Room
		if r, err = tr.mroom.get(roomID); err != nil {
			return
		}
		r.clientReady(cliID)
	}

	// Save client data (todo: this is first position only)
	if len(cli.data) == 0 {
		cli.data = append(cli.data, []byte{})
	}
	cli.data[0] = data

	// Send data to all (connected and loaded) clients except himself
	r, err := tr.mroom.get(roomID)
	if err != nil {
		return
	}
	for id, cli := range r.client {
		if id == cliID || cli == nil {
			continue
		} else if err = f(cli.L0PacketData, cli.name, cmd, data); err != nil {
			return
		}
	}

	return
}

// Process receiver to process teoroom commands
type Process struct{ tr *Teoroom }

// ComRoomRequest process clients request, connect to room controller,
// entering to room and send room request answer with client room id
// (room position)
func (p *Process) ComRoomRequest(pac *teonet.Packet) (err error) {

	// Find client in room map
	client := pac.From()
	if _, ok := p.tr.mcli.find(client); ok {
		err = fmt.Errorf("client %s is already in room", client)
		return
	}
	// Find or create room for client
	roomID, cliID, err := p.tr.newClient(pac).roomRequest()
	if err != nil {
		return
	}
	// Send answer with client room id (room position)
	data := append([]byte{}, byte(cliID))
	// TODO: do than SendAnswer can send answer to client
	// p.tr.teo.SendAnswer(pac, teoroomcli.ComRoomRequestAnswer, data)
	_, err = p.tr.teo.SendToClientAddr(pac.GetL0(), client,
		teoroomcli.ComRoomRequestAnswer, data)

	// Send client status ClientAdded to statistic
	r, err := p.tr.mroom.get(roomID)
	if err != nil {
		return
	}
	cli, err := p.tr.mcli.get(client)
	if err != nil {
		return err
	}
	cli.sendState(stats.ClientAdded, r.roomID)

	return
}

// ComRoomData process data received from client and resend it to all clients
// connected to room with this client
func (p *Process) ComRoomData(pac *teonet.Packet) (err error) {
	p.tr.resendData(pac.From(), teoroomcli.ComRoomData, pac.Data(), func(
		l0 *teonet.L0PacketData, client string, cmd byte, data []byte) (err error) {
		_, err = p.tr.teo.SendToClientAddr(l0, client, cmd, data)
		return
	})
	return
}

// ComDisconnect remove client from room
func (p *Process) ComDisconnect(pac interface{}) (err error) {

	// Process input interface
	var client string
	switch cli := pac.(type) {
	case string:
		client = cli
	case *teonet.Packet:
		client = cli.From()
	}

	// Find client
	cli, err := p.tr.mcli.get(client)
	if err != nil {
		return
	}

	// Resend command to players in this room and delete client from room
	roomID, cliID, err := cli.roomClientID()
	if err == nil {
		err = p.tr.resendData(client, teoroomcli.ComDisconnect, []byte{byte(cliID)}, func(
			l0 *teonet.L0PacketData, client string, cmd byte, data []byte) (err error) {
			_, err = p.tr.teo.SendToClientAddr(l0, client, cmd, data)
			return
		})
		var r *Room
		if r, err = p.tr.mroom.get(roomID); err != nil {
			return
		}
		r.client[cliID] = nil
	}
	p.tr.mcli.delete(client)

	return
}

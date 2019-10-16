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

	"github.com/kirill-scherba/teonet-go/services/teoroom/teoroomcli"
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
	teo      *teonet.Teonet     // Pointer to teonet
	roomID   int                // Next room id
	creating []int              // Creating rooms slice with creating rooms id
	mroom    map[int]*Room      // Rooms map contain created rooms
	mcli     map[string]*Client // Clients map contain clients connected to room controller
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
	tr = &Teoroom{teo: teo, mcli: make(map[string]*Client), mroom: make(map[int]*Room)}
	tr.Process = &Process{tr}
	return
}

// Destroy the room controller.
func (tr *Teoroom) Destroy() {}

// sendExistingData sends saved data of all connected and loaded clients to
// this new clientg
func (tr *Teoroom) sendExistingData(client string, f func(
	l0 *teonet.L0PacketData, client string, cmd byte, data []byte) error) {

	c, ok := tr.mcli[client]
	if !ok {
		return
	}
	roomID, cliID, _ := c.roomClientID()
	for id, cli := range tr.mroom[roomID].client {
		if id != cliID && cli != nil {
			for _, d := range cli.data {
				f(c.L0PacketData, client, teoroomcli.ComRoomData, d)
			}
		}
	}
}

// resendData resend client data to all clients connected to this room with him.
// Process data for first clients command: Save client data and send him data
// saved by existing and ready clients.
func (tr *Teoroom) resendData(client string, cmd byte, data []byte, f func(
	l0 *teonet.L0PacketData, client string, cmd byte, data []byte) error) (err error) {

	// If client does not exists in map - skip this request
	cli, ok := tr.mcli[client]
	if !ok {
		err = fmt.Errorf("client %s does not in room", client)
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
		tr.mroom[roomID].clientReady(cliID)
	}

	// Save client data (todo: this is first position only)
	if len(tr.mcli[client].data) == 0 {
		tr.mcli[client].data = append(tr.mcli[client].data, []byte{})
	}
	tr.mcli[client].data[0] = data

	// Send data to all (connected and loaded) clients except himself
	for id, cli := range tr.mroom[roomID].client {
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
	if _, ok := p.tr.mcli[client]; ok {
		err = fmt.Errorf("client %s is already in room", client)
		return
	}
	// Find or create room for client
	_, cliID, err := p.tr.newClient(pac).roomRequest()
	if err != nil {
		return
	}
	// Send answer with client room id (room position)
	data := append([]byte{}, byte(cliID))
	//p.tr.teo.SendAnswer(pac, teoroomcli.ComRoomRequestAnswer, data)
	_, err = p.tr.teo.SendToClientAddr(pac.GetL0(), client,
		teoroomcli.ComRoomRequestAnswer, data)

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
	cli, ok := p.tr.mcli[client]
	if !ok {
		err = fmt.Errorf("client %s not fount in room", client)
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
		p.tr.mroom[roomID].client[cliID] = nil
	}
	delete(p.tr.mcli, client)

	return
}

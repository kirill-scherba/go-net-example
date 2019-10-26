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
	"sync"

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
	teo      *teonet.Teonet     // Pointer to teonet
	roomID   uint32             // Next room id
	creating []uint32           // Creating rooms slice with creating rooms id
	mroom    *mroomType         // Rooms map contained created rooms and map mutex
	mcli     map[string]*Client // Clients map contain clients connected to room controller
	mclix    sync.RWMutex       // Clients map mutex
	*Process
}

type mroomType struct {
	m  map[uint32]*Room
	mx sync.RWMutex
}

// newMroom return new mroomType
func newMroom() *mroomType {
	return &mroomType{m: make(map[uint32]*Room)}
}

// find value by key from mroomType map
func (m *mroomType) find(key uint32) (val *Room, ok bool) {
	m.mx.RLock()
	val, ok = m.m[key]
	m.mx.RUnlock()
	return
}

// get value by key from mroomType map and return error if key does not not exists
func (m *mroomType) get(key uint32) (val *Room, err error) {
	var ok bool
	if val, ok = m.find(key); !ok {
		err = fmt.Errorf("roomID %d does not exists", key)
	}
	return
}

// set value by key from mroomType map
func (m *mroomType) set(key uint32, val *Room) {
	m.mx.Lock()
	m.m[key] = val
	m.mx.Unlock()
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
	tr = &Teoroom{teo: teo, mcli: make(map[string]*Client), mroom: newMroom()}
	tr.Process = &Process{tr}
	return
}

// Destroy the room controller.
func (tr *Teoroom) Destroy() {}

// sendExistingData sends saved data of all connected and loaded clients to
// this new clientg
func (tr *Teoroom) sendExistingData(client string, f func(
	l0 *teonet.L0PacketData, client string, cmd byte, data []byte) error) (err error) {

	c, ok := tr.mcli[client]
	if !ok {
		err = fmt.Errorf("client %s does not exists", client)
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
		var r *Room
		if r, err = tr.mroom.get(roomID); err != nil {
			return
		}
		r.clientReady(cliID)
	}

	// Save client data (todo: this is first position only)
	if len(tr.mcli[client].data) == 0 {
		tr.mcli[client].data = append(tr.mcli[client].data, []byte{})
	}
	tr.mcli[client].data[0] = data

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
	if _, ok := p.tr.mcli[client]; ok {
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
	//p.tr.teo.SendAnswer(pac, teoroomcli.ComRoomRequestAnswer, data)
	_, err = p.tr.teo.SendToClientAddr(pac.GetL0(), client,
		teoroomcli.ComRoomRequestAnswer, data)

	// Send client status ClientAdded to statistic
	r, err := p.tr.mroom.get(roomID)
	if err != nil {
		return
	}
	cli := p.tr.mcli[client]
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
		var r *Room
		if r, err = p.tr.mroom.get(roomID); err != nil {
			return
		}
		r.client[cliID] = nil
	}
	delete(p.tr.mcli, client)

	return
}

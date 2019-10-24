// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Room Data
type Room struct {
	tr     *Teoroom                 // Pointer to Teoroom receiver
	id     uint32                   // Room id
	roomID gocql.UUID               // Room UUID
	state  byte                     // Room state: 0 - creating; 1 - running; 2 - closed; 3 - stopped
	client []*Client                // List of clients in room by position: client[0] - position 1 ... client[0] - position 10
	cliwas map[string]*ClientInRoom // Map of clients which was in room (included clients connected now)
	gparam *GameParameters          // Game parameters
}

// ClientInRoom Data
type ClientInRoom struct {
	*Client
	state int // 1 - started
}

// Room state
const (
	RoomCreating = iota // 0 - Creating room state
	RoomRunning         // 1 - Running room state
	RoomClosed          // 2 - Closed room state: running but adding clients is prohibited
	RoomStopped         // 3 - Stopped room state (game over)
)

// newRoom creates new room.
func (tr *Teoroom) newRoom() (r *Room) {
	r = &Room{
		tr:     tr,
		id:     tr.roomID,
		roomID: gocql.TimeUUID(),
		cliwas: make(map[string]*ClientInRoom),
		state:  RoomCreating,
	}

	// Read game parameters from config and cdb
	// TDOD: get game name from new room requrest parameters
	r.newGameParameters("g001")

	// Save room create statistic to cdb
	r.sendStateCreate()

	fmt.Printf("Room id %d created, UUID: %s\n", r.id, r.roomID.String())

	tr.creating = append(tr.creating, r.id)
	tr.mroom[r.id] = r
	tr.roomID++
	return
}

// numClients return number of clients in room.
func (r *Room) numClients() (num int) {
	for _, cli := range r.client {
		if cli != nil {
			num++
		}
	}
	return
}

// addClient adds client to room.
func (r *Room) addClient(cli *Client) (clientID int) {
	r.client = append(r.client, cli)
	clientID = len(r.client) - 1
	r.cliwas[cli.name] = &ClientInRoom{cli, RoomCreating}
	fmt.Printf("Client %s added to room id %d, id in room: %d,\n",
		cli.name, r.id, clientID)
	return
}

// clientReady calls when client loaded, sends his start position and ready to
// run. When number of reday clients has reached `MinClientsToStart` the game
// started.
func (r *Room) clientReady(clientID int) {
	cli := r.client[clientID]
	cli.sendState(stats.ClientLoadded, r.roomID)

	// If room already closed or stoppet than send disconnect for this client
	if r.state == RoomClosed || r.state == RoomStopped {
		cli.sendState(stats.ClientDisconnected, r.roomID)
		r.tr.teo.SendToClientAddr(cli.L0PacketData, cli.name,
			teoroomcli.ComDisconnect, nil)
	}

	// Set client state 'running' in this room
	// TODO: possible there should be (or may be) enother constant
	r.cliwas[cli.name].state = RoomRunning

	// If room already started: send command ComStart to this new client.
	if r.state == RoomRunning {
		// TODO: send RoomClientAdd to room statistic
		fmt.Printf("Client %s added to running room id %d, id in room: %d\n",
			cli.name, r.id, clientID)
		cli.sendState(stats.ClientAdded, r.roomID)
		r.sendToClients(teoroomcli.ComStart, nil)
		return
	}

	// If room in Creating state: Check number of ready clients and start game
	// if it more than 'min clients to start' parameter
	var numReady int
	for _, cli := range r.cliwas {
		if cli.state == RoomRunning {
			numReady++
			if numReady >= r.gparam.MinClientsToStart {
				r.startRoom()
				break
			}
		}
	}
}

// startRoom calls when room started. Send ComStart command to all rooms clients
// and start goroutine which will send command ComDisconnect when room closed
// after `GameTime`.
func (r *Room) startRoom() {
	fmt.Printf("Room id %d started (game time: %d)\n", r.id, r.gparam.GameTime)
	r.setState(RoomRunning)

	r.sendToClients(teoroomcli.ComStart, nil)
	r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
		if cli, ok := r.tr.mcli[client]; ok {
			cli.sendState(stats.ClientStarted, r.roomID)
		}
	})

	// send disconnect to rooms clients after GameTime
	go func() {
		<-time.After(time.Duration(r.gparam.GameTime) * time.Millisecond)
		fmt.Printf("Room id %d stopped\n", r.id)
		r.setState(RoomStopped)
		r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
			if cli, ok := r.tr.mcli[client]; ok {
				cli.sendState(stats.ClientDisconnected, r.roomID)
			}
			r.tr.teo.SendToClientAddr(l0, client, teoroomcli.ComDisconnect, nil)
			r.tr.Process.ComDisconnect(client)
		})
	}()

	// set room state to Closed after CloseAfterTime
	if r.gparam.GameClosedAfter >= r.gparam.GameTime {
		return
	}
	go func() {
		<-time.After(time.Duration(r.gparam.GameClosedAfter) * time.Millisecond)
		if r.state == RoomStopped {
			return
		}
		fmt.Printf("Room id %d closed (close to add users)\n", r.id)
		r.setState(RoomClosed)
	}()
}

// funcToClients calls function for all room clients.
func (r *Room) funcToClients(f func(l0 *teonet.L0PacketData, client string)) {
	for _, cli := range r.client {
		if cli != nil {
			f(cli.L0PacketData, cli.name)
		}
	}
}

// sendToClients send command with data to all room clients.
func (r *Room) sendToClients(cmd int, data []byte) {
	r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
		r.tr.teo.SendToClientAddr(l0, client, byte(cmd), data)
	})
}

// SendRoomCreate sends room created state to cdb room statistic
func (r *Room) sendStateCreate() {
	stats.SendRoomCreate(r.tr.teo, r.roomID, r.id)
}

// sendState sends room state to cdb room statistic
func (r *Room) sendState() {
	stats.SendRoomState(r.tr.teo, r.roomID, r.state)
}

// setState set room state and send it to cdb room statistic
func (r *Room) setState(state byte) {
	r.state = state
	r.sendState()
}

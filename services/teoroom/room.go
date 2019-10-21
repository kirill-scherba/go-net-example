// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Room Data
type Room struct {
	tr     *Teoroom                 // Pointer to Teoroom receiver
	id     int                      // Room id
	state  int                      // Room state: 0 - creating; 1 - running; 2 - closed; 3 - stopped
	client []*Client                // List of clients in room by position: client[0] - position 1 ... client[0] - position 10
	cliwas map[string]*ClientInRoom // Map of clients which was in room (included clients connected now)
	gparam *GameParameters          // Game parameters
}

// Client data
type Client struct {
	tr                   *Teoroom // Pointer to Teoroom receiver
	name                 string   // Client name
	data                 [][]byte // Client data (which sends to new room clients)
	*teonet.L0PacketData          // Client L0 address
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
func (tr *Teoroom) newRoom() (room *Room) {
	room = &Room{
		tr:     tr,
		id:     tr.roomID,
		cliwas: make(map[string]*ClientInRoom),
		state:  RoomCreating,
	}
	room.newGameParameters("g001")
	tr.creating = append(tr.creating, room.id)
	tr.mroom[room.id] = room
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
	fmt.Printf("Client name: %s, id in room: %d, added to room id %d\n",
		cli.name, clientID, r.id)
	return
}

// clientReady calls when client loaded, sends his start position and ready to
// run. When number of reday clients has reached `MinClientsToStart` the game
// started.
func (r *Room) clientReady(cliID int) {
	client := r.client[cliID]

	// If room already closed or stoppet than find new room for client
	if r.state == RoomClosed || r.state == RoomStopped {
		// TODO: do somesing if room already closed or stoppet
	}

	// Set client state 'running' in this room
	// TODO: possible there should be enother constant
	r.cliwas[client.name].state = RoomRunning

	// If room already started: send command ComStart to this new client.
	if r.state == RoomRunning {
		// TODO:
		fmt.Printf("Client id %d added to running room id %d (game time: %d)\n",
			cliID, r.id, r.gparam.GameTime)
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
			}
		}
	}
}

// startRoom calls when room started. Send ComStart command to all rooms clients
// and start goroutine which will send command ComDisconnect when room closed
// after `GameTime`.
func (r *Room) startRoom() {
	r.state = RoomRunning
	fmt.Printf("Room id %d started (game time: %d)\n", r.id, r.gparam.GameTime)
	r.sendToClients(teoroomcli.ComStart, nil)
	// send disconnect to rooms clients after GameTime
	go func() {
		<-time.After(time.Duration(r.gparam.GameTime) * time.Millisecond)
		r.state = RoomStopped
		fmt.Printf("Room id %d closed\n", r.id)
		r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
			r.tr.teo.SendToClientAddr(l0, client, teoroomcli.ComDisconnect, nil)
			r.tr.Process.ComDisconnect(client)
		})
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

// newClient creates new Client (or add Client to Room controller).
func (tr *Teoroom) newClient(c *teonet.Packet) (cli *Client) {
	l0 := c.GetL0()
	client := c.From()
	cli = &Client{tr: tr, name: client, L0PacketData: l0}
	tr.mcli[client] = cli
	return
}

// roomRequest finds room for client or create new room and adds client to this
// room. It returns roomID and cliID or error if not found. The RoomID is an
// unical room number since this application started. The cliID is a client
// number (and position) in this room.
func (cli *Client) roomRequest() (roomID, cliID int, err error) {
	for _, rid := range cli.tr.creating {
		if r, ok := cli.tr.mroom[rid]; ok &&
			r.state != RoomClosed && r.state != RoomStopped &&
			func() bool { _, ok := r.cliwas[cli.name]; return !ok }() &&
			len(r.client) < r.gparam.MaxClientsInRoom {
			return r.id, r.addClient(cli), nil
		}
	}
	r := cli.tr.newRoom()
	// send disconnect to rooms clients if can't find clients during WaitForMinClients
	go func() {
		<-time.After(time.Duration(r.gparam.WaitForMinClients) * time.Millisecond)
		if r.state == RoomCreating &&
			r.numClients() < r.gparam.MinClientsToStart {
			r.state = RoomStopped
			fmt.Printf("Room id %d closed. "+
				"Can't find players during start timeout.\n", r.id)
			r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
				r.tr.teo.SendToClientAddr(l0, client, teoroomcli.ComDisconnect, nil)
				r.tr.Process.ComDisconnect(client)
			})
		}
	}()
	return r.id, r.addClient(cli), nil
}

// roomClientID finds client in rooms and returns roomID and cliID or error if
// not found. The RoomID is an unical room number since this application
// started. The cliID is a client number (and position) in this room.
func (cli *Client) roomClientID() (roomID, cliID int, err error) {
	var r *Room
	for roomID, r = range cli.tr.mroom {
		for id, c := range r.client {
			if c == cli {
				cliID = id
				return
			}
		}
	}
	err = fmt.Errorf("Can't find client %s in room structure", cli.name)
	return
}

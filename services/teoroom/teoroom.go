// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teoroom (teo-room: teonet room controller service) package
//
// Teoroom unites users to room and send commands between it

package teoroom

import (
	"errors"
	"fmt"
)

// Room controller commands
const (
	// ComRoomRequest [in] #129 Room request
	// [input] command for room controller
	ComRoomRequest = 129

	// ComRoomRequest [out] #129 Room request answer
	// [output] command from room controller
	ComRoomRequestAnswer = 129

	// ComRoomData [in,out] #130 Data transfer
	// [input or output] command for room controller
	ComRoomData = 130

	// ComDisconnect [in] #131 Disconnect from room controller (from room)
	// [input] command for room controller
	ComDisconnect = 131
)

// Room controller rooms constant
const (
	maxClientsInRoom = 10
)

// Teoroom teonet room controller data
type Teoroom struct {
	roomID   int                // Next room id
	creating []int              // Creating rooms: slice with creating rooms id
	mroom    map[int]*Room      // Rooms map contain created rooms
	mcli     map[string]*Client // Clients map contain clients connected to room controller
}

// Room Data
type Room struct {
	tr     *Teoroom  // Pointer to Teoroom receiver
	id     int       // Room id
	state  int       // Room state: 0 - creating; 1 - running; 2 - closed
	client []*Client // List of clients in room
}

// addClient adds client to room
func (r *Room) addClient(cli *Client) (clientID int) {
	r.client = append(r.client, cli)
	clientID = len(r.client) - 1
	fmt.Printf("Client name: %s, id in room: %d, added to room id %d\n",
		cli.name, clientID, r.id)
	return
}

// Client data
type Client struct {
	tr   *Teoroom // Pointer to Teoroom receiver
	name string   // Client name
	data []byte   // Client data (which sends to new room clients)
}

// Init initialize room controller
func Init() (tr *Teoroom, err error) {
	tr = &Teoroom{}
	tr.mcli = make(map[string]*Client)
	tr.mroom = make(map[int]*Room)
	//maxClientsInRoom
	return
}

// Destroy close room controller
func (tr *Teoroom) Destroy() {

}

// RoomRequest request connect client to room
func (tr *Teoroom) RoomRequest(client string) (roomID, cliID int, err error) {
	if _, ok := tr.mcli[client]; ok {
		err = fmt.Errorf("Client %s is already in room", client)
		return
	}
	return tr.clientNew(client).roomRequest()
}

// ResendData process data received from client and resend it to all connected
func (tr *Teoroom) ResendData(client string, data []byte, f func(l0, client string, data []byte)) {

	// If client does not exists in map - create it
	// \TODO it shoud be deprecated: we can't create new user without roomRequest
	if _, ok := tr.mcli[client]; !ok {
		tr.clientNew(client)
	}

	roomID, cliID, _ := tr.mcli[client].getRoomClientId()

	// If client send first data than it looaded and ready to play - send him "NewClient Data"
	if tr.mcli[client].data == nil {
		fmt.Printf("Client %s loaded and ready to play, roomID: %d, client id: %d\n",
			client, roomID, cliID)
		tr.NewClient(client, f)
	}

	// Save data
	tr.mcli[client].data = data

	// Send data to all (connected and loaded) clients except himself
	for key, c := range tr.mcli {
		if key != client && c.data != nil {
			f("", key, nil)
		}
	}
}

// NewClient send data of all connected and loaded clients to new client
func (tr *Teoroom) NewClient(client string, f func(l0, client string, data []byte)) {
	for key, c := range tr.mcli {
		if key != client && c.data != nil {
			f("", client, append(c.data, []byte(key)...))
		}
	}
}

// Disconnect disconnects client from room
func (tr *Teoroom) Disconnect(client string) (err error) {
	cli, ok := tr.mcli[client]
	if !ok {
		err = errors.New("client not in room")
		return
	}
	if roomID, cliID, err := cli.getRoomClientId(); err == nil {
		tr.mroom[roomID].client[cliID] = nil
	}
	delete(tr.mcli, client)
	return
}

// clientNew create new client
func (tr *Teoroom) roomNew() (r *Room) {
	r = &Room{tr: tr, id: tr.roomID}
	tr.creating = append(tr.creating, r.id)
	tr.mroom[r.id] = r
	tr.roomID++
	return
}

// clientNew create new client
func (tr *Teoroom) clientNew(client string) (cli *Client) {
	cli = &Client{tr: tr, name: client}
	tr.mcli[client] = cli
	return
}

// roomRequest finds room for client or create new
func (cli *Client) roomRequest() (roomID, cliID int, err error) {
	for _, rid := range cli.tr.creating {
		if r, ok := cli.tr.mroom[rid]; ok && len(r.client) < maxClientsInRoom {
			return r.id, r.addClient(cli), nil
		}
	}
	r := cli.tr.roomNew()
	return r.id, r.addClient(cli), nil
}

// getRoomClientId find client in rooms and return clients id
func (cli *Client) getRoomClientId() (roomID, cliID int, err error) {
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

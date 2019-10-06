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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
	"github.com/kirill-scherba/teonet-go/services/teoroom/teoroomcli"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Rooms constant default
const (
	maxClientsInRoom  = 10    // Maximum lients in room
	minClientsToStart = 2     // Minimum clients to start room
	waitForMinClients = 30000 // Wait for minimum clients connected
	waitForMaxClients = 10000 // Wait for maximum clients connected after minimum clients connected
	gameTime          = 12000 // Game time in millisecond = 2 min * 60 sec * 1000
	gameClosedAfter   = 30000 // Game closed after (does not add new clients)
)

// GameParameters holds game parameters running in room
type GameParameters struct {
	Name              string `json:"name,omitempty"`                 // Name of game
	GameTime          int    `json:"game_time,omitempty"`            // Game time in millisecond = 2 min * 60 sec * 1000
	GameClosedAfter   int    `json:"game_closed_after,omitempty"`    // Game closed after (does not add new clients)
	MaxClientsInRoom  int    `json:"max_clients_in_room,omitempty"`  // Maximum lients in room
	MinClientsToStart int    `json:"min_clients_to_start,omitempty"` // Minimum clients to start room
	WaitForMinClients int    `json:"wait_for_min_clients,omitempty"` // Wait for minimum clients connected
	WaitForMaxClients int    `json:"wait_for_max_clients,omitempty"` // Wait for maximum clients connected after minimum clients connected
}

// Teorooms errors code
const (
	GetError = iota
	ConfigCdbDoesNotExists
	UnmarshalJSON
)

// errorTeoroom is Teoroom errors data structure
type errorTeoroom struct {
	code int
	prob string
}

// Error returns error in string format
func (e *errorTeoroom) Error() string {
	return fmt.Sprintf("%d - %s", e.code, e.prob)
}

// newGameParameters create new GameParameters, sets default parameters and read
// parameters from config file
func (r *Room) newGameParameters(name string) (gp *GameParameters) {
	gp = &GameParameters{
		Name:              name,
		GameTime:          gameTime,
		GameClosedAfter:   gameClosedAfter,
		MaxClientsInRoom:  maxClientsInRoom,
		MinClientsToStart: minClientsToStart,
		WaitForMinClients: waitForMinClients,
		WaitForMaxClients: waitForMaxClients,
	}
	if err := gp.readConfig(); err != nil {
		fmt.Printf("Read game config error: %s\n", err)
	}
	// TODO: read game parameters from teo-cdb here
	// Read game parameters from teo-cdb and applay if changed, than write
	// it to config file
	//
	go func() {
		if err := gp.readConfigCdb(r.tr.teo); err != nil {
			fmt.Printf("Read cdb game config  error: %s\n", err)
			if err.code == ConfigCdbDoesNotExists {
				gp.writeConfigCdb(r.tr.teo)
			}
		}
		gp.writeConfig()
	}()

	r.gparam = gp
	return
}

// configDir return configuration files folder
func (gp *GameParameters) configDir() string {
	home := os.Getenv("HOME")
	return home + "/.config/teonet/teoroom/"
}

// readConfig reads game parameters from config file and replace current
// parameters
func (gp *GameParameters) readConfig() (err error) {
	fileName := gp.Name
	dirName := gp.configDir()
	f, err := os.Open(dirName + fileName + ".json")
	if err != nil {
		return
	}
	fi, err := f.Stat()
	if err != nil {
		return
	}
	data := make([]byte, fi.Size())
	if _, err = f.Read(data); err != nil {
		return
	}

	// Unmarshal json to the GameParameters structure
	if err = json.Unmarshal(data, gp); err == nil {
		fmt.Println("Game parameters read from file: ", gp)
	}

	return
}

// writeConfig writes game parameters to config file
func (gp *GameParameters) writeConfig() (err error) {
	fileName := gp.Name
	confDir := gp.configDir()
	f, err := os.Open(confDir + fileName + ".json")
	if err != nil {
		return
	}
	// Marshal json from the GameParameters structure
	data, err := json.Marshal(gp)
	if err != nil {
		return
	}
	_, err = f.Write(data)
	return
}

// configKeyCdb return configuration key
func (gp *GameParameters) configKeyCdb() string {
	return "conf.game." + gp.Name
}

// readConfigCdb read game parameters from config in teo-cdb
func (gp *GameParameters) readConfigCdb(con teocdbcli.TeoConnector) (errt *errorTeoroom) {

	// Create teocdb client
	cdb := teocdbcli.NewTeocdbCli(con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, gp.configKeyCdb())
	if err != nil {
		errt = &errorTeoroom{GetError, err.Error()}
		return
	} else if data == nil || len(data) == 0 {
		errt = &errorTeoroom{ConfigCdbDoesNotExists, "config does not exists"}
		return
	}

	// Unmarshal json to the GameParameters structure
	if err = json.Unmarshal(data, gp); err != nil {
		errt = &errorTeoroom{UnmarshalJSON, err.Error()}
		return
	}
	fmt.Println("Game parameters was read from cdb: ", gp)
	return
}

// writeConfigCdb writes game parameters to config in teo-cdb
func (gp *GameParameters) writeConfigCdb(con teocdbcli.TeoConnector) (err error) {

	// Create teocdb client
	cdb := teocdbcli.NewTeocdbCli(con)

	// Marshal json from the GameParameters structure
	data, err := json.Marshal(gp)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	_, err = cdb.Send(teocdbcli.CmdSet, gp.configKeyCdb(), data)
	return
}

// Teoroom is room controller data
type Teoroom struct {
	teo      *teonet.Teonet     // Pointer to teonet
	roomID   int                // Next room id
	creating []int              // Creating rooms slice with creating rooms id
	mroom    map[int]*Room      // Rooms map contain created rooms
	mcli     map[string]*Client // Clients map contain clients connected to room controller
}

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
	tr   *Teoroom // Pointer to Teoroom receiver
	name string   // Client name
	data [][]byte // Client data (which sends to new room clients)
}

// ClientInRoom Data
type ClientInRoom struct {
	*Client
	state int // 1 - started
}

// Room state
const (
	RoomCreating = iota // Creating room state
	RoomRunning         // Running room state
	RoomClosed          // Closed room state: running but adding clients is prohibited
	RoomStopped         // Stopped room state (game over)
)

// New teonet room controller
func New(teo *teonet.Teonet) (tr *Teoroom, err error) {
	return &Teoroom{teo: teo, mcli: make(map[string]*Client),
		mroom: make(map[int]*Room)}, nil
}

// Destroy close room controller
func (tr *Teoroom) Destroy() {}

// numClients return number of clients in room
func (r *Room) numClients() (num int) {
	for _, cli := range r.client {
		if cli != nil {
			num++
		}
	}
	return
}

// addClient adds client to room
func (r *Room) addClient(cli *Client) (clientID int) {
	r.client = append(r.client, cli)
	clientID = len(r.client) - 1
	r.cliwas[cli.name] = &ClientInRoom{cli, RoomCreating}
	fmt.Printf("Client name: %s, id in room: %d, added to room id %d\n",
		cli.name, clientID, r.id)
	return
}

// clientReady calls when client loaded, send his position and ready to run
func (r *Room) clientReady(cliID int) {
	client := r.client[cliID]
	r.cliwas[client.name].state = RoomRunning
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

// startRoom calls when room started
func (r *Room) startRoom() {
	r.state = RoomRunning
	fmt.Printf("Room id %d started (time: %d)\n", r.id, r.gparam.GameTime)
	r.sendToClients(teoroomcli.ComStart, nil)
	// send disconnect to rooms clients after GameTime
	go func() {
		<-time.After(time.Duration(r.gparam.GameTime) * time.Millisecond)
		r.state = RoomStopped
		fmt.Printf("Room id %d closed\n", r.id)
		r.funcToClients(nil, func(l0, client string, data []byte) {
			r.tr.teo.SendToClient("teo-l0", client, teoroomcli.ComDisconnect, data)
			r.tr.Disconnect(client)
		})
	}()
}

// funcToClients calls function for all room clients
func (r *Room) funcToClients(data []byte, f func(l0, client string, data []byte)) {
	for _, cli := range r.client {
		if cli != nil {
			f("", cli.name, data)
		}
	}
}

// sendToClients send command with data to all room clients
func (r *Room) sendToClients(cmd int, data []byte) {
	r.funcToClients(data, func(l0, client string, data []byte) {
		r.tr.teo.SendToClient("teo-l0", client, byte(cmd), data)
	})
}

// RoomRequest requests client connection to room controller and enterint to room
func (tr *Teoroom) RoomRequest(client string) (roomID, cliID int, err error) {
	if _, ok := tr.mcli[client]; ok {
		err = fmt.Errorf("Client %s is already in room", client)
		return
	}
	return tr.newClient(client).roomRequest()
}

// ResendData process data received from client and resend it to all clients
// connected to room with this client
func (tr *Teoroom) ResendData(client string, data []byte, f func(l0,
	client string, data []byte)) {

	// If client does not exists in map - skip this request
	if _, ok := tr.mcli[client]; !ok {
		return
	}

	roomID, cliID, _ := tr.mcli[client].getRoomClientID()

	// If client send first data than it looaded and ready to play - send him
	// Existing Data (existing clients saved data )"
	if tr.mcli[client].data == nil {
		fmt.Printf(
			"Client %s loaded and ready to play, roomID: %d, client id: %d\n",
			client, roomID, cliID)
		tr.sendExistingData(client, f)
		tr.mroom[roomID].clientReady(cliID)
	}

	// Save data
	if len(tr.mcli[client].data) == 0 {
		tr.mcli[client].data = append(tr.mcli[client].data, []byte{})
	}
	tr.mcli[client].data[0] = data

	// Send data to all (connected and loaded) clients except himself
	for id, cli := range tr.mroom[roomID].client {
		if id != cliID && cli != nil {
			f("", cli.name, data)
		}
	}
}

// sendExistingData sends saved data of all connected and loaded clients to
// this new client
func (tr *Teoroom) sendExistingData(client string, f func(l0, client string,
	data []byte)) {

	c, ok := tr.mcli[client]
	if !ok {
		return
	}
	roomID, cliID, _ := c.getRoomClientID()
	for id, cli := range tr.mroom[roomID].client {
		if id != cliID && cli != nil {
			for _, d := range cli.data {
				f("", client, d)
			}
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
	if roomID, cliID, err := cli.getRoomClientID(); err == nil {
		tr.mroom[roomID].client[cliID] = nil
	}
	delete(tr.mcli, client)
	return
}

// newRoom create new room
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

// newClient create new Client
func (tr *Teoroom) newClient(client string) (cli *Client) {
	cli = &Client{tr: tr, name: client}
	tr.mcli[client] = cli
	return
}

// roomRequest finds room for client or create new and add client to room
func (cli *Client) roomRequest() (roomID, cliID int, err error) {
	for _, rid := range cli.tr.creating {
		if r, ok := cli.tr.mroom[rid]; ok &&
			r.state != RoomClosed && r.state != RoomStopped &&
			func() bool { _, ok := r.cliwas[cli.name]; return !ok }() &&
			len(r.client) < maxClientsInRoom {
			return r.id, r.addClient(cli), nil
		}
	}
	r := cli.tr.newRoom()
	// send disconnect to rooms clients if can't find clients during WaitForMinClients
	go func() {
		<-time.After(time.Duration(r.gparam.WaitForMinClients) * time.Millisecond)
		if r.numClients() < r.gparam.MinClientsToStart {
			r.state = RoomStopped
			fmt.Printf("Room id %d closed\n", r.id)
			r.funcToClients(nil, func(l0, client string, data []byte) {
				r.tr.teo.SendToClient("teo-l0", client, teoroomcli.ComDisconnect, data)
				r.tr.Disconnect(client)
			})
		}
	}()
	return r.id, r.addClient(cli), nil
}

// getRoomClientID find client in rooms and return clients id
func (cli *Client) getRoomClientID() (roomID, cliID int, err error) {
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

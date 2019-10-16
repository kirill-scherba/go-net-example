// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"os"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teoconf"
)

// Rooms GameParameters constant default
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

type paramConf struct {
	*teoconf.Teoconf
}

// newGameParameters create new GameParameters, sets default parameters and read
// parameters from config file
func (r *Room) newGameParameters(name string) (p *paramConf) {

	gp := &GameParameters{
		Name:              name,
		GameTime:          gameTime,
		GameClosedAfter:   gameClosedAfter,
		MaxClientsInRoom:  maxClientsInRoom,
		MinClientsToStart: minClientsToStart,
		WaitForMinClients: waitForMinClients,
		WaitForMaxClients: waitForMaxClients,
	}

	p = &paramConf{teoconf.New(r.tr.teo, gp)}

	r.gparam = gp
	return
}

// Default return default value in json format.
func (gp *GameParameters) Default() []byte {
	return nil
}

// Value real value as interfaxe
func (gp *GameParameters) Value() interface{} {
	return gp
}

// Dir return configuration file folder.
func (gp *GameParameters) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teoroom/"
}

// FileName return configuration file name.
func (gp *GameParameters) FileName() string {
	return gp.Name
}

// Key return configuration key.
func (gp *GameParameters) Key() string {
	return "conf.game." + gp.FileName()
}

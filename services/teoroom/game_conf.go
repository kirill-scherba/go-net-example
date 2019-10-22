// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"os"

	"github.com/kirill-scherba/teonet-go/services/teocdbcli/conf"
)

// Rooms GameParameters constant default
const (
	maxClientsInRoom  = 10    // Maximum lients in room
	minClientsToStart = 2     // Minimum clients to start room
	waitForMinClients = 30000 // Wait for minimum clients connected
	waitForMaxClients = 10000 // Wait for maximum clients connected after minimum clients connected
	gameTime          = 12000 // Game time in millisecond = 2 min * 60 sec * 1000
	gameClosedAfter   = 10000 // Game closed after (does not add new clients after closed)
)

// GameParameters holds game parameters running in room
type GameParameters struct {
	Name              string `json:"name,omitempty"`                 // Name of game
	GameTime          int    `json:"game_time,omitempty"`            // Game time in millisecond = 2 min * 60 sec * 1000
	GameClosedAfter   int    `json:"game_closed_after,omitempty"`    // Game closed after (does not add new clients after closed)
	MaxClientsInRoom  int    `json:"max_clients_in_room,omitempty"`  // Maximum lients in room
	MinClientsToStart int    `json:"min_clients_to_start,omitempty"` // Minimum clients to start room
	WaitForMinClients int    `json:"wait_for_min_clients,omitempty"` // Wait for minimum clients connected
	WaitForMaxClients int    `json:"wait_for_max_clients,omitempty"` // Wait for maximum clients connected after minimum clients connected
}

//ConfHolder is configuration methods holder
type ConfHolder struct {
	*GameParameters
}

// newGameParameters create new GameParameters, sets default parameters and read
// parameters from config file
func (r *Room) newGameParameters(name string) (p *conf.Teoconf) {

	r.gparam = &GameParameters{
		Name:              name,
		GameTime:          gameTime,
		GameClosedAfter:   gameClosedAfter,
		MaxClientsInRoom:  maxClientsInRoom,
		MinClientsToStart: minClientsToStart,
		WaitForMinClients: waitForMinClients,
		WaitForMaxClients: waitForMaxClients,
	}

	p = conf.New(r.tr.teo, &ConfHolder{GameParameters: r.gparam})
	return
}

// Default return default value in json format.
func (c *ConfHolder) Default() []byte {
	return nil
}

// Value real value as interfaxe
func (c *ConfHolder) Value() interface{} {
	return c.GameParameters
}

// Dir return configuration file folder.
func (c *ConfHolder) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teoroom/"
}

// Name return configuration file name.
func (c *ConfHolder) Name() string {
	return c.GameParameters.Name
}

// Key return configuration key.
func (c *ConfHolder) Key() string {
	return "conf.game." + c.Name()
}

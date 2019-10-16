// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
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
	f, err := os.Create(confDir + fileName + ".json")
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
	cdb := teocdbcli.New(con)

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
	cdb := teocdbcli.New(con)

	// Marshal json from the GameParameters structure
	data, err := json.Marshal(gp)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	_, err = cdb.Send(teocdbcli.CmdSet, gp.configKeyCdb(), data)
	return
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 configuration parameters module.
//
// Read-write teonet L0 configuration from file and from teo-cdb.

// TODO: create separate service package to read config using this module.
// Now it uses here and in teoroom package  gameparameters module

package teonet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
)

var (
	// ErrConfigCdbDoesNotExists error returns by cdb config read function when
	// config does not exists in cdb
	ErrConfigCdbDoesNotExists = errors.New("config does not exists")
)

// parameters is l0 configuration parameters
type parameters struct {
	l0     *l0Conn
	Descr  string   // L0 configuration parameters description
	Prefix []string // Prefixes allowed quick registration with teonet
}

// parametersNew initialize parameters module
func (l0 *l0Conn) parametersNew() (lp *parameters) {
	lp = &parameters{l0: l0}
	lp.readConfigAndCdb()
	return
}

// read sets default parameters, then read parameters from local config file,
// than read parameters from teo-cdb and save it to local file
func (lp *parameters) readConfigAndCdb() (err error) {

	// TODO: Set defaults
	// gp = &GameParameters{
	// 	Name:              name,
	// 	GameTime:          gameTime,
	// 	GameClosedAfter:   gameClosedAfter,
	// 	MaxClientsInRoom:  maxClientsInRoom,
	// 	MinClientsToStart: minClientsToStart,
	// 	WaitForMinClients: waitForMinClients,
	// 	WaitForMaxClients: waitForMaxClients,
	// }

	// Read from local file
	if err := lp.readConfig(); err != nil {
		fmt.Printf("Read l0 config error: %s\n", err)
	}

	// Read parameters from teo-cdb and applay it if changed, than write
	// it to local config file
	go func() {
		if err := lp.readConfigCdb(); err != nil {
			fmt.Printf("Read cdb config error: %s\n", err)
			if err == ErrConfigCdbDoesNotExists {
				lp.writeConfigCdb(lp.l0.teo)
			}
		}
		if err = lp.writeConfig(); err != nil {
			fmt.Printf("Write config error: %s\n", err)
		}
	}()

	return
}

// eventProcess process teonet events to get teo-cdb connected and read config
func (lp *parameters) eventProcess(ev *EventData) {
	// Pocss event #3:  New peer connected to this host
	if ev.Event == EventConnected && ev.Data.From() == "teo-cdb" {
		fmt.Printf("Teo-cdb peer connectd. Read config...\n")
		if err := lp.readConfigAndCdb(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}

// configDir return configuration files folder
func (lp *parameters) configDir() string {
	home := os.Getenv("HOME")
	return home + "/.config/teonet/teol0/"
}

func (lp *parameters) configName() string {
	return "l0"
}

// readConfig reads l0 parameters from config file and replace current
// parameters
func (lp *parameters) readConfig() (err error) {
	f, err := os.Open(lp.configDir() + lp.configName() + ".json")
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

	// Unmarshal json to the l0 parameters structure
	if err = json.Unmarshal(data, lp); err == nil {
		fmt.Println("l0 parameters was read from file: ", lp)
	}

	return
}

// writeConfig writes game parameters to config file
func (lp *parameters) writeConfig() (err error) {
	f, err := os.Create(lp.configDir() + lp.configName() + ".json")
	if err != nil {
		return
	}
	// Marshal json from the parameters structure
	data, err := json.Marshal(lp)
	if err != nil {
		return
	}
	_, err = f.Write(data)
	return
}

// configKeyCdb return configuration key
func (lp *parameters) configKeyCdb() string {
	return "conf.network.l0"
}

// readConfigCdb read l0 parameters from config in teo-cdb
func (lp *parameters) readConfigCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.NewTeocdbCli(lp.l0.teo)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, lp.configKeyCdb())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = ErrConfigCdbDoesNotExists
		return
	}

	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, lp); err != nil {
		return
	}
	fmt.Println("l0 parameters was read from teo-cdb: ", lp)
	return
}

// writeConfigCdb writes l0 parameters config to teo-cdb
func (lp *parameters) writeConfigCdb(con teocdbcli.TeoConnector) (err error) {

	// Create teocdb client
	cdb := teocdbcli.NewTeocdbCli(con)

	// Marshal json from the parameters structure
	data, err := json.Marshal(lp)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	_, err = cdb.Send(teocdbcli.CmdSet, lp.configKeyCdb(), data)
	return
}

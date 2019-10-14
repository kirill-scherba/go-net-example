// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 configuration parameters module.
//
// Read-write teonet L0 configuration from file and from teo-cdb.

package teonet

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
)

// parameters is l0 configuration parameters
type parameters struct {
	l0     *l0Conn
	Descr  string   // L0 configuration parameters description
	Prefix []string // Prefixes allowed quick registration with teonet
}

// parametersNew initialize parameters module
func (l0 *l0Conn) parametersNew() *parameters {
	return &parameters{l0: l0}
}

// eventProcess process teonet events to get teo-cdb connected and read config
func (lp *parameters) eventProcess(ev *EventData) {
	// Pocss event #3:  New peer connected to this host
	if ev.Event == EventConnected && ev.Data.From() == "teo-cdb" {
		fmt.Printf("Teo-cdb peer connectd. Read config...\n")
		lp.readConfigCdb(lp.l0.teo)
	}
}

// configKeyCdb return configuration key
func (lp *parameters) configKeyCdb() string {
	return "conf.network.l0"
}

// readConfigCdb read l0 parameters from config in teo-cdb
func (lp *parameters) readConfigCdb(con teocdbcli.TeoConnector) (err error) {

	// Create teocdb client
	cdb := teocdbcli.NewTeocdbCli(con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, lp.configKeyCdb())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = errors.New("config does not exists")
		return
	}

	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, lp); err != nil {
		return
	}
	fmt.Println("l0 parameters was read from teo-cdb: ", lp)
	return
}

// // writeConfigCdb writes game parameters to config in teo-cdb
// func (gp *GameParameters) writeConfigCdb(con teocdbcli.TeoConnector) (err error) {

// 	// Create teocdb client
// 	cdb := teocdbcli.NewTeocdbCli(con)

// 	// Marshal json from the GameParameters structure
// 	data, err := json.Marshal(gp)
// 	if err != nil {
// 		return
// 	}

// 	// Send config to teo-cdb
// 	_, err = cdb.Send(teocdbcli.CmdSet, gp.configKeyCdb(), data)
// 	return
// }

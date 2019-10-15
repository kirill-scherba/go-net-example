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
	"fmt"
	"os"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teoconf"
)

// Config example
// {"key":"conf.network.l0","id":7,"value":{"descr":"Normal network L0 server","prefix":["tg001"]}}

// parameters is l0 configuration parameters
type parameters struct {
	Descr  string   // L0 configuration parameters description
	Prefix []string // Prefixes allowed quick registration with teonet
	//
	*teoconf.Teoconf
}

// parametersNew initialize parameters module
func (l0 *l0Conn) parametersNew() (lp *parameters) {
	lp = &parameters{}
	lp.Teoconf = teoconf.New(l0.teo, &confLocation{}, lp)
	return
}

// eventProcess process teonet events to get teo-cdb connected and read config
func (lp *parameters) eventProcess(ev *EventData) {
	// Pocss event #3:  New peer connected to this host
	if ev.Event == EventConnected && ev.Data.From() == "teo-cdb" {
		fmt.Printf("Teo-cdb peer connectd. Read config...\n")
		if err := lp.ReadBoth(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}

type confLocation struct {
}

// configDir return configuration file folder
func (where *confLocation) ConfigDir() string {
	return os.Getenv("HOME") + "/.config/teonet/teol0/"
}

// ConfigName return configuration file name
func (where *confLocation) ConfigName() string {
	return "l0"
}

// configKeyCdb return configuration key
func (where *confLocation) ConfigKeyCdb() string {
	return "conf.network." + where.ConfigName()
}

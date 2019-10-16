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
// {"descr":"Normal network 1 L0 server","prefix":["tg001","tg002","tg003"]}

// parameters is l0 configuration parameters
type parameters struct {
	Descr  string   // L0 configuration parameters description
	Prefix []string // Prefixes allowed quick registration with teonet
}

type paramConf struct {
	*teoconf.Teoconf
}

// funConf is functions receiver
type funConf struct{}

// parametersNew initialize parameters module
func (l0 *l0Conn) parametersNew() (p *paramConf) {
	fun := &funConf{}
	val := &parameters{}
	p = &paramConf{teoconf.New(l0.teo, val, fun)}
	return
}

// eventProcess process teonet events to get teo-cdb connected and read config
func (p *paramConf) eventProcess(ev *EventData) {
	// Pocss event #3:  New peer connected to this host
	if ev.Event == EventConnected && ev.Data.From() == "teo-cdb" {
		fmt.Printf("Teo-cdb peer connectd. Read config...\n")
		if err := p.ReadBoth(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
}

// Default return default value in json format
func (p *funConf) Default() []byte {
	return []byte(`{"descr":"Normal network L0 server","prefix":["tg001"]}`)
}

// Dir return configuration file folder
func (p *funConf) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teol0/"
}

// Name return configuration file name
func (p *funConf) Name() string {
	return "l0"
}

// Key return configuration key
func (p *funConf) Key() string {
	return "conf.network." + p.Name()
}

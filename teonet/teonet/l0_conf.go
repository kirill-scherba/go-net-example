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

	"github.com/kirill-scherba/teonet-go/services/teocdbcli/conf"
)

// param is l0 configuration parameters.
type param struct {
	Descr  string   // L0 configuration parameters description
	Prefix []string // Prefixes allowed quick registration with teonet
}

// paramConf is module receiver.
type paramConf struct {
	*conf.Teoconf
}

// parametersNew initialize parameters module.
func (l0 *l0Conn) parametersNew() (p *paramConf) {
	p = &paramConf{conf.New(l0.teo, &param{})}
	return
}

// eventProcess process teonet events to get teo-cdb connected and read config.
func (p *paramConf) eventProcess(ev *EventData) {
	if p == nil {
		return
	}
	// Pocss event #3:  New peer connected to this host
	if ev.Event == EventConnected && ev.Data.From() == "teo-cdb" {
		fmt.Printf("Teo-cdb peer connectd. Read config...\n")
		if err := p.ReadBoth(); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		var v = p.Value().(*param)
		fmt.Printf("Descr: %s\n", v.Descr)
	}
}

// Default return default value in json format.
func (p *param) Default() []byte {
	return []byte(`{"descr":"Normal network L0 server","prefix":["tg001"]}`)
}

// Struct real value as interfaxe
func (p *param) Value() interface{} {
	return p
}

// Dir return configuration file folder.
func (p *param) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teol0/"
}

// Name return configuration file name.
func (p *param) Name() string {
	return "l0"
}

// Key return configuration key.
func (p *param) Key() string {
	return "conf.network." + p.Name()
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Config is teocdb application module. It read teocdb config.

package main

import (
	"os"

	"github.com/kirill-scherba/teonet-go/services/teoconf"
)

// CdbParameters is teocdb application config data structure
type CdbParameters struct {
	Keyspace `json:"keyspace,omitempty"`
	Hosts    []string `json:"hosts,omitempty"`
}

// Keyspace is keyspaces of teocdb application config data structure
type Keyspace struct {
	Cdb   string `json:"cdb,omitempty"`
	Users string `json:"users,omitempty"`
	Room  string `json:"room,omitempty"`
}

// ConfHolder is teocdb application config data structure
type ConfHolder struct {
	*CdbParameters
	name string
}

// Config return teocdb config
func Config(name string) (conf *CdbParameters) {

	// Set default values
	val := &ConfHolder{
		CdbParameters: &CdbParameters{
			Keyspace: Keyspace{
				Cdb:   "teocdb",
				Users: "teousers",
				Room:  "teoroom",
			},
			Hosts: []string{"172.18.0.2", "172.18.0.3", "172.18.0.4"},
		},
		name: name,
	}

	// Create and Read config
	c := teoconf.New(val)

	// Return config data structure
	return c.Value().(*CdbParameters)
}

// KeyAndHosts combine keyspace and hosts
func (c *CdbParameters) KeyAndHosts(keyspace string) []string {
	return append([]string{keyspace}, c.Hosts...)
}

// Default return default value in json format.
func (c *ConfHolder) Default() []byte {
	return nil
}

// Value real value as interfaxe
func (c *ConfHolder) Value() interface{} {
	return c.CdbParameters
}

// Name return configuration file name.
func (c *ConfHolder) Name() string {
	return c.name
}

// Dir return configuration files folder
func (c *ConfHolder) Dir() string {
	return os.Getenv("HOME") + "/.config/teonet/teocdb/"
}

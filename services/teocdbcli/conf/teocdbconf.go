// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package conf is the Teonet config reader based on teocdb key-value
// databaseb service client package.
package conf

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teocdbcli"
	"github.com/kirill-scherba/teonet-go/services/teoconf"
)

var (
	// ErrConfigCdbDoesNotExists error returns by cdb config read function when
	// config does not exists in cdb
	ErrConfigCdbDoesNotExists = errors.New("config does not exists")
)

// Value is an conf interface
type Value interface {
	Key() string         // return teocdb configuration key
	Dir() string         // return configuration files folder
	Struct() interface{} // return struct value
	Default() []byte     // return default value in json
	FileName() string    // return configuration files name
}

// Teoconf is an conf receiver
type Teoconf struct {
	Value
	fconf *teoconf.Teoconf
	con   teocdbcli.TeoConnector
}

// New create and initialize teonet config
func New(con teocdbcli.TeoConnector, val Value) (c *Teoconf) {
	c = &Teoconf{Value: val, con: con}
	c.fconf = teoconf.New(val)
	c.ReadBoth()
	fmt.Println("config name:", c.Value.FileName())
	return
}

// setDefault sets default values from json string
func (c *Teoconf) setDefault() (err error) {
	data := c.Value.Default()
	if data == nil {
		return
	}
	// Unmarshal json to the value structure
	if err = json.Unmarshal(data, c.Value); err == nil {
		fmt.Printf("set default config value: %v\n", c.Value)
	}
	return
}

// ReadBoth sets default parameters, then read parameters from local config
// file, than read parameters from teo-cdb and save it to local file.
func (c *Teoconf) ReadBoth() (err error) {

	// Set defaults
	c.setDefault()

	// Read from local file
	if err := c.fconf.Read(); err != nil {
		fmt.Printf("read config error: %s\n", err)
	}

	// Read parameters from teo-cdb and applay it if changed, than write
	// it to local config file
	go func() {
		if err := c.ReadCdb(); err != nil {
			fmt.Printf("read cdb config error: %s\n", err)
			if err == ErrConfigCdbDoesNotExists {
				c.WriteCdb()
			}
			return
		}
		if err = c.fconf.Write(); err != nil {
			fmt.Printf("write config error: %s\n", err)
		}
	}()
	return
}

// fileName returns file name.
func (c *Teoconf) fileName() string {
	return c.Value.Dir() + c.Value.FileName() + ".json"
}

// ReadCdb read l0 parameters from config in teo-cdb.
func (c *Teoconf) ReadCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, c.Value.Key())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = ErrConfigCdbDoesNotExists
		return
	}

	fmt.Printf("config %s was read from teo-cdb: %s\n",
		c.Value.Key(), string(data))
	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, c.Value); err != nil {
		return
	}
	fmt.Println("config was read from teo-cdb: ", c.Value)
	return
}

// WriteCdb writes l0 parameters config to teo-cdb.
func (c *Teoconf) WriteCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Marshal json from the parameters structure
	data, err := json.Marshal(c.Value)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	if _, err = cdb.Send(teocdbcli.CmdSet, c.Value.Key(), data); err != nil {
		return
	}
	fmt.Println("config was write to teo-cdb: ", c.Value)
	return
}

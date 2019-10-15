// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoconf is the Teonet config reader based on teocdb key-value
// databaseb service client package.
package teoconf

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

// ConfLocation is an teoconf interface
type ConfLocation interface {
	ConfigDir() string    // return configuration files folder
	ConfigName() string   // return configuration files name
	ConfigKeyCdb() string // return teocdb configuration key
	// value interface{} // return config value
}

// Teoconf is an teoconf receiver
type Teoconf struct {
	loc ConfLocation
	val interface{}
	con teocdbcli.TeoConnector
}

// New create and initialize teonet config
func New(con teocdbcli.TeoConnector, loc ConfLocation, val interface{}) (c *Teoconf) {
	c = &Teoconf{con: con, loc: loc, val: val}
	c.ReadBoth()
	fmt.Println(c.loc.ConfigName())
	return
}

// ReadBoth sets default parameters, then read parameters from local config file,
// than read parameters from teo-cdb and save it to local file
func (c *Teoconf) ReadBoth() (err error) {

	// TODO: Set defaults

	// Read from local file
	if err := c.Read(); err != nil {
		fmt.Printf("Read config error: %s\n", err)
	}

	// Read parameters from teo-cdb and applay it if changed, than write
	// it to local config file
	go func() {
		if err := c.ReadCdb(); err != nil {
			fmt.Printf("Read cdb config error: %s\n", err)
			if err == ErrConfigCdbDoesNotExists {
				c.WriteCdb()
			}
		}
		if err = c.Write(); err != nil {
			fmt.Printf("Write config error: %s\n", err)
		}
	}()

	return
}

// Read reads l0 parameters from config file and replace current
// parameters
func (c *Teoconf) Read() (err error) {

	c.loc.ConfigDir()

	f, err := os.Open(c.loc.ConfigDir() + c.loc.ConfigName() + ".json")
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

	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, c.val); err == nil {
		fmt.Println("l0 parameters was read from file: ", c.val)
	}

	return
}

// Write writes game parameters to config file
func (c *Teoconf) Write() (err error) {
	f, err := os.Create(c.loc.ConfigDir() + c.loc.ConfigName() + ".json")
	if err != nil {
		return
	}
	// Marshal json from the parameters structure
	data, err := json.Marshal(c.val)
	if err != nil {
		return
	}
	_, err = f.Write(data)
	return
}

// ReadCdb read l0 parameters from config in teo-cdb
func (c *Teoconf) ReadCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, c.loc.ConfigKeyCdb())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = ErrConfigCdbDoesNotExists
		return
	}

	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, c.val); err != nil {
		return
	}
	fmt.Println("parameters was read from teo-cdb: ", c.val)
	return
}

// WriteCdb writes l0 parameters config to teo-cdb
func (c *Teoconf) WriteCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Marshal json from the parameters structure
	data, err := json.Marshal(c.val)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	_, err = cdb.Send(teocdbcli.CmdSet, c.loc.ConfigKeyCdb(), data)
	return
}

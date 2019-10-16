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

// ConfValue is an teoconf interface
type ConfValue interface {
	Key() string        // return teocdb configuration key
	Dir() string        // return configuration files folder
	Value() interface{} // return value
	Default() []byte    // return default value in json
	FileName() string   // return configuration files name
}

// Teoconf is an teoconf receiver
type Teoconf struct {
	ConfValue
	con teocdbcli.TeoConnector
}

// New create and initialize teonet config
func New(con teocdbcli.TeoConnector, val ConfValue) (c *Teoconf) {
	c = &Teoconf{ConfValue: val, con: con}
	c.ReadBoth()
	fmt.Println("config name:", c.ConfValue.FileName())
	return
}

// setDefault sets default values from json string
func (c *Teoconf) setDefault() (err error) {
	data := c.ConfValue.Default()
	if data == nil {
		return
	}
	// Unmarshal json to the value structure
	if err = json.Unmarshal(data, c.ConfValue); err == nil {
		fmt.Printf("set default config value: %v\n", c.ConfValue)
	}
	return
}

// ReadBoth sets default parameters, then read parameters from local config
// file, than read parameters from teo-cdb and save it to local file.
func (c *Teoconf) ReadBoth() (err error) {

	// Set defaults
	c.setDefault()

	// Read from local file
	if err := c.Read(); err != nil {
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
		if err = c.Write(); err != nil {
			fmt.Printf("write config error: %s\n", err)
		}
	}()
	return
}

// fileName returns file name.
func (c *Teoconf) fileName() string {
	return c.ConfValue.Dir() + c.ConfValue.FileName() + ".json"
}

// Read reads parameters from config file and replace current
// parameters.
func (c *Teoconf) Read() (err error) {
	f, err := os.Open(c.fileName())
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
	if err = json.Unmarshal(data, c.ConfValue); err == nil {
		fmt.Printf("config was read from file %s, value: %v\n",
			c.fileName(), c.ConfValue)
	}
	return
}

// Write writes game parameters to config file.
func (c *Teoconf) Write() (err error) {
	f, err := os.Create(c.fileName())
	if err != nil {
		return
	}
	// Marshal json from the parameters structure
	data, err := json.Marshal(c.ConfValue)
	if err != nil {
		return
	}
	if _, err = f.Write(data); err != nil {
		return
	}
	fmt.Printf("config was write to file %s, value: %v\n", c.fileName(),
		c.ConfValue)
	return
}

// ReadCdb read l0 parameters from config in teo-cdb.
func (c *Teoconf) ReadCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, c.ConfValue.Key())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = ErrConfigCdbDoesNotExists
		return
	}

	fmt.Printf("config %s was read from teo-cdb: %s\n",
		c.ConfValue.Key(), string(data))
	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, c.ConfValue); err != nil {
		return
	}
	fmt.Println("config was read from teo-cdb: ", c.ConfValue)
	return
}

// WriteCdb writes l0 parameters config to teo-cdb.
func (c *Teoconf) WriteCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Marshal json from the parameters structure
	data, err := json.Marshal(c.ConfValue)
	if err != nil {
		return
	}

	// Send config to teo-cdb
	if _, err = cdb.Send(teocdbcli.CmdSet, c.ConfValue.Key(), data); err != nil {
		return
	}
	fmt.Println("config was write to teo-cdb: ", c.ConfValue)
	return
}

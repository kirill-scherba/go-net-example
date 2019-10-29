// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package conf is the Teonet config reader based on teocdb key-value
// databaseb service client package.
package conf

import (
	"encoding/json"
	"errors"

	"github.com/kirill-scherba/teonet-go/services/teocdbcli"
	"github.com/kirill-scherba/teonet-go/services/teoconf"
	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// MODULE is this package module name
var MODULE = teokeys.Color(teokeys.ANSICyan, "(cdbconf)")

var (
	// ErrConfigCdbDoesNotExists error returns by cdb config read function when
	// config does not exists in cdb
	ErrConfigCdbDoesNotExists = errors.New("config does not exists")
)

// Config is an config interface
type Config interface {
	teoconf.Config
	// Key return teocdb configuration key
	Key() string
}

// Teoconf is an conf receiver
type Teoconf struct {
	Config
	fconf *teoconf.Teoconf
	con   teocdbcli.TeoConnector
}

// New create and initialize teonet config
func New(con teocdbcli.TeoConnector, val Config) (c *Teoconf) {
	c = &Teoconf{Config: val, con: con}
	c.fconf = teoconf.New(val)
	c.ReadBoth()
	teolog.Debug(MODULE, "config name:", c.Name())
	return
}

// setDefault sets default values from json string
func (c *Teoconf) setDefault() (err error) {
	data := c.Default()
	if data == nil {
		return
	}
	// Unmarshal json to the value structure
	if err = json.Unmarshal(data, c.Value()); err == nil {
		teolog.Debugf(MODULE, "set default config value: %v\n", c.Value())
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
		teolog.Debugf(MODULE, "read config error: %s\n", err)
	}

	// Read parameters from teo-cdb and applay it if changed, than write
	// it to local config file
	go func() {
		if err := c.ReadCdb(); err != nil {
			teolog.Debugf(MODULE, "read cdb config error: %s\n", err)
			if err == ErrConfigCdbDoesNotExists {
				c.WriteCdb()
			}
			return
		}
		if err = c.fconf.Write(); err != nil {
			teolog.Debugf(MODULE, "write config error: %s\n", err)
		}
	}()
	return
}

// ReadCdb read l0 parameters from config in teo-cdb.
func (c *Teoconf) ReadCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Get config from teo-cdb
	data, err := cdb.Send(teocdbcli.CmdGet, c.Key())
	if err != nil {
		return
	} else if data == nil || len(data) == 0 {
		err = ErrConfigCdbDoesNotExists
		return
	}

	teolog.Debugf(MODULE, "config %s was read from teo-cdb: %s\n",
		c.Key(), string(data))
	// Unmarshal json to the parameters structure
	if err = json.Unmarshal(data, c.Value()); err != nil {
		return
	}
	teolog.Debug(MODULE, "config was read from teo-cdb: ", c.Value())
	return
}

// WriteCdb writes l0 parameters config to teo-cdb.
func (c *Teoconf) WriteCdb() (err error) {

	// Create teocdb client
	cdb := teocdbcli.New(c.con)

	// Marshal json from the parameters structure
	data, err := json.Marshal(c.Value())
	if err != nil {
		return
	}

	// Send config to teo-cdb
	if _, err = cdb.Send(teocdbcli.CmdSet, c.Key(), data); err != nil {
		return
	}
	teolog.Debug(MODULE, "config was write to teo-cdb: ", c.Value())
	return
}

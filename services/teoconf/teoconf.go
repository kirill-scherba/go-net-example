// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoconf is the Teonet file in json format config reader.
package teoconf

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is an configration data interface
type Config interface {
	// Dir return configuration files folder
	Dir() string

	// Struct return pointer to input interface
	Struct() interface{}

	// Default return default value in json string or directly set default
	// values inside Struct with data
	Default() []byte

	// Return configuration files name or(and) key name when this packet used
	// inside teocdbcli/conf packet
	FileName() string
}

// Teoconf is an conf receiver
type Teoconf struct {
	Config
}

// New create and initialize teonet config
func New(val Config) (c *Teoconf) {
	c = &Teoconf{val}
	c.setDefault()
	c.Read()
	fmt.Println("config name:", c.Config.FileName())
	return
}

// setDefault sets default values from json string
func (c *Teoconf) setDefault() (err error) {
	data := c.Config.Default()
	if data == nil {
		return
	}
	// Unmarshal json to the value structure
	if err = json.Unmarshal(data, c.Config); err == nil {
		fmt.Printf("set default config value: %v\n", c.Config)
	}
	return
}

// fileName returns file name.
func (c *Teoconf) fileName() string {
	return c.Config.Dir() + c.Config.FileName() + ".json"
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
	if err = json.Unmarshal(data, c.Config); err == nil {
		fmt.Printf("config was read from file %s, value: %v\n",
			c.fileName(), c.Config)
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
	data, err := json.Marshal(c.Config)
	if err != nil {
		return
	}
	if _, err = f.Write(data); err != nil {
		return
	}
	fmt.Printf("config was write to file %s, value: %v\n", c.fileName(),
		c.Config)
	return
}

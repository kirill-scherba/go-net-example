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

// Value is an conf interface
type Value interface {
	Dir() string         // return configuration files folder
	Struct() interface{} // return struct value
	Default() []byte     // return default value in json
	FileName() string    // return configuration files name
}

// Teoconf is an conf receiver
type Teoconf struct {
	Value
}

// New create and initialize teonet config
func New(val Value) (c *Teoconf) {
	c = &Teoconf{Value: val}
	c.setDefault()
	c.Read()
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

// fileName returns file name.
func (c *Teoconf) fileName() string {
	return c.Value.Dir() + c.Value.FileName() + ".json"
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
	if err = json.Unmarshal(data, c.Value); err == nil {
		fmt.Printf("config was read from file %s, value: %v\n",
			c.fileName(), c.Value)
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
	data, err := json.Marshal(c.Value)
	if err != nil {
		return
	}
	if _, err = f.Write(data); err != nil {
		return
	}
	fmt.Printf("config was write to file %s, value: %v\n", c.fileName(),
		c.Value)
	return
}

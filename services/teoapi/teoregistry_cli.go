// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoapi is the Teonet registry service client package
package teoapi

import (
	"fmt"

	"github.com/gocql/gocql"
)

// Application is the Table 'applications': Teonet applications (services)
// description.
type Application struct {
	UUID    gocql.UUID
	Name    string
	Version string
	Descr   string
	Author  string
	License string
	Goget   string
	Git     string
	Com     []Command
}

// Command is the Table 'commands': Teonet applications commands description
// - cmdType values:  0 - input; 1 - input/output (same parameters); 2 - output
type Command struct {
	AppID       gocql.UUID
	Cmd         byte
	Type        uint8
	Descr       string
	TxtF        bool
	TxtNum      uint8
	TxtDescr    string
	JSONF       bool
	JSON        string
	BinaryF     bool
	BinaryDescr string
	Func        func(pac Packet) (err error)
	Message     func(pac Packet) (err error)
}

// Teoapi is api receiver.
type Teoapi struct {
	app *Application
	com []*Command
}

// Packet implement teonet packet interface.
type Packet interface {
	Cmd() byte
	From() string
	Data() []byte
	RemoveTrailingZero(data []byte) []byte
}

// New create new Teoregistrycli.
func New(app *Application) (api *Teoapi) {
	return &Teoapi{app: app}
}

// Add command description.
func (api *Teoapi) Add(com *Command) *Teoapi {
	api.com = append(api.com, com)
	return api
}

// Sprint print all added command to output string.
func (api *Teoapi) Sprint() (str string) {
	str = fmt.Sprintf("\b \nThe %s api commands:\n", api.app.Name)
	for i, c := range api.com {
		str += fmt.Sprintf("%2d. Command %d: %s\n", i+1, c.Cmd, c.Descr)
	}
	return
}

// Descr return command description.
func (api *Teoapi) Descr(cmd byte) (descr string) {
	for _, d := range api.com {
		if cmd == d.Cmd {
			descr = d.Descr
			return
		}
	}
	return
}

// Process packet commands.
func (api *Teoapi) Process(pac Packet) (err error) {
	for _, com := range api.com {
		if com.Cmd == pac.Cmd() {
			if com.Message != nil {
				err = com.Message(pac)
			}
			err = com.Func(pac)
			return
		}
	}
	return
}

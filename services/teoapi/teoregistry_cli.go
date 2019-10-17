// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoapi is the Teonet registry service client package
package teoapi

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoregistry"
)

// Application is Teonet applications description
type Application *teoregistry.Application

// Command is Teonet command description
type Command *teoregistry.Command

// Teoapi is api receiver
type Teoapi struct {
	app Application
	com []Command
}

// NewTeoapi create new Teoregistrycli
func NewTeoapi(app Application) (api *Teoapi) {
	return &Teoapi{app: app}
}

// Add command description
func (api *Teoapi) Add(c Command) *Teoapi {
	api.com = append(api.com, c)
	return api
}

// Sprint print all added command to output string
func (api *Teoapi) Sprint() (str string) {
	str = fmt.Sprintf("The %s api commands:\n", api.app.Name)
	for i, c := range api.com {
		str += fmt.Sprintf("%2d. Command %d: %s\n", i+1, c.Cmd, c.Descr)
	}
	return
}

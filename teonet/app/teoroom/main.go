// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet room controller (teo-room) micro service application.
//
// Teonet room controller combine users to room and send commands between it.
//
// Install this application:
//
//   go get github.com/kirill-scherba/teonet-go/teonet/app/teoroom/
//
// Run this application:
//   go run . teo-room
//
package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoapi"
	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// MODULE is this application module name
var MODULE = teokeys.Color(teokeys.ANSIMagenta, "(teoroom)")

func main() {

	// Version this teonet application version
	const Version = "0.1.0"

	// Teonet logo
	teonet.Logo("Teonet-go room conroller application", Version)

	// Applications teonet registy api description
	api := teoapi.New(&teoapi.Application{
		Name:    "teoroom",
		Version: Version,
		Descr:   "Teonet-go room controller service",
	}, 6)

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params(api)

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Teonet connect, init room controller package and run teonet
	teo := teonet.Connect(param, []string{"teo-go", "teo-room"}, Version, api)
	tr, err := teoroom.New(teo)
	if err != nil {
		panic(err)
	}
	defer tr.Destroy()

	// Teoapi command description
	//
	// Command #129: [in,out] Room request
	api.Add(&teoapi.Command{
		Cmd:   teoroomcli.ComRoomRequest,
		Descr: "Room request",
		Func: func(pac teoapi.Packet) (err error) {
			tpac := pac.(*teonet.Packet)
			if err = tr.Process.ComRoomRequest(tpac); err != nil {
				teolog.Debugf(MODULE, "%s\n", err.Error())
			}
			return
		},
	})

	// Command #130: [in,out] Data transfer
	api.Add(&teoapi.Command{
		Cmd:   teoroomcli.ComRoomData,
		Descr: "Data transfer",
		Func: func(pac teoapi.Packet) (err error) {
			tpac := pac.(*teonet.Packet)
			if err = tr.Process.ComRoomData(tpac); err != nil {
				teolog.Debugf(MODULE, "%s\n", err.Error())
			}
			return
		},
	})

	// Command #131: [in] Disconnect (exit) from room
	api.Add(&teoapi.Command{
		Cmd:   teoroomcli.ComDisconnect,
		Descr: "Disconnect (exit) from room",
		Func: func(pac teoapi.Packet) (err error) {
			if err = tr.Process.ComDisconnect(pac); err != nil {
				teolog.Debugf(MODULE, "Error Disconnect %s: %s\n", pac.From(),
					err.Error())
			}
			return
		},
	})

	// Teonet run
	teo.Run(func(teo *teonet.Teonet) {

		// Add teonet hotkey menu item to call termui interface
		teo.Menu().Add('m', "mui dashboard", func() {
			teo.SetLoglevel(teolog.NONE)
			fmt.Print("\b \b")
			go termui(teo, api)
		})

		// Teonet event loop
		for ev := range teo.Event() {

			// Event processing
			switch ev.Event {

			// When teonet started
			case teonet.EventStarted:
				fmt.Printf("Event Started\n")

			// When teonet peer connected
			case teonet.EventConnected:
				pac := ev.Data
				fmt.Printf("Event Connected from: %s\n", pac.From())

			// When teonet peer connected
			case teonet.EventDisconnected:
				pac := ev.Data
				fmt.Printf("Event Disconnected from: %s\n", pac.From())
			}
		}
	})
}

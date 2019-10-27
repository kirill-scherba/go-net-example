// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet cdb (teo-cdb) database service service application.
//
// Install this application:
//   go get github.com/kirill-scherba/teonet-go/teonet/app/teoroom/
//
// Before you execute this application, you need install database schemas.
// Launch `cqlsh` and execute next commands:
/*
  create keyspace teocdb with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };
  create table teocdb.map(key text, data blob, PRIMARY KEY(key));
*/
//
// Run this application:
//   go run . teo-cdb
//
package main

import (
	"fmt"
	"sync/atomic"

	"github.com/kirill-scherba/teonet-go/services/teoapi"
	"github.com/kirill-scherba/teonet-go/services/teocdb"
	"github.com/kirill-scherba/teonet-go/services/teocdbcli"
	teoroomStats "github.com/kirill-scherba/teonet-go/services/teoroom/stats"
	teoroomStatsCli "github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/services/teousers"
	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Version this teonet application version
const Version = "0.0.1"

// MODULE is this application module name
var MODULE = teokeys.Color(teokeys.ANSIMagenta, "(teocdb)")

func main() {

	// Teonet logo
	teonet.Logo("Teonet-go CQL Database service", Version)

	// Applications teonet registy api description
	api := teoapi.New(&teoapi.Application{
		Name:    "teocdb",
		Version: Version,
		Descr:   "Teonet-go CQL Database service",
	})

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params(api)

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Teonet connect
	teo := teonet.Connect(param, []string{"teo-go", "teo-cdb"}, Version, api)

	// Teonet cdb database
	tcdb, _ := teocdb.Connect(teo)
	defer tcdb.Close()

	// Teonet users database
	usr, _ := teousers.Connect(teo)
	defer usr.Close()

	// Teonet room controller statistic database
	room, _ := teoroomStats.Connect(teo)
	defer usr.Close()

	// Teoapi command description
	api.Add(&teoapi.Command{
		// # 129: Binary command execute all cammands Set, Get and GetList in
		// binary format
		Cmd:   teocdbcli.CmdBinary,
		Descr: "Binary set, get or get list",
		Func: func(pac teoapi.Packet) (err error) {
			err = tcdb.Process.CmdBinary(pac)
			if err != nil {
				fmt.Printf("CmdBinary Error: %s\n", err.Error())
			}
			return
		},
	}).Add(&teoapi.Command{
		// # 130: Set (insert or update) text or json {key,value} to database
		Cmd:   teocdbcli.CmdSet,
		Descr: "Set text or json {key,value} to key-value database",
		Func: func(pac teoapi.Packet) (err error) {
			err = tcdb.Process.CmdSet(pac)
			if err != nil {
				fmt.Printf("CmdSet Error: %s\n", err.Error())
			}
			return
		},
	}).Add(&teoapi.Command{
		// # 131: Get key value and send answer with value in text or json format
		Cmd:   teocdbcli.CmdGet,
		Descr: "Get key and send answer with value in text or json format",
		Func: func(pac teoapi.Packet) (err error) {
			err = tcdb.Process.CmdGet(pac)
			if err != nil {
				fmt.Printf("CmdGet Error: %s\n", err.Error())
			}
			return
		},
	}).Add(&teoapi.Command{
		// # 132: Get list of keys (by not complete key) and send answer with
		// array of keys in text or json format
		Cmd:   teocdbcli.CmdList,
		Descr: "List get not completed key and send answer with array of keys in text or json format",
		Func: func(pac teoapi.Packet) (err error) {
			err = tcdb.Process.CmdList(pac)
			if err != nil {
				fmt.Printf("CmdList Error: %s\n", err.Error())
			}
			return
		},
	}).Add(&teoapi.Command{
		// # 133: Check and register user
		Cmd:   teocdbcli.CheckUser,
		Descr: "Check and register user",
		Func: func(pac teoapi.Packet) (err error) {
			// Check access token
			res, err := usr.Process.ComCheckAccess(pac)
			if err == nil {
				teolog.Debugf(MODULE, "user Validated: %s, %s, %s\n",
					res.ID, res.AccessToken, res.Prefix)
				return
			}
			// Create user
			res, err = usr.Process.ComCreateUser(pac)
			if err != nil {
				fmt.Printf("ComCreateUser Error: %s\n", err.Error())
				return
			}
			teolog.Debugf(MODULE, "User Created: %s, %s, %s\n\n",
				res.ID, res.AccessToken, res.Prefix)
			return
		},
	}).Add(&teoapi.Command{
		// # 134: Room created
		Cmd:   teoroomStatsCli.CmdRoomCreated,
		Descr: "Room created",
		Func: func(pac teoapi.Packet) (err error) {
			room.ComRoomCreated(pac)
			return
		},
	}).Add(&teoapi.Command{
		// # 135: Room state changed
		Cmd:   teoroomStatsCli.CmdRoomState,
		Descr: "Room state changed",
		Func: func(pac teoapi.Packet) (err error) {
			room.ComRoomStateChanged(pac)
			return
		},
	}).Add(&teoapi.Command{
		// # 136: Client status changed
		Cmd:   teoroomStatsCli.CmdClientState,
		Descr: "Client state changed",
		Func: func(pac teoapi.Packet) (err error) {
			room.ComClientStatus(pac)
			return
		},
	})

	// Commands processing workers pool
	const numWorkers = 6
	workerRun := make([]uint64, numWorkers)
	commandChan := make(chan teoapi.Packet, numWorkers*4)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for {
				pac, ok := <-commandChan
				if !ok {
					return
				}
				//workerRun[workerID]++
				atomic.AddUint64(&workerRun[workerID], 1)
				teolog.Debugf(MODULE, "worker #%d got cmd %d: '%s', from: %s",
					workerID, pac.Cmd(), api.Descr(pac.Cmd()), pac.From())
				api.Process(pac)
			}
		}(i)
	}

	// Add teonet hotkey menu item to call termui interface
	teo.Menu().Add('m', "mui dashboard", func() {
		teo.SetLoglevel(teolog.NONE)
		fmt.Print("\b \b")
		go termui(api, workerRun)
	})

	// Teonet run
	teo.Run(func(teo *teonet.Teonet) {
		//fmt.Println("Teonet even loop started")
		for ev := range teo.Event() {

			// Event processing
			switch ev.Event {

			// When teonet started
			case teonet.EventStarted:
				fmt.Printf("Event Started\n")
			// case teonet.EventStoppedBefore:
			// case teonet.EventStopped:
			// 	fmt.Printf("Event Stopped\n")

			// When teonet peer connected
			case teonet.EventConnected:
				pac := ev.Data
				fmt.Printf("Event Connected from: %s\n", pac.From())

			// When teonet peer connected
			case teonet.EventDisconnected:
				pac := ev.Data
				fmt.Printf("Event Disconnected from: %s\n", pac.From())

			// When received command from teonet peer or client
			case teonet.EventReceived:
				pac := ev.Data
				commandChan <- pac
			}
		}
	})
}

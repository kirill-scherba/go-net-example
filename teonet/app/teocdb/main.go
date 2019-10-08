// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet cdb (teo-cdb) database service application.
//
// Before you execute this application, you need install database schemas.
// Launch `cqlsh` and execute next commands:
/*
  create keyspace teocdb with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };
  create table teocdb.map(key text, data blob, PRIMARY KEY(key));
*/
package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teocdb"
	cdb "github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
	"github.com/kirill-scherba/teonet-go/services/teoregistry"
	"github.com/kirill-scherba/teonet-go/services/teoregistry/teoapi"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Version this teonet application version
const Version = "0.0.1"

func main() {

	// Teonet logo
	teonet.Logo("Teonet-go CQL Database service", Version)

	// Applications teonet registy api description
	api := teoapi.NewTeoapi(&teoregistry.Application{
		Name:    "teocdb",
		Version: Version,
		Descr:   "Teonet-go CQL Database service",
	}).Add(&teoregistry.Command{
		Cmd: 129, Descr: "Binary set, get or get list binary {key,value} to/from key-value database",
	}).Add(&teoregistry.Command{
		Cmd: 130, Descr: "Set (insert or update) text or json {key,value} to key-value database",
	}).Add(&teoregistry.Command{
		Cmd: 131, Descr: "Get key and send answer with value in text or json format from key-value database",
	}).Add(&teoregistry.Command{
		Cmd: 132, Descr: "List get not completed key and send answer with array of keys in text or json format from key-value database",
	})

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params(api)

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Teonet connect
	teo := teonet.Connect(param, []string{"teo-go", "teo-cdb"}, Version, api)

	// Connect to the cql cluster
	// cluster := gocql.NewCluster("172.17.0.2", "172.17.0.3", "172.17.0.4")
	// cluster.Keyspace = "teocdb"
	// cluster.Consistency = gocql.Quorum
	// session, _ := cluster.CreateSession()
	// defer session.Close()
	tcdb, _ := teocdb.Connect(teo)
	defer tcdb.Close()

	// Commands processing
	commands := func(pac *teonet.Packet) {
		switch pac.Cmd() {

		// # 129: Binary command execute all cammands Set, Get and GetList in binary format
		case cdb.CmdBinary:
			err := tcdb.Process().CmdBinary(pac)
			if err != nil {
				fmt.Printf("CmdBinary Error: %s\n", err.Error())
			}

		// # 130: Set (insert or update) text or json {key,value} to database
		case cdb.CmdSet:
			err := tcdb.Process().CmdSet(pac)
			if err != nil {
				fmt.Printf("CmdSet Error: %s\n", err.Error())
			}

		// # 131: Get key value and send answer with value in text or json format
		case cdb.CmdGet:
			err := tcdb.Process().CmdGet(pac)
			if err != nil {
				fmt.Printf("CmdGet Error: %s\n", err.Error())
			}

		// # 132: Get list of keys (by not complete key) and send answer with array of keys in text or json format
		case cdb.CmdList:
			err := tcdb.Process().CmdList(pac)
			if err != nil {
				fmt.Printf("CmdList Error: %s\n", err.Error())
			}
		}
	}

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
				fmt.Printf("Event Received from: %s, cmd: %d, data: %s\n",
					pac.From(), pac.Cmd(), pac.Data())

				// Commands processing
				go commands(pac)
			}
		}
	})
}

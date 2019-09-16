// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet cql(cassandra database) application
//
/* Before you execute the program, Launch `cqlsh` and execute:
create keyspace teocdb with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table teocdb.map(key text, data blob, PRIMARY KEY(key));
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/app/teocdb/teocdb"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Version this teonet application version
const Version = "0.0.1"

func main() {

	// Teonet logo
	teonet.Logo("Teonet-go CQL Database service", Version)

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params()

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Connect to the cql cluster
	cluster := gocql.NewCluster("172.17.0.2")
	cluster.Keyspace = "teocdb"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	// Teonet connect and run
	teo := teonet.Connect(param, []string{"teo-go"}, Version)
	teo.Run(func(teo *teonet.Teonet) {
		//fmt.Println("Teonet even loop started")
		for ev := range teo.Event() {
			switch ev.Event {
			case teonet.EventStarted:
				fmt.Printf("Event Started\n")
			// case teonet.EventStoppedBefore:
			// case teonet.EventStopped:
			// 	fmt.Printf("Event Stopped\n")
			case teonet.EventConnected:
				pac := ev.Data
				fmt.Printf("Event Connected from: %s\n", pac.From())
			case teonet.EventDisconnected:
				pac := ev.Data
				fmt.Printf("Event Disconnected from: %s\n", pac.From())
			case teonet.EventReceived:
				pac := ev.Data
				fmt.Printf("Event Received from: %s, cmd: %d, data: %s\n",
					pac.From(), pac.Cmd(), pac.Data())

				// Commands processing
				switch pac.Cmd() {

				// Insert binary {key,value} to database
				case 129:
					key, value := teocdb.Unmarshal(pac.Data())
					fmt.Println(key, value)
					if err := session.Query(`INSERT INTO map (key, data) VALUES (?, ?)`,
						key, value).Exec(); err != nil {
						fmt.Printf("Insert Error: %s\n", err.Error())
					}

				// Insert text {key,value} to database
				case 130:
					d := strings.Split(string(pac.Data()), ",")
					if len(d) >= 2 {
						fmt.Println(d[0], d[1])
						if err := session.Query(`INSERT INTO map (key, data) VALUES (?, ?)`,
							d[0], []byte(d[1])).Exec(); err != nil {
							fmt.Printf("Insert Error: %s\n", err.Error())
						}
					}

				// Insert json {key,value} to database
				case 131:
					var v struct {
						Key   string      `json:"key"`
						Value interface{} `json:"value"`
					}
					json.Unmarshal(pac.Data(), &v)
					fmt.Println(v.Key, v.Value)
					if err := session.Query(`INSERT INTO map (key, data) VALUES (?, ?)`,
						v.Key, v.Value).Exec(); err != nil {
						fmt.Printf("Insert Error: %s\n", err.Error())
					}

				// Read data and send answer with it
				case 132:
					var data []byte
					if err := session.Query(`SELECT data FROM map WHERE key = ? LIMIT 1`,
						string(pac.Data())).Consistency(gocql.One).Scan(&data); err == nil {
						fmt.Printf("Got from db: %v\n", data)
						teo.SendTo(pac.From(), pac.Cmd(), data)
					}
				}
			}
		}
		//fmt.Println("Teonet even loop stopped")
	})
}

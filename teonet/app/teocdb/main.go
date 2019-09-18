// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet cdb(cassandra database) application
//
/* Before you execute the program, Launch `cqlsh` and execute:
create keyspace teocdb with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };
create table teocdb.map(key text, data blob, PRIMARY KEY(key));
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kirill-scherba/teonet-go/services/teocdb"
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
	// cluster := gocql.NewCluster("172.17.0.2", "172.17.0.3", "172.17.0.4")
	// cluster.Keyspace = "teocdb"
	// cluster.Consistency = gocql.Quorum
	// session, _ := cluster.CreateSession()
	// defer session.Close()
	tdb, _ := teocdb.Connect()
	defer tdb.Close()

	// Teonet connect and run
	teo := teonet.Connect(param, []string{"teo-go", "teo-cdb"}, Version)
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

				// updateKeyValue Parse input parameters and update key value in database
				updateKeyValue := func(data []byte) (key string, value []byte, err error) {
					if teonet.DataIsJSON(data) {
						var v teocdb.JData
						json.Unmarshal(data, &v)
						key = v.Key
						value, _ = json.Marshal(v.Value)
					} else {
						d := strings.Split(string(data), ",")
						if len(d) < 2 {
							err = errors.New("not enough parameters in text request")
							return
						}
						key = d[0]
						value = []byte(d[1])
					}
					err = tdb.Update(key, value)
					return
				}

				// readKeyValue Parse input parameters and read key value
				readKeyValue := func(req []byte) (data []byte, jsonReqF bool, err error) {
					var jsonData teocdb.JData

					// Unmarshal request
					if jsonReqF = teonet.DataIsJSON(req); !jsonReqF {
						jsonData.Key = string(req)
					} else if err = json.Unmarshal(req, &jsonData); err != nil {
						return
					}

					// Get result from database
					if data, err = tdb.Get(jsonData.Key); err != nil {
						return
					}
					fmt.Printf("Got from db: %v\n", data)

					// Marshal responce
					if jsonReqF {
						if err := json.Unmarshal(data, &jsonData.Value); err != nil {
							jsonData.Value = string(data)
						}
						data, err = json.Marshal(jsonData)
					}
					return
				}

				// readKeyList Parse input parameters and read list of keys
				readKeyList := func(req []byte) (data []byte, jsonReqF bool, err error) {
					var jsonData teocdb.JData

					// Unmarshal request
					if jsonReqF = teonet.DataIsJSON(req); !jsonReqF {
						jsonData.Key = string(req)
					} else if err = json.Unmarshal(req, &jsonData); err != nil {
						return
					}

					// Read list
					jdata, err := tdb.List(jsonData.Key)
					if err != nil {
						return
					}
					fmt.Printf("Got from db: %v\n", jdata)
					sort.Strings(jdata)
					data, _ = json.Marshal(jdata)

					// Marshal responce
					if jsonReqF {
						if err = json.Unmarshal(data, &jsonData.Value); err != nil {
							return
						}
						data, err = json.Marshal(jsonData)
					}
					return
				}

				// Commands processing
				switch pac.Cmd() {

				// Insert(or Update) binary {key,value} to database
				case 129:
					key, value := teocdb.Unmarshal(pac.Data())
					fmt.Println(key, value)
					if err := tdb.Update(key, value); err != nil {
						fmt.Printf("Insert Error: %s\n", err.Error())
					}

				// Insert(or Update) text or json "key,value" to database
				case 130:
					key, value, err := updateKeyValue(pac.Data())
					if err != nil {
						fmt.Printf("Insert(or Update) Error: %s\n", err.Error())
						break
					}
					fmt.Println(key, value)

				// Read key data and send answer with data in text or json format
				case 131:
					data, _, err := readKeyValue(pac.Data())
					if err != nil {
						fmt.Printf("Get Error: %s\n", err.Error())
						break
					}
					teo.SendTo(pac.From(), pac.Cmd(), data)

				// Read list of keys and send answer with list array in text or json format
				case 132:
					data, _, err := readKeyList(pac.Data())
					if err != nil {
						fmt.Printf("Read List Error: %s\n", err.Error())
						break
					}
					teo.SendTo(pac.From(), pac.Cmd(), data)
				}
			}
		}
		//fmt.Println("Teonet even loop stopped")
	})
}

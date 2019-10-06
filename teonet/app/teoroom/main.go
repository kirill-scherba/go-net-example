// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet room controller (teo-room) application
//
// Teonet room controller combine users to room and send commands between it.
//
// Install this application:
//   go get github.com/kirill-scherba/teonet-go/teonet/app/teoroom/
//
// Run this application:
//   go run .
//
package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/services/teoroom/teoroomcli"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func main() {

	// Version this teonet application version
	const Version = "0.0.1"

	// Teonet logo
	teonet.Logo("Teonet-go room conroller service", Version)

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params()

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Teonet connect, init room controller package and run teonet
	teo := teonet.Connect(param, []string{"teo-go", "teo-room"}, Version)
	tr, err := teoroom.New(teo)
	if err != nil {
		panic(err)
	}
	defer tr.Destroy()
	teo.Run(func(teo *teonet.Teonet) {
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
				fmt.Printf("Event Received from: %s, cmd: %d, data: %v\n",
					pac.From(), pac.Cmd(), pac.Data())

				// Commands processing
				switch pac.Cmd() {

				// Command #129: [in,out] Room request
				case teoroomcli.ComRoomRequest:
					_, clientID, err := tr.RoomRequest(pac.From())
					if err != nil {
						fmt.Printf("%s\n", err.Error())
						break
					}
					teo.SendToClient("teo-l0", pac.From(),
						teoroomcli.ComRoomRequestAnswer,
						append([]byte{}, byte(clientID)))

				// Command #130: [in,out] Data transfer
				case teoroomcli.ComRoomData:
					tr.ResendData(pac.From(), pac.Data(), func(l0, client string,
						data []byte) {
						teo.SendToClient("teo-l0", client, teoroomcli.ComRoomData,
							data)
					})

				// Command #131 [in] Disconnect (exit) from room
				case teoroomcli.ComDisconnect:
					tr.ResendData(pac.From(), pac.Data(), func(l0, client string,
						data []byte) {
						teo.SendToClient("teo-l0", client, teoroomcli.ComDisconnect,
							data)
					})
					if err := tr.Disconnect(pac.From()); err != nil {
						fmt.Printf("Client %s is already disconnected\n",
							pac.From())
						break
					}
				}
			}
		}
	})
}

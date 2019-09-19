// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet room controller application.
// Teo room unites users to room and send commands between it

package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Room controller commands
const (
	// ComRoomRequest [in,out] #129 Room request or Room request answer
	// [in]  Room request
	// [out] Room request anwser
	ComRoomRequest = 129

	// ComRoomData [in,out] #130 Data transfer
	ComRoomData = 130
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

	// Start room controller
	tr, err := teoroom.Init()
	if err != nil {
		panic(err)
	}
	defer tr.Destroy()

	// Teonet connect and run
	teo := teonet.Connect(param, []string{"teo-go", "teo-room"}, Version)
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
				fmt.Printf("Event Received from: %s, cmd: %d, data: %s\n",
					pac.From(), pac.Cmd(), pac.Data())

				// Commands processing
				switch pac.Cmd() {

				// Command #129: [in,out] Room request
				case 129:
					teo.SendToClient("teo-l0", pac.From(), pac.Cmd(), pac.Data())

				// Command #130: [in,out] Data transfer
				case 130:
				}
			}
		}
	})
}

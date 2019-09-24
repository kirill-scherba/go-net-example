// Main teonet sample application
//
package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Version this teonet application version
const Version = "0.0.1"

func main() {

	// Teonet logo
	teonet.Logo("Teonet-go test application", Version)

	// Read Teonet parameters from configuration file and parse application
	// flars and arguments
	param := teonet.Params()

	// Show host and network name
	fmt.Printf("\nhost: %s\nnetwork: %s\n", param.Name, param.Network)

	// Teonet connect and run
	teo := teonet.Connect(param, []string{"teo-go"}, Version)
	teo.Run(func(teo *teonet.Teonet) {
		fmt.Println("Teonet even loop started")
		for ev := range teo.Event() {
			switch ev.Event {
			case teonet.EventStarted:
				fmt.Printf("Event Started\n")
			case teonet.EventStoppedBefore:
			case teonet.EventStopped:
				fmt.Printf("Event Stopped\n")
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
				switch pac.Cmd() {
				case 129:
					teo.SendTo(pac.From(), pac.Cmd()+1, pac.Data())
				}
			default:
				fmt.Printf("event: %d\n", ev.Event)
			}
		}
		fmt.Println("Teonet even loop stopped")
	})
}

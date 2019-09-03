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
	teo := teonet.Connect(param)    // connect to teonet
	teo.SetType([]string{"teo-go"}) // set this teonet application type
	teo.SetVersion(Version)         // set this teonet application version
	teo.CtrlC()                     // set process Ctrl+C to exit
	teo.Run()                       // run teonet
}

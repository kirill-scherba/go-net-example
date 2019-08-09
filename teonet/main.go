package main

import (
	"flag"
	"fmt"

	"github.com/kirill-scherba/net-example-go/teonet/teonet"
)

func main() {

	// Teonet logo
	fmt.Println("" +
		" _____                     _   \n" +
		"|_   _|__  ___  _ __   ___| |_ \n" +
		"  | |/ _ \\/ _ \\| '_ \\ / _ \\ __|\n" +
		"  | |  __/ (_) | | | |  __/ |_ \n" +
		"  |_|\\___|\\___/|_| |_|\\___|\\__|\n" +
		"\n" +
		"Teonet test application ver " +
		teonet.Version +
		"\n",
	)

	// Parameters variables
	var name string      // this client name
	var port int         // local port
	var raddr string     // remote host address
	var rport, rchan int // remote host port and channel (for TRUDP)

	// Parameters
	flag.IntVar(&port, "p", 0, "local host port")
	flag.StringVar(&name, "n", "teonet-go-01", "local host teonet name")
	flag.StringVar(&raddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 0, "remote host channel (to connect to remote host TRUDP channel)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.Parse()

	// Teonet connect and run
	teo := teonet.Connect(name, port, raddr, rport)
	teo.Run()
}

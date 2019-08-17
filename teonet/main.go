package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kirill-scherba/net-example-go/teonet/teonet"
)

// Version is this application version
const Version = "0.0.1"

func main() {

	// Teonet logo
	fmt.Println("" +
		" _____                     _   \n" +
		"|_   _|__  ___  _ __   ___| |_ \n" +
		"  | |/ _ \\/ _ \\| '_ \\ / _ \\ __|\n" +
		"  | |  __/ (_) | | | |  __/ |_ \n" +
		"  |_|\\___|\\___/|_| |_|\\___|\\__|\n" +
		"\n" +
		"Teonet-go test application ver " + Version +
		", based on teonet ver " + teonet.Version +
		"\n",
	)

	// Teonet parameters
	param := new(teonet.Parameters)
	// This application Usage message
	flag.Usage = func() {
		fmt.Printf("usage: %s [OPTIONS] host_name\n\noptions:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	// Teonet flags
	flag.IntVar(&param.Port, "p", 0, "local host port")
	flag.StringVar(&param.Network, "n", "local", "teonet network name")
	flag.StringVar(&param.RAddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&param.RChan, "c", 0, "remote host channel (to connect to remote host TRUDP channel)")
	flag.IntVar(&param.RPort, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&param.LogLevel, "log-level", "DEBUG", "show log messages level")
	flag.StringVar(&param.LogFilter, "log-filter", "", "set log messages filter")
	flag.BoolVar(&param.ShowTrudpStatF, "show-trudp", false, "show trudp statistic")
	flag.BoolVar(&param.ForbidHotkeysF, "forbid-hotkeys", false, "forbid hotkeys")
	flag.BoolVar(&param.ShowPeersStatF, "show-peers", false, "show peers table")
	flag.BoolVar(&param.ShowHelpF, "h", false, "show this help message")
	flag.Parse()
	// Teonet Arguments
	args := flag.Args()
	if param.ShowHelpF || len(args) == 0 {
		if len(args) == 0 {
			fmt.Printf("argument host_name not defined\n")
		}
		flag.Usage()
		os.Exit(0)
	}
	param.Name = flag.Arg(0)
	fmt.Printf("host name: %s\nnetwork: %s\n", param.Network, param.Name)

	// Teonet connect and run
	teo := teonet.Connect(param)
	teo.SetType([]string{"teo-go"})
	teo.CtrlC()
	teo.Run()
}

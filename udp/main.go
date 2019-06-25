package main

import (
	"flag"
	"fmt"

	"github.com/kirill-scherba/go-net-example/udp/trudp"
)

func main() {
	fmt.Println("UDP test application ver 1.0.0")

	var (
		rhost     string
		rport     int
		port      int
		log       string
		noLogTime bool
	)

	flag.BoolVar(&noLogTime, "no_log_time", false, "don't show time in application log")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&log, "log", "DEBUG_V", "application log level")
	flag.Parse()

	conn := trudp.Init(port)
	conn.LogLevel(log, !noLogTime)
	if rport != 0 {
		conn.ConnectChannel(rhost, rport, 1)
	}
	conn.Run()
}

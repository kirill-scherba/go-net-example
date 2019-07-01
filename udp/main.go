package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kirill-scherba/net-example-go/udp/trudp"
)

func main() {
	fmt.Println("UDP test application ver 1.0.0")

	var (
		rhost     string
		rport     int
		port      int
		logLevel  string
		noLogTime bool
	)

	flag.BoolVar(&noLogTime, "no_log_time", false, "don't show time in application log")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&logLevel, "log", "DEBUGv", "application log level")
	flag.Parse()

	conn := trudp.Init(port)
	conn.LogLevel(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds)
	if rport != 0 {
		tcd := conn.ConnectChannel(rhost, rport, 1)
		tcd.SendTestMsg(true)
	}
	go func() {
		for ev := range conn.Event {
			switch ev.Event {
			case trudp.GOT_DATA:
				log.Println("(main) got data:", ev.Data, string(ev.Data))
			}
		}
	}()
	conn.Run()
}

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

	tru := trudp.Init(port)
	tru.LogLevel(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds)
	if rport != 0 {
		tcd := tru.ConnectChannel(rhost, rport, 1)
		tcd.SendTestMsg(true)
	}
	go func() {
		for ev := range tru.Event {
			switch ev.Event {

			case trudp.GOT_DATA:
				log.Println("(main) GOT_DATA: ", ev.Data, string(ev.Data))

			case trudp.SEND_DATA:
				log.Println("(main) SEND_DATA:", ev.Data, string(ev.Data))

			case trudp.INITIALIZE:
				log.Println("(main) INITIALIZE, listen at:", string(ev.Data))

			case trudp.CONNECTED:
				log.Println("(main) CONNECTED", string(ev.Data))

			case trudp.DISCONNECTED:
				log.Println("(main) DISCONNECTED", string(ev.Data))

			case trudp.RESET_LOCAL:
				log.Println("(main) RESET_LOCAL executed")
			}
		}
	}()
	tru.Run()
}

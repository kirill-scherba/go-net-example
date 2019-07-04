package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kirill-scherba/net-example-go/udp/trudp"
)

func main() {
	fmt.Println("UDP test application ver 1.0.0")

	// trudp.BenchmarkSyscallUDP()
	// return

	var (
		rhost     string
		rport     int
		rchan     int
		port      int
		logLevel  string
		noLogTime bool
		sendTest  bool
	)

	flag.BoolVar(&noLogTime, "no_log_time", false, "don't show time in application log")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 1, "remote host channel (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&logLevel, "log", "DEBUGv", "application log level")
	flag.BoolVar(&sendTest, "send_test", false, "send test data")
	flag.Parse()

	tru := trudp.Init(port)
	tru.LogLevel(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds)
	if rport != 0 {
		tcd := tru.ConnectChannel(rhost, rport, rchan)
		// Auto sender
		if sendTest {
			tcd.SendTestMsg(true)
		}
		// Sender
		num := 0
		f := func() {
			defer func() { log.Println("(main) channels sender stopped") }()
			const sleepTime = 250
			for {
				time.Sleep(sleepTime * time.Microsecond)
				// if num%100 == 0 { // 100
				// 	time.Sleep(500 * time.Microsecond) // 500
				// }
				data := []byte("Hello-" + strconv.Itoa(num) + "!")
				err := tcd.WriteTo(data)
				if err != nil {
					return
				}
				num++
			}
		}
		for i := 0; i < 1; i++ {
			go f()
		}
	}
	// Receiver
	go func() {
		for ev := range tru.Event {
			switch ev.Event {

			case trudp.GOT_DATA:
				log.Println("(main) GOT_DATA: ", ev.Data, string(ev.Data), fmt.Sprintf("%.3f ms", ev.Tcd.TripTime()))
				if rport == 0 {
					ev.Tcd.WriteTo([]byte(string(ev.Data) + " - answer"))
				}

			case trudp.SEND_DATA:
				log.Println("(main) SEND_DATA:", ev.Data, string(ev.Data))

			case trudp.INITIALIZE:
				log.Println("(main) INITIALIZE, listen at:", string(ev.Data))

			case trudp.CONNECTED:
				log.Println("(main) CONNECTED", string(ev.Data))

			case trudp.DISCONNECTED:
				log.Println("(main) DISCONNECTED", string(ev.Data))

			case trudp.RESET_LOCAL:
				log.Println("(main) RESET_LOCAL executed at channel:", ev.Tcd.MakeKey())

			case trudp.SEND_RESET:
				log.Println("(main) SEND_RESET to channel:", ev.Tcd.MakeKey())
			}
		}
	}()
	tru.Run()
}

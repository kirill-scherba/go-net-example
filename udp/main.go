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

	var (
		rhost        string
		rport        int
		rchan        int
		port         int
		logLevel     string
		maxQueueSize int

		// Control flags
		noLogTime bool
		sendTest  bool
		showStat  bool
	)

	flag.IntVar(&maxQueueSize, "Q", trudp.DefaultQueueSize, "maximum send and receive queues size")
	flag.BoolVar(&noLogTime, "no_log_time", false, "don't show time in application log")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 1, "remote host channel (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&logLevel, "log", "CONNECT", "application log level")
	flag.BoolVar(&sendTest, "send_test", false, "send test data")
	flag.BoolVar(&showStat, "S", false, "show statistic")
	flag.Parse()

	tru := trudp.Init(port)
	tru.LogLevel(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds)

	// Show statictic flag
	if showStat {
		tru.ShowStatistic(true)
	}

	// Set default queue size
	tru.SetDefaultQueueSize(maxQueueSize)

	// Connect to remote server flag and send data when connected
	if rport != 0 {
		go func() {
			// Try to connect to remote hosr every 5 seconds
			for {
				tcd := tru.ConnectChannel(rhost, rport, rchan)

				// Auto sender flag
				if sendTest {
					tcd.SendTestMsg(true)
				}

				// Sender
				func() {
					num := 0
					const sleepTime = 0
					for {
						if sleepTime > 0 {
							time.Sleep(sleepTime * time.Microsecond)
						}
						data := []byte("Hello-" + strconv.Itoa(num) + "!")
						err := tcd.WriteTo(data)
						if err != nil {
							break
						}
						num++
					}
					tru.Log(trudp.CONNECT, "(main) channels sender stopped")
				}()

				time.Sleep(5 * time.Second)
				tru.Log(trudp.CONNECT, "(main) reconnect")
			}
		}()
	}

	// Receiver
	go func() {
		for ev := range tru.ChRead() {
			switch ev.Event {

			case trudp.GOT_DATA:
				tru.Log(trudp.DEBUG, "(main) GOT_DATA: ", ev.Data, string(ev.Data), fmt.Sprintf("%.3f ms", ev.Tcd.TripTime()))
				// Send answer if this host not connected to remote hosr
				if rport == 0 {
					ev.Tcd.WriteTo([]byte(string(ev.Data) + " - answer"))
				}

			case trudp.SEND_DATA:
				tru.Log(trudp.DEBUG, "(main) SEND_DATA:", ev.Data, string(ev.Data))

			case trudp.INITIALIZE:
				tru.Log(trudp.DEBUG, "(main) INITIALIZE, listen at:", string(ev.Data))

			case trudp.CONNECTED:
				tru.Log(trudp.CONNECT, "(main) CONNECTED", string(ev.Data))

			case trudp.DISCONNECTED:
				tru.Log(trudp.CONNECT, "(main) DISCONNECTED", string(ev.Data))

			case trudp.RESET_LOCAL:
				tru.Log(trudp.CONNECT, "(main) RESET_LOCAL executed at channel:", ev.Tcd.MakeKey())

			case trudp.SEND_RESET:
				tru.Log(trudp.CONNECT, "(main) SEND_RESET to channel:", ev.Tcd.MakeKey())
			}
		}
	}()

	// Run trudp and start listen
	tru.Run()
}

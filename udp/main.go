package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kirill-scherba/net-example-go/udp/trudp"
)

func main() {
	fmt.Println("UDP test application ver 1.0.0")

	var (
		rhost string
		rport int
		rchan int
		port  int

		// Logger parameters
		logLevel string

		// Integer parameters
		maxQueueSize  int
		sendSleepTime int

		// Control flags parameters
		noLogTime  bool
		sendTest   bool
		showStat   bool
		sendAnswer bool
	)

	flag.IntVar(&maxQueueSize, "Q", trudp.DefaultQueueSize, "maximum send and receive queues size")
	flag.BoolVar(&noLogTime, "no-log-time", false, "don't show time in application log")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 1, "remote host channel (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&logLevel, "log", "CONNECT", "application log level")
	flag.IntVar(&sendSleepTime, "t", 0, "send timeout in microseconds")
	flag.BoolVar(&sendTest, "send-test", false, "send test data")
	flag.BoolVar(&sendAnswer, "answer", false, "send answer")
	flag.BoolVar(&showStat, "S", false, "show statistic")

	flag.Parse()

	tru := trudp.Init(port)

	// Set log level
	tru.LogLevel(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds)

	// Set 'show statictic' flag
	tru.ShowStatistic(showStat)

	// Set default queue size
	tru.SetDefaultQueueSize(maxQueueSize)

	// Connect to remote server flag and send data when connected
	if rport != 0 {
		go func() {
			// Try to connect to remote hosr every 5 seconds
			for {
				tcd := tru.ConnectChannel(rhost, rport, rchan)

				// Auto sender flag
				tcd.SendTestMsg(sendTest)

				// Sender
				num := 0
				for {
					if sendSleepTime > 0 {
						time.Sleep(time.Duration(sendSleepTime) * time.Microsecond)
					}
					data := []byte("Hello-" + strconv.Itoa(num) + "!")
					err := tcd.WriteTo(data)
					if err != nil {
						break
					}
					num++
				}
				tru.Log(trudp.CONNECT, "(main) channels sender stopped")

				time.Sleep(5 * time.Second)
				tru.Log(trudp.CONNECT, "(main) reconnect")
			}
		}()
	}

	// Receiver
	go func() {
		for ev := range tru.ChanEvent() {
			//go func(ev *trudp.EventData) {
			switch ev.Event {

			case trudp.GOT_DATA:
				tru.Log(trudp.DEBUG, "(main) GOT_DATA: ", ev.Data, string(ev.Data), fmt.Sprintf("%.3f ms", ev.Tcd.TripTime()))
				// Send answer
				if sendAnswer {
					ev.Tcd.WriteTo([]byte(string(ev.Data) + " - answer"))
				}

			case trudp.SEND_DATA:
				tru.Log(trudp.DEBUG, "(main) SEND_DATA:", ev.Data, string(ev.Data))

			case trudp.INITIALIZE:
				tru.Log(trudp.ERROR, "(main) INITIALIZE, listen at:", string(ev.Data))

			case trudp.DESTROY:
				tru.Log(trudp.ERROR, "(main) DESTROY", string(ev.Data))

			case trudp.CONNECTED:
				tru.Log(trudp.CONNECT, "(main) CONNECTED", string(ev.Data))

			case trudp.DISCONNECTED:
				tru.Log(trudp.CONNECT, "(main) DISCONNECTED", string(ev.Data))

			case trudp.RESET_LOCAL:
				tru.Log(trudp.CONNECT, "(main) RESET_LOCAL executed at channel:", ev.Tcd.MakeKey())

			case trudp.SEND_RESET:
				tru.Log(trudp.CONNECT, "(main) SEND_RESET to channel:", ev.Tcd.MakeKey())

			default:
				tru.Log(trudp.ERROR, "(main)")
			}
			//}(ev)
		}
	}()

	// Ctrl+C process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGINT:
				fmt.Println("syscall.SIGINT")
				tru.Close()
			}
		}
	}()

	// Run trudp and start listen
	tru.Run()

	//time.Sleep(5 * time.Second)
	fmt.Println("bay...")
}

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

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

const MODULE = "main"

func main() {
	fmt.Println("TRUDP test application ver " + trudp.Version)

	var (
		rhost string
		rport int
		rchan int
		port  int

		// Logger parameters
		logLevel  string
		logFilter string

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
	flag.StringVar(&logLevel, "log-level", "CONNECT", "application log level")
	flag.StringVar(&logFilter, "log-filter", "", "application log filter")
	flag.IntVar(&sendSleepTime, "t", 0, "send timeout in microseconds")
	flag.BoolVar(&sendTest, "send-test", false, "send test data")
	flag.BoolVar(&sendAnswer, "answer", false, "send answer")
	flag.BoolVar(&showStat, "S", false, "show statistic")

	flag.Parse()

	for reconnectF := false; ; {

		tru := trudp.Init(port)

		// Set log level
		teolog.Init(logLevel, !noLogTime, log.LstdFlags|log.Lmicroseconds, logFilter)

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
					tcd.AllowSendTestMsg(sendTest)

					// Sender
					num := 0
					for tru.Running() {
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

					teolog.Log(teolog.CONNECT, MODULE, "(main) channel "+tcd.GetKey()+" sender stopped")
					if !tru.Running() {
						break
					}
					time.Sleep(5 * time.Second)
					teolog.Log(teolog.CONNECT, MODULE, "(main) reconnect")
				}

				teolog.Log(teolog.CONNECT, MODULE, "(main) sender worker stopped")
			}()
		}

		// Receiver
		go func() {
			defer tru.ChanEventClosed()
			for ev := range tru.ChanEvent() {
				switch ev.Event {

				case trudp.GOT_DATA:
					teolog.Log(teolog.DEBUG, MODULE, "(main) GOT_DATA: ", ev.Data, string(ev.Data), fmt.Sprintf("%.3f ms", ev.Tcd.TripTime()))
					if sendAnswer {
						ev.Tcd.WriteTo([]byte(string(ev.Data) + " - answer"))
					}

				// case trudp.SEND_DATA:
				// 	teolog.Log(teolog.DEBUG, MODULE, "(main) SEND_DATA:", ev.Data, string(ev.Data))

				case trudp.INITIALIZE:
					teolog.Log(teolog.CONNECT, MODULE, "(main) INITIALIZE, listen at:", string(ev.Data))

				case trudp.DESTROY:
					teolog.Log(teolog.CONNECT, MODULE, "(main) DESTROY", string(ev.Data))

				case trudp.CONNECTED:
					teolog.Log(teolog.CONNECT, MODULE, "(main) CONNECTED", string(ev.Data))

				case trudp.DISCONNECTED:
					teolog.Log(teolog.CONNECT, MODULE, "(main) DISCONNECTED", string(ev.Data))

				case trudp.RESET_LOCAL:
					teolog.Log(teolog.DEBUG, MODULE, "(main) RESET_LOCAL executed at channel:", ev.Tcd.GetKey())

				case trudp.SEND_RESET:
					teolog.Log(teolog.DEBUG, MODULE, "(main) SEND_RESET to channel:", ev.Tcd.GetKey())

				default:
					teolog.Log(teolog.ERROR, MODULE, "(main) event:", ev.Event)
				}
			}
		}()

		// Ctrl+C process
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				switch sig {
				case syscall.SIGINT:
					// reconnectF = true
					// tru.Close()
					// return

					var str string
					fmt.Print("\033[2K\033[0E" + "Press Q to exit or R to reconnect: ")
					fmt.Scanf("%s\n", &str)
					switch str {
					case "r", "R":
						reconnectF = true
						tru.Close()
						return
					case "q", "Q":
						reconnectF = false
						tru.Close()
					}
				case syscall.SIGCLD:
					fallthrough
				default:
					fmt.Printf("sig: %x\n", sig)
				}
			}
		}()

		// Run trudp and start listen
		tru.Run()

		if !reconnectF {
			fmt.Println("bay...")
			break
		}
		fmt.Println("reonnect...")
	}
}

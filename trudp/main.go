// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This is main Trudp sample application
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"

	"golang.org/x/crypto/ssh/terminal"
)

// MODULE is name in log
const MODULE = "(main)"

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
		logToSyslogF bool
		sendTest     bool
		showStat     bool
		sendAnswer   bool
	)

	flag.IntVar(&maxQueueSize, "Q", trudp.DefaultQueueSize, "maximum send and receive queues size")
	flag.IntVar(&port, "p", 0, "this host port (to remote hosts connect to this host)")
	flag.StringVar(&rhost, "a", "", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 1, "remote host channel (to connect to remote host)")
	flag.IntVar(&rport, "r", 0, "remote host port (to connect to remote host)")
	flag.StringVar(&logLevel, "log-level", "CONNECT", "application log level")
	flag.StringVar(&logFilter, "log-filter", "", "application log filter")
	flag.BoolVar(&logToSyslogF, "log-to-syslog", false, "save log to syslog")
	flag.IntVar(&sendSleepTime, "t", 0, "send timeout in microseconds")
	flag.BoolVar(&sendTest, "send-test", false, "send test data")
	flag.BoolVar(&sendAnswer, "answer", false, "send answer")
	flag.BoolVar(&showStat, "S", false, "show statistic")

	flag.Parse()

	for reconnectF := false; ; {

		tru := trudp.Init(&port)

		// Set log level
		teolog.Init(logLevel, log.Lmicroseconds|log.Lshortfile, logFilter,
			logToSyslogF, "trudp")

		// Set 'show statictic' flag
		tru.SetShowStatistic(showStat)

		// Set default queue size
		tru.SetDefaultQueueSize(maxQueueSize)

		// Event Receiver
		go func() {
			defer tru.ChanEventClosed()
			teolog.Log(teolog.CONNECT, MODULE, "event receiver started")

			for ev := range tru.ChanEvent() {
				switch ev.Event {

				case trudp.EvGotData:
					teolog.Log(teolog.DEBUGv, MODULE, "GOT_DATA: ",
						ev.Data, string(ev.Data), fmt.Sprintf("%.3f ms", ev.Tcd.TripTime()))
					if sendAnswer {
						ev.Tcd.WriteNowait([]byte(string(ev.Data)+" - answer"), func() {})
						// ev.Tcd.Write([]byte(string(ev.Data) + " - answer"))
					}

				// case trudp.SEND_DATA:
				// 	teolog.Log(teolog.DEBUG, MODULE, "SEND_DATA:", ev.Data, string(ev.Data))

				case trudp.EvInitialize:
					teolog.Log(teolog.CONNECT, MODULE, "INITIALIZE, start listen at:",
						string(ev.Data))

				case trudp.EvDestroy:
					teolog.Log(teolog.CONNECT, MODULE, "DESTROY",
						string(ev.Data))

				case trudp.EvConnected:
					teolog.Log(teolog.CONNECT, MODULE, "CONNECTED to:",
						string(ev.Data))

				case trudp.EvDisconnected:
					teolog.Log(teolog.CONNECT, MODULE, "DISCONNECTED from:",
						string(ev.Data))

				case trudp.EvResetLocal:
					teolog.Log(teolog.DEBUG, MODULE, "RESET_LOCAL executed at channel:",
						ev.Tcd.GetKey())

				case trudp.EvSendReset:
					teolog.Log(teolog.DEBUG, MODULE, "SEND_RESET to channel:",
						ev.Tcd.GetKey())

				default:
					teolog.Errorf(MODULE,
						"event: %d, data_len: %d, data: %v %s\n",
						ev.Event, len(ev.Data), ev.Data, string(ev.Data),
					)
				}
			}
		}()

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
						_, err := tcd.Write(data)
						if err != nil {
							break
						}
						num++
					}

					teolog.Log(teolog.CONNECT, MODULE, "channel "+
						tcd.GetKey()+" sender stopped")
					if !tru.Running() {
						break
					}
					time.Sleep(5 * time.Second)
					teolog.Log(teolog.CONNECT, MODULE, "reconnect")
				}

				teolog.Log(teolog.CONNECT, MODULE, "sender worker stopped")
			}()
		}

		// Ctrl+C process
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				switch sig {
				case syscall.SIGINT:

					// Set terminal to raw mode (fd 0 is stdin)
					state, err := terminal.MakeRaw(0)
					if err != nil {
						log.Fatalln("setting stdin to raw:", err)
					}

					// Restore terminal
					rest := func() {
						if err := terminal.Restore(0, state); err != nil {
							log.Println("warning, failed to restore terminal:", err)
						}
					}

					// Get one rune
					getch := func() (r rune) {
						in := bufio.NewReader(os.Stdin)
						r, _, err = in.ReadRune()
						if err != nil {
							log.Println("stdin:", err)
						}
						// fmt.Printf("read rune %q\r\n", r)
						return
					}

					fmt.Print("\033[2K\033[0E" + "Press Q to exit or R to reconnect, or any other key to continue\r\n")
					switch getch() {
					case 'r', 'R':
						reconnectF = true
						tru.Close()
						return
					case 'q', 'Q', '\x03':
						reconnectF = false
						tru.Close()
					}
					rest()

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

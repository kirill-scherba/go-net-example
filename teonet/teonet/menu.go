package teonet

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// Teonet hotkey menu module

func (teo *Teonet) createMenu() {
	if !teo.param.ForbidHotkeysF {

		//in := bufio.NewReader(os.Stdin)
		setLogLevel := func(loglevel int) {
			fmt.Print("\b")
			logstr := teolog.LevelString(loglevel)
			if teo.param.LogLevel == logstr {
				logstr = teolog.LevelString(teolog.NONE)
			}
			teo.param.LogLevel = logstr
			teolog.Init(teo.param.LogLevel, true,
				log.LstdFlags|log.Lmicroseconds|log.Lshortfile, teo.param.LogFilter)
		}
		readString := func(in *bufio.Reader, prompt string) (str string) {
			var err error
			fmt.Print(prompt)
			if str, err = in.ReadString('\n'); err == nil {
				str = strings.TrimRightFunc(str, func(c rune) bool {
					return c == '\r' || c == '\n'
				})
			}
			return
		}
		readInt := func(in *bufio.Reader, prompt string) (i int) {
			i, _ = strconv.Atoi(readString(in, prompt))
			return
		}

		teo.menu = teokeys.CreateMenu("\bHot keys list:", "")

		teo.menu.Add([]int{'h', '?', 'H'}, "show this help screen", func() {
			//logLevel := param.LogLevel
			setLogLevel(teolog.NONE)
			teo.menu.Usage()
		})

		teo.menu.Add('p', "show peers", func() {
			var mode string
			teo.param.ShowPeersStatF = !teo.param.ShowPeersStatF
			if !teo.param.ShowPeersStatF {
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowClientsStatF = false
				teo.param.ShowTrudpStatF = false
				teo.td.ShowStatistic(teo.param.ShowTrudpStatF)
				teo.arp.print()
				mode = "on"
			}
			fmt.Println("\nshow peers", mode)
		})

		teo.menu.Add('u', "show trudp statistics", func() {
			var mode string
			teo.param.ShowTrudpStatF = !teo.param.ShowTrudpStatF
			teo.td.ShowStatistic(teo.param.ShowTrudpStatF)
			if !teo.param.ShowTrudpStatF {
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowPeersStatF = false
				teo.param.ShowClientsStatF = false
				mode = "on"
			}
			fmt.Println("\nshow trudp", mode)
		})

		if teo.l0.allow {
			teo.menu.Add('l', "show clients", func() {
				var mode string
				teo.param.ShowClientsStatF = !teo.param.ShowClientsStatF
				if !teo.param.ShowClientsStatF {
					mode = "off" + "\033[r" + "\0338"
				} else {
					teo.param.ShowPeersStatF = false
					teo.param.ShowTrudpStatF = false
					teo.td.ShowStatistic(teo.param.ShowTrudpStatF)
					teo.l0.stat.process()
					mode = "on"
				}
				fmt.Println("\nshow clients", mode)
			})
		}

		teo.menu.Add('n', "show 'none' log messages", func() { setLogLevel(teolog.NONE) })
		teo.menu.Add('c', "show 'connect' log messages", func() { setLogLevel(teolog.CONNECT) })
		teo.menu.Add('d', "show 'debug' log messages", func() { setLogLevel(teolog.DEBUG) })
		teo.menu.Add('v', "show 'debug_v log' messages", func() { setLogLevel(teolog.DEBUGv) })
		teo.menu.Add('w', "show 'debug_vv' log messages", func() { setLogLevel(teolog.DEBUGvv) })

		teo.menu.Add('f', "set log messages filter", func() {
			logLevel := teo.param.LogLevel
			setLogLevel(teolog.NONE)
			teo.menu.Stop(true)

			go func() {
				in := bufio.NewReader(os.Stdin)
				teo.param.LogFilter = readString(in, "\b"+"enter log filter: ")
				teolog.SetFilter(teo.param.LogFilter)
				setLogLevel(teolog.LogLevel(logLevel))
				teo.menu.Stop(false)
			}()
		})

		teo.menu.Add('s', "send command", func() {
			setLogLevel(teolog.NONE)
			teo.menu.Stop(true)

			go func() {
				defer teo.menu.Stop(false)
				in := bufio.NewReader(os.Stdin)
				fmt.Printf("send command to peer\n")
				var data []byte
				var to, str string
				var cmd, answerCmd int
				if to = readString(in, "to: "); to == "" {
					return
				}
				if cmd = readInt(in, "cmd: "); cmd == 0 {
					return
				}
				if str = readString(in, "data: "); str != "" {
					data = []byte(str)
				}
				answerCmd = readInt(in, "cmd(answer): ")

				// Send to Teonet peer
				if err := teo.SendTo(to, byte(cmd), data); err != nil {
					fmt.Printf("error: %s\n", err.Error())
					return
				}
				fmt.Printf("sent to: %s, cmd: %d, data: %s\n", to, cmd, str)

				// Wait answer from Teonet peer
				if answerCmd > 0 {
					r := <-teo.WaitFrom(to, byte(answerCmd), 1*time.Second)
					// Show timeout error
					if r.Err != nil {
						fmt.Printf("error: %s\n", r.Err.Error())
						return
					}
					// Show valid result
					fmt.Printf(""+
						"got data (string): %v\n"+
						"got data (buffer): %v\n", string(r.Data), r.Data)
				}
			}()
		})

		teo.menu.Add('x', "clear screen", func() {
			fmt.Print("\033[0;0H" + teokeys.ANSICls)
		})

		teo.menu.Add('r', "reconnect this application", func() {
			fmt.Printf("\b")
			teo.Reconnect()
		})

		teo.menu.Add('q', "quit this application", func() {
			logLevel := teo.param.LogLevel
			setLogLevel(teolog.NONE)
			fmt.Printf("\bPress y to quit application: ")
			teo.menu.Stop(true)
			ch := teo.menu.Getch()
			fmt.Println()
			setLogLevel(teolog.LogLevel(logLevel))
			if ch == 'y' || ch == 'Y' {
				teo.menu.Stop(false)
				teo.menu.Quit()
				teo.Close()
			} else {
				teo.menu.Stop(false)
			}
		})
	}
}

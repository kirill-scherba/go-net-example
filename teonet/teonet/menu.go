package teonet

import (
	"fmt"
	"log"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// Teonet hotkey menu module

func (teo *Teonet) createMenu() {
	if !teo.param.ForbidHotkeysF {

		setLogLevel := func(loglevel int) {
			fmt.Print("\b")
			logstr := teolog.LevelString(loglevel)
			if teo.param.LogLevel == logstr {
				logstr = teolog.LevelString(teolog.NONE)
			}
			teo.param.LogLevel = logstr
			teolog.Init(teo.param.LogLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile, teo.param.LogFilter)
		}

		teo.menu = teokeys.CreateMenu("\bHot keys list:", "")

		teo.menu.Add([]int{'h', '?', 'H'}, "show this help screen", func() {
			//logLevel := param.LogLevel
			setLogLevel(teolog.NONE)
			teo.menu.Usage()
		})

		teo.menu.Add('p', "show peers", func() {
			var mode string
			if teo.param.ShowPeersStatF {
				teo.param.ShowPeersStatF = false
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowPeersStatF = true
				teo.param.ShowTrudpStatF = false
				teo.arp.print()
				mode = "on"
			}
			teo.td.ShowStatistic(teo.param.ShowTrudpStatF)
			fmt.Println("\nshow peers", mode)
		})

		teo.menu.Add('u', "show trudp statistics", func() {
			var mode string
			if teo.param.ShowTrudpStatF {
				teo.param.ShowTrudpStatF = false
				mode = "off" + "\033[r" + "\0338"
			} else {
				teo.param.ShowTrudpStatF = true
				teo.param.ShowPeersStatF = false
				mode = "on"
			}
			teo.td.ShowStatistic(teo.param.ShowTrudpStatF)
			fmt.Println("\nshow trudp", mode)
		})

		teo.menu.Add('n', "show 'none' log messages", func() { setLogLevel(teolog.NONE) })
		teo.menu.Add('c', "show 'connect' log messages", func() { setLogLevel(teolog.CONNECT) })
		teo.menu.Add('d', "show 'debug' log messages", func() { setLogLevel(teolog.DEBUG) })
		teo.menu.Add('v', "show 'debug_v log' messages", func() { setLogLevel(teolog.DEBUGv) })
		teo.menu.Add('w', "show 'debug_vv' log messages", func() { setLogLevel(teolog.DEBUGvv) })

		teo.menu.Add('f', "set log messages filter", func() {
			logLevel := teo.param.LogLevel
			setLogLevel(teolog.NONE)
			teo.menu.Stop(true)

			func() {
				var filter string
				fmt.Printf("\benter log filter: ")
				fmt.Scanf("%s", &filter)
				teo.param.LogFilter = filter
				teolog.SetFilter(filter)

				setLogLevel(teolog.LogLevel(logLevel))
				teo.menu.Stop(false)
			}()
		})

		teo.menu.Add('r', "reconnect this application", func() {
			teo.reconnect = true
			teo.menu.Quit()
			teo.Close()
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

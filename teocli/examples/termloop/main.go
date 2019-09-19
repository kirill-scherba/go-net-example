// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet client using termloop game engine

package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	tl "github.com/JoelOtter/termloop"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// Version this teonet application version
const Version = "0.0.1"

func main() {
	fmt.Println("Teocli termloop application ver " + Version)

	// game()
	// os.Exit(0)

	// Flags variables
	var name string      // this client name
	var peer string      // remote server name (to send commands to)
	var raddr string     // remote host address
	var rport, rchan int // remote host port and channel (for TRUDP)
	var timeout int      // send echo timeout (in microsecond)
	var tcp bool         // connect by TCP flag

	// Flags
	flag.StringVar(&name, "n", "teocli-go-main-test-01", "this application name")
	flag.StringVar(&peer, "peer", "ps-server", "remote server name (to send commands to)")
	flag.StringVar(&raddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 0, "remote host channel (to connect to remote host TRUDP channel)")
	flag.IntVar(&rport, "r", 9010, "remote host port (to connect to remote host)")
	flag.IntVar(&timeout, "t", 1000000, "send echo timeout (in microsecond)")
	flag.BoolVar(&tcp, "tcp", false, "connect by TCP")
	flag.Parse()

	// Run teocli
	func() {
		for {
			var network string
			running := true

			// Connect to L0 server
			if tcp {
				network = "TCP"
			} else {
				network = "TRUDP"
			}
			fmt.Printf("try %s connecting to %s:%d ...\n", network, raddr, rport)
			teo, err := teocli.Connect(raddr, rport, tcp)
			if err != nil {
				fmt.Println(err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Send L0 login (requered after connect)
			fmt.Printf("send login\n")
			if _, err := teo.SendLogin(name); err != nil {
				panic(err)
			}

			// Send peers command
			fmt.Printf("send peers request\n")
			teo.Send(teocli.CmdLPeers, peer, nil)

			// Start game request
			teo.Send(129, peer, nil)

			teo.Send(teocli.CmdLPeers, peer, nil)

			// Sender (send echo in loop)
			go func() {
				for i := 0; running; i++ {
					//switch {
					// // Send peers command
					// case i%9 == 1:
					// 	fmt.Printf("send peers request (%d,%d)\n", i, i%9)
					// 	teo.Send(teocli.CmdLPeers, peer, nil)
					//
					// // Send large data packet with cmd 129
					// case i%19 == 1:
					// 	data := append([]byte(strings.Repeat("Q", 2000)), 0)
					// 	fmt.Printf("send large data packet with cmd 129, data_len: %d (%d,%d)\n",
					// 		len(data), i, i%19)
					// 	teo.Send(129, peer, data)
					//
					// // Send echo
					// default:
					// 	fmt.Printf("send echo %d\n", i)
					// 	teo.SendEcho(peer, fmt.Sprintf("Hello from Go(No %d)!", i))
					//
					//}
					time.Sleep(time.Duration(timeout) * time.Microsecond)
				}
			}()

			// Reader (read data and display it)
			for {
				packet, err := teo.Read()
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Printf("got cmd %d from %s, data len: %d, data: %v\n",
					packet.Command(), packet.From(), len(packet.Data()), packet.Data())

				switch packet.Command() {

				// Game Started
				case 129:
					go game()

				// Echo answer
				case teocli.CmdLEchoAnswer:
					if t, err := packet.TripTime(); err != nil {
						fmt.Println("trip time error:", err)
					} else {
						fmt.Println("trip time (ms):", t)
					}

				// Peers answer
				case teocli.CmdLPeersAnswer:
					ln := strings.Repeat("-", 59)
					fmt.Println("PeerAnswer received\n"+ln, "\n"+packet.Peers()+ln)
				}
			}
			teo.Disconnect()
			running = false
			time.Sleep(5 * time.Second)
		}
	}()
}

type Player struct {
	*tl.Entity
	prevX int
	prevY int
	level *tl.BaseLevel
}

// func (player *Player) Draw(screen *tl.Screen) {
// 	screenWidth, screenHeight := screen.Size()
// 	x, y := player.Position()
// 	player.level.SetOffset(screenWidth/2-x, screenHeight/2-y)
// 	player.Entity.Draw(screen)
// }

func (player *Player) Tick(event tl.Event) {
	if event.Type == tl.EventKey { // Is it a keyboard event?
		player.prevX, player.prevY = player.Position()
		switch event.Key { // If so, switch on the pressed key.
		case tl.KeyArrowRight:
			player.SetPosition(player.prevX+1, player.prevY)
		case tl.KeyArrowLeft:
			player.SetPosition(player.prevX-1, player.prevY)
		case tl.KeyArrowUp:
			player.SetPosition(player.prevX, player.prevY-1)
		case tl.KeyArrowDown:
			player.SetPosition(player.prevX, player.prevY+1)
		}
	}
}

func (player *Player) Collide(collision tl.Physical) {
	// Check if it's a Rectangle we're colliding with
	if _, ok := collision.(*tl.Rectangle); ok {
		player.SetPosition(player.prevX, player.prevY)
	}
}

// Run game
func game() {
	game := tl.NewGame()
	game.Screen().SetFps(30)
	level := tl.NewBaseLevel(tl.Cell{
		Bg: tl.ColorGreen,
		Fg: tl.ColorBlack,
		Ch: 'H',
	})
	level.AddEntity(tl.NewRectangle(10, 10, 50, 20, tl.ColorBlue))
	player := Player{
		Entity: tl.NewEntity(1, 1, 1, 1),
		level:  level,
	}
	// Set the character at position (0, 0) on the entity.
	player.SetCell(0, 0, &tl.Cell{Fg: tl.ColorRed, Ch: '옷'})
	level.AddEntity(&player)
	game.Screen().SetLevel(level)
	game.Start()
}

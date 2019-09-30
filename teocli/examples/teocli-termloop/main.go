// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet client using linux terminal game engine
//
// This is simple terminal game with teonet client connected to teonet l0
// server and teonet room controller. This application connect to l0 server
// first and than request room in room controller. When room controller answer
// with room request answer this game application can send its hero position and
// show position of other players entered to the same room.
//
// Install client and server:
//
//  go get github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go get github.com/kirill-scherba/teonet-go/teonet
//
// Run server applications:
//
//  # run teonet l0 server
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teonet
//  go run . -p 7050 -l0-allow teo-l0
//
//  # run teonet room controller
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teonet/app/teoroom
//  go run . -r 7050 teo-room
//
// Run this game client application:
//
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go run . -r 7050 -peer teo-room -n game-01
//
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go run . -r 7050 -peer teo-room -n game-02
//
// To exit from this game type Ctrl+C twice. When you start play next time
// you'll be connected to another room.
//
package main

import (
	"flag"
	"fmt"
	"time"

	tl "github.com/JoelOtter/termloop"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// Version this teonet application version
const Version = "0.0.1"

// Game levels
const (
	Game int = iota
	Menu
	Meta
)

// Teogame this game data structure
type Teogame struct {
	game   *tl.Game               // Game
	level  []*tl.BaseLevel        // Game levels
	hero   *Hero                  // Game Hero
	player map[byte]*Player       // Game Players map
	state  *GameState             // Game state
	teo    *teocli.TeoLNull       // Teonet connetor
	peer   string                 // Teonet room controller peer name
	com    *outputCommands        // Teonet output commands receiver
	rra    *roomRequestAnswerData // Room request answer data
}

// main parse aplication parameters and connect to Teonet. When teonet connected
// the game started
func main() {
	fmt.Println("Teocli termloop application ver " + Version)

	// Flags variables
	var name string  // this client name
	var peer string  // room controller peer name
	var raddr string // l0 server address
	var rport int    // l0 server port
	var timeout int  // reconnect timeout (in seconds)
	var tcp bool     // connect by TCP flag

	// Flags
	flag.StringVar(&name, "n", "teocli-go-main-test-01", "this application name")
	flag.StringVar(&peer, "peer", "teo-room", "teo-room peer name (to send commands to)")
	flag.StringVar(&raddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&rport, "r", 9010, "l0 server port (to connect to l0 server)")
	flag.BoolVar(&tcp, "tcp", false, "connect by TCP")
	flag.IntVar(&timeout, "t", 5, "reconnect after timeout (in second)")
	flag.Parse()

	// Run teonet game (connect to teonet, start game and process received commands)
	run(name, peer, raddr, rport, tcp, time.Duration(timeout)*time.Second)
}

// Run connect to teonet, start game and process received commands
func run(name, peer, raddr string, rport int, tcp bool, timeout time.Duration) (tg *Teogame) {
	tg = &Teogame{peer: peer, player: make(map[byte]*Player)}
	teocli.Run(name, raddr, rport, tcp, timeout, startCommand(tg), inputCommands(tg)...)
	return
}

// start create, initialize and start game
func (tg *Teogame) start(rra *roomRequestAnswerData) {
	tg.game = tl.NewGame()
	tg.game.Screen().SetFps(30)
	tg.rra = rra

	// Level 0: Game
	tg.level = append(tg.level, func() (level *tl.BaseLevel) {

		level = tl.NewBaseLevel(tl.Cell{
			Bg: tl.ColorBlack,
			Fg: tl.ColorWhite,
			Ch: ' ',
		})

		// Lake
		level.AddEntity(tl.NewRectangle(10, 5, 10, 5, tl.ColorWhite|tl.ColorBlack))

		// Game state
		tg.state = tg.newGameState(level)

		// Hero
		tg.hero = tg.addHero(level, int(rra.clientID)*3, 2)

		return
	}())

	// Level 1: Game menu
	tg.level = append(tg.level, func() (level *tl.BaseLevel) {
		level = tl.NewBaseLevel(tl.Cell{
			Bg: tl.ColorBlack,
			Fg: tl.ColorWhite,
			Ch: ' ',
		})
		tg.newGameMenu(level, " Game Over! ")
		return
	}())

	// Start and run
	gameLevel := Game
	tg.game.Screen().SetLevel(tg.level[gameLevel])
	_, err := tg.com.sendData(tg.hero)
	if err != nil {
		panic(err)
	}
	tg.game.Start()

	// When stopped (press exit from game or Ctrl+C)
	fmt.Printf("game stopped\n")
	tg.com.disconnect()
	tg.com.stop()
}

// reset reset game to its default values
func (tg *Teogame) reset(rra *roomRequestAnswerData) {
	for _, p := range tg.player {
		tg.level[Game].RemoveEntity(p)
	}
	tg.rra = rra
	tg.player = make(map[byte]*Player)
	tg.hero.SetPosition(int(rra.clientID)*3, 2)
	tg.game.Screen().SetLevel(tg.level[Game])
	tg.com.sendData(tg.hero)

	tg.state.setLoaded()
}

// setLevel switch to selected game level
func (tg *Teogame) setLevel(gameLevel int) {
	tg.game.Screen().SetLevel(tg.level[gameLevel])
}

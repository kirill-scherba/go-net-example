// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet client using termloop game engine

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"time"

	tl "github.com/JoelOtter/termloop"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// Version this teonet application version
const Version = "0.0.1"

// Teogame this game data structure
type Teogame struct {
	game   *tl.Game               // Game
	level  *tl.BaseLevel          // Game BaseLevel
	hero   *Hero                  // Game Hero
	player map[byte]*Player       // Game Players map
	teo    *teocli.TeoLNull       // Teonet connetor
	peer   string                 // Teonet room controller peer name
	com    *outputCommands        // Teonet output commands receiver
	rra    *roomRequestAnswerData // Room request answer data
}

// Player stucture of player
type Player struct {
	*tl.Entity
	prevX int
	prevY int
	level *tl.BaseLevel
	tg    *Teogame
}

// Hero struct of hero
type Hero struct {
	Player
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

// startGame initialize and start game
func (tg *Teogame) startGame(rra *roomRequestAnswerData) {
	tg.game = tl.NewGame()
	tg.game.Screen().SetFps(30)
	tg.rra = rra

	// Base level
	level := tl.NewBaseLevel(tl.Cell{
		Bg: tl.ColorBlack,
		Fg: tl.ColorWhite,
		Ch: ' ',
	})
	tg.level = level

	// Lake
	level.AddEntity(tl.NewRectangle(10, 5, 10, 5, tl.ColorWhite|tl.ColorBlack))

	// Hero
	tg.hero = tg.addHero(int(rra.clientID)*3, 0)

	// Start and run
	tg.game.Screen().SetLevel(level)
	_, err := tg.com.sendData(tg.hero)
	if err != nil {
		panic(err)
	}
	tg.game.Start()

	// When stopped (press exit from game or Ctrl+C)
	fmt.Printf("game stopped\n")
	tg.com.disconnect()
	//tg.teo.Disconnect()
}

// addHero add new Player to game
func (tg *Teogame) addHero(x, y int) (hero *Hero) {
	hero = &Hero{Player{
		Entity: tl.NewEntity(1, 1, 1, 1),
		level:  tg.level,
		tg:     tg,
	}}
	// Set the character at position (0, 0) on the entity.
	hero.SetCell(0, 0, &tl.Cell{Fg: tl.ColorGreen, Ch: 'Ω'})
	hero.SetPosition(x, y)
	tg.level.AddEntity(hero)
	return
}

// addPlayer add new Player to game or return existing if already exist
func (tg *Teogame) addPlayer(id byte) (player *Player) {
	player, ok := tg.player[id]
	if !ok {
		player = &Player{
			Entity: tl.NewEntity(2, 2, 1, 1),
			level:  tg.level,
			tg:     tg,
		}
		// Set the character at position (0, 0) on the entity.
		player.SetCell(0, 0, &tl.Cell{Fg: tl.ColorBlue, Ch: 'Ö'})
		tg.level.AddEntity(player)
		tg.player[id] = player
		//fmt.Printf("addPlayer, id: %d\n", id)
	}
	return
}

// Set player at center of map
// func (player *Player) Draw(screen *tl.Screen) {
// 	screenWidth, screenHeight := screen.Size()
// 	x, y := player.Position()
// 	player.level.SetOffset(screenWidth/2-x, screenHeight/2-y)
// 	player.Entity.Draw(screen)
// }

// Tick frame tick
func (player *Hero) Tick(event tl.Event) {
	if event.Type == tl.EventKey { // Is it a keyboard event?

		player.prevX, player.prevY = player.Position()

		// Save previouse position and set to new position
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

		// Check position changed and send to Teonet if so
		x, y := player.Position()
		if x != player.prevX || y != player.prevY {
			_, err := player.tg.com.sendData(player)
			if err != nil {
				panic(err)
			}
		}
	}
}

// Collide check colliding
func (player *Player) Collide(collision tl.Physical) {
	// Check if it's a Rectangle we're colliding with
	if _, ok := collision.(*tl.Rectangle); ok {
		player.SetPosition(player.prevX, player.prevY)
		_, err := player.tg.com.sendData(player)
		if err != nil {
			panic(err)
		}
	}
}

// MarshalBinary marshal players data to binary
func (player *Player) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	x, y := player.Position()
	binary.Write(buf, binary.LittleEndian, player.tg.rra.clientID)
	err = binary.Write(buf, binary.LittleEndian, int64(x))
	err = binary.Write(buf, binary.LittleEndian, int64(y))
	data = buf.Bytes()
	return
}

// UnmarshalBinary unmarshal binary data and sen it yo player
func (player *Player) UnmarshalBinary(data []byte) (err error) {
	var cliID byte
	var x, y int64
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.LittleEndian, &cliID)
	err = binary.Read(buf, binary.LittleEndian, &x)
	// if err != nil {
	// 	return
	// }
	err = binary.Read(buf, binary.LittleEndian, &y)
	// if err != nil {
	// 	return
	// }
	player.SetPosition(int(x), int(y))
	return
}

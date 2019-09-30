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

// Game levels
const (
	Game int = iota
	gameMenu
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

// newGame create, initialize and start game
func (tg *Teogame) newGame(rra *roomRequestAnswerData) {
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

		// Text entry
		tg.state = &GameState{tl.NewText(0, 0, "time: ", tl.ColorBlack, tl.ColorBlue), Loaded, 0}
		level.AddEntity(tg.state)

		// Hero
		tg.hero = tg.addHero(level, int(rra.clientID)*3, 2)

		return
	}())

	// Level 1: Game over
	tg.level = append(tg.level, func() (level *tl.BaseLevel) {
		level = tl.NewBaseLevel(tl.Cell{
			Bg: tl.ColorBlack,
			Fg: tl.ColorWhite,
			Ch: ' ',
		})
		level.AddEntity(tg.newGameMenu(" Game Over! "))
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

// gameMenu switch to 'game over' screen
func (tg *Teogame) gameMenu() {
	tg.game.Screen().SetLevel(tg.level[gameMenu])
}

// roomRequest execute start command (as usual it send room request command)
func (tg *Teogame) roomRequest() {
	tg.com.start.Command(tg.teo, nil)
}

// resetGame reset game to it default values
func (tg *Teogame) resetGame(rra *roomRequestAnswerData) {
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

// addHero add Hero to game
func (tg *Teogame) addHero(level *tl.BaseLevel, x, y int) (hero *Hero) {
	hero = &Hero{Player{
		Entity: tl.NewEntity(1, 1, 1, 1),
		level:  level,
		tg:     tg,
	}}
	// Set the character at position (0, 0) on the entity.
	hero.SetCell(0, 0, &tl.Cell{Fg: tl.ColorGreen, Ch: 'Ω'})
	hero.SetPosition(x, y)
	level.AddEntity(hero)
	return
}

// addPlayer add new Player to game or return existing if already exist
func (tg *Teogame) addPlayer(level *tl.BaseLevel, id byte) (player *Player) {
	player, ok := tg.player[id]
	if !ok {
		player = &Player{
			Entity: tl.NewEntity(2, 2, 1, 1),
			level:  level,
			tg:     tg,
		}
		// Set the character at position (0, 0) on the entity.
		player.SetCell(0, 0, &tl.Cell{Fg: tl.ColorBlue, Ch: 'Ö'})
		level.AddEntity(player)
		tg.player[id] = player
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
	if player.tg.state.State() == Running && event.Type == tl.EventKey {
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
	if err = binary.Write(buf, binary.LittleEndian, int64(x)); err != nil {
		return
	} else if err = binary.Write(buf, binary.LittleEndian, int64(y)); err != nil {
		return
	}
	data = buf.Bytes()
	return
}

// UnmarshalBinary unmarshal binary data and sen it yo player
func (player *Player) UnmarshalBinary(data []byte) (err error) {
	var cliID byte
	var x, y int64
	buf := bytes.NewReader(data)
	if err = binary.Read(buf, binary.LittleEndian, &cliID); err != nil {
		return
	} else if err = binary.Read(buf, binary.LittleEndian, &x); err != nil {
		return
	} else if err = binary.Read(buf, binary.LittleEndian, &y); err != nil {
		return
	}
	player.SetPosition(int(x), int(y))
	return
}

// Game states
const (
	Loaded int = iota
	Running
	Finished
)

// GameState hold game state and some text entrys
type GameState struct {
	*tl.Text     // Text entry
	state    int // Game state
	count    int // Frame counter
}

// State return game state
func (state *GameState) State() int {
	return state.state
}

// String return string with state name
func (state *GameState) String() string {
	switch state.state {
	case Loaded:
		return "Loaded"
	case Running:
		return "Running"
	default:
		return "Undefined"
	}
}

// setRunning set running state
func (state *GameState) setRunning() {
	state.count = 0
	state.state = Running
}

// setLoaded set loaded state
func (state *GameState) setLoaded() {
	state.count = 0
	state.state = Loaded
}

// Draw GameState object
func (state *GameState) Draw(screen *tl.Screen) {
	state.count++
	state.Text.SetText(
		fmt.Sprintf(
			"state: %s, time: %.3f",
			state.String(), float64(state.count)/30.0,
		))
	switch state.State() {
	case Loaded:
		state.Text.SetColor(tl.ColorBlack, tl.ColorYellow)
	case Running:
		state.Text.SetColor(tl.ColorBlack, tl.ColorGreen)
	}
	state.Text.Draw(screen)
}

// GameMenu is type of text
type GameMenu struct {
	t  []*tl.Text
	tg *Teogame
}

// newGameMenu create GameOverText object
func (tg *Teogame) newGameMenu(txt string) (text *GameMenu) {
	t := []*tl.Text{}
	t = append(t,
		tl.NewText(0, 0, txt, tl.ColorBlack, tl.ColorBlue),
		tl.NewText(0, 0, "", tl.ColorBlack, tl.ColorBlue),
		tl.NewText(0, 0, " 'g' - start new game ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, " 'm' - meta ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, "   ----------------   ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, " press Ctrl+C to quit ", tl.ColorDefault, tl.ColorBlack),
	)
	text = &GameMenu{t, tg}
	return
}

// Draw game over text
func (got *GameMenu) Draw(screen *tl.Screen) {
	screenWidth, screenHeight := screen.Size()
	var x, y int
	for i, t := range got.t {
		width, height := t.Size()
		if i < 3 {
			x, y = (screenWidth-width)/2, i+(screenHeight-height)/2
		} else {
			y++
		}
		t.SetPosition(x, y)
		t.Draw(screen)
	}
}

// Tick check key pressed and start new game or quit
func (got *GameMenu) Tick(event tl.Event) {
	if event.Type == tl.EventKey { // Is it a keyboard event?
		switch event.Ch { // If so, switch on the pressed key.
		case 'g':
			got.tg.roomRequest()
		}
	}
}

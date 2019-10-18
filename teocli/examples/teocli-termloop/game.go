package main

import (
	"fmt"

	tl "github.com/JoelOtter/termloop"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// Teogame this game data structure
type Teogame struct {
	game   *tl.Game               // Game
	level  []*tl.BaseLevel        // Game levels
	hero   *Hero                  // Game Hero
	player map[byte]*Player       // Game Players map
	state  *GameState             // Game state
	menu   *GameMenu              // Game menu
	teo    *teocli.TeoLNull       // Teonet connetor
	peer   string                 // Teonet room controller peer name
	com    *outputCommands        // Teonet output commands receiver
	rra    *roomRequestAnswerData // Room request answer data
}

// Game levels
const (
	Game int = iota
	Menu
	Meta
)

// start create, initialize and start the game
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
		var x = 0
		if rra != nil {
			x = int(rra.clientID) * 3
		}
		tg.hero = tg.addHero(level, x, 2)

		return
	}())

	// Level 1: Game menu
	tg.level = append(tg.level, func() (level *tl.BaseLevel) {
		level = tl.NewBaseLevel(tl.Cell{
			Bg: tl.ColorBlack,
			Fg: tl.ColorWhite,
			Ch: ' ',
		})
		tg.menu = tg.newGameMenu(level, " New Game! ")
		return
	}())

	// Start and run
	if rra != nil {
		tg.game.Screen().SetLevel(tg.level[Game])
		_, err := tg.com.sendData(tg.hero)
		if err != nil {
			panic(err)
		}
	} else {
		tg.game.Screen().SetLevel(tg.level[Menu])
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

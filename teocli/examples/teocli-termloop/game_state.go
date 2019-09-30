package main

import (
	"fmt"

	tl "github.com/JoelOtter/termloop"
)

// Game states
const (
	Loading int = iota
	Running
	Finished
)

// GameState hold game state and some text entrys
type GameState struct {
	*tl.Text     // Text entry
	state    int // Game state
	count    int // Frame counter
}

// newGameState create new GameState
func (tg *Teogame) newGameState(level *tl.BaseLevel) (state *GameState) {
	state = &GameState{
		tl.NewText(0, 0, "time: ", tl.ColorBlack, tl.ColorBlue),
		Loading,
		0,
	}
	level.AddEntity(state)
	return
}

// State return game state
func (state *GameState) State() int {
	return state.state
}

// String return string with state name
func (state *GameState) String() string {
	switch state.state {
	case Loading:
		return "Loading"
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
	state.state = Loading
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
	case Loading:
		state.Text.SetColor(tl.ColorBlack, tl.ColorYellow)
	case Running:
		state.Text.SetColor(tl.ColorBlack, tl.ColorGreen)
	}
	state.Text.Draw(screen)
}

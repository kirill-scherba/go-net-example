package main

import (
	"bytes"
	"encoding/binary"

	tl "github.com/JoelOtter/termloop"
)

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

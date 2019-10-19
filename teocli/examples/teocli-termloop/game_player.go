package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"

	tl "github.com/JoelOtter/termloop"
)

// Player data stucture
type Player struct {
	*tl.Entity
	prevX int
	prevY int
	level *tl.BaseLevel
	tg    *Teogame
}

// Hero struct of hero
type Hero struct {
	Player       // Player data
	BotMode      // Bot strategy
	bot     bool // Bot mode enable
}

// BotMode is bot strategy
type BotMode struct {
	xDirection int
	yDirection int
}

// addHero add Hero to game
func (tg *Teogame) addHero(level *tl.BaseLevel, x, y int) (hero *Hero) {
	hero = &Hero{Player: Player{
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
func (hero *Hero) Tick(event tl.Event) {
	if hero.tg.state.State() == Running && (event.Type == tl.EventKey || !hero.bot) {

		// Get current position
		hero.prevX, hero.prevY = hero.Position()

		// Save previouse position and set to new position
		switch event.Key { // If so, switch on the pressed key.
		case tl.KeyArrowRight:
			hero.SetPosition(hero.prevX+1, hero.prevY)
		case tl.KeyArrowLeft:
			hero.SetPosition(hero.prevX-1, hero.prevY)
		case tl.KeyArrowUp:
			hero.SetPosition(hero.prevX, hero.prevY-1)
		case tl.KeyArrowDown:
			hero.SetPosition(hero.prevX, hero.prevY+1)
		}

		// Set new position in bot mode
		if !hero.bot {

			const (
				MaxX            = 50
				MaxY            = 50
				ChangeDirection = 5
			)

			if hero.BotMode.xDirection == 0 && hero.BotMode.yDirection == 0 {
				hero.BotMode.xDirection = rand.Intn(2) - 1
				hero.BotMode.yDirection = rand.Intn(2) - 1
			}

			hero.SetPosition(hero.prevX+hero.BotMode.xDirection, hero.prevY+hero.BotMode.yDirection)

			if hero.BotMode.xDirection != 0 {
				switch {

				case hero.prevX > MaxX:
					hero.BotMode.xDirection = -1

				case hero.prevX <= 1:
					hero.BotMode.xDirection = 1

				default:
					if rand.Intn(99) < ChangeDirection {
						hero.BotMode.xDirection = rand.Intn(2) - 1
					}
				}
			}

			if hero.BotMode.yDirection != 0 {
				switch {

				case hero.prevY > MaxY:
					hero.BotMode.yDirection = -1

				case hero.prevY <= 1:
					hero.BotMode.yDirection = 1

				default:
					if rand.Intn(99) < ChangeDirection {
						hero.BotMode.yDirection = rand.Intn(2) - 1
					}
				}
			}
		}

		// Check position changed and send it to Teonet if so
		x, y := hero.Position()
		if x != hero.prevX || y != hero.prevY {
			_, err := hero.tg.com.sendData(hero)
			if err != nil {
				panic(err)
			}
		}
	}
}

// AutoPlay get auto play mode
func (hero *Hero) AutoPlay() bool {
	return hero.bot
}

// SetAutoPlay set auto play mode on-off
func (hero *Hero) SetAutoPlay(bot bool) {
	hero.bot = bot
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

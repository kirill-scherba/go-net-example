package main

import tl "github.com/JoelOtter/termloop"

// GameMenu is text entries collection
type GameMenu struct {
	t  []*tl.Text
	tg *Teogame
}

// newGameMenu create GameOverText object
func (tg *Teogame) newGameMenu(level *tl.BaseLevel, txt string) (menu *GameMenu) {
	t := []*tl.Text{}
	t = append(t,
		tl.NewText(0, 0, txt, tl.ColorBlack, tl.ColorBlue),
		tl.NewText(0, 0, "", tl.ColorBlack, tl.ColorBlue),
		tl.NewText(0, 0, " 'g' - start new game ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, " 'm' - meta ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, "   ----------------   ", tl.ColorDefault, tl.ColorBlack),
		tl.NewText(0, 0, " press Ctrl+C to quit ", tl.ColorDefault, tl.ColorBlack),
	)
	menu = &GameMenu{t, tg}
	level.AddEntity(menu)
	return
}

// Draw game over text
func (menu *GameMenu) Draw(screen *tl.Screen) {
	screenWidth, screenHeight := screen.Size()
	var x, y int
	for i, t := range menu.t {
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
func (menu *GameMenu) Tick(event tl.Event) {
	if event.Type == tl.EventKey { // Is it a keyboard event?
		switch event.Ch { // If so, switch on the pressed key.
		case 'g':
			menu.tg.com.start.Command(menu.tg.teo, nil)
		}
	}
}

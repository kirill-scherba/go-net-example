package main

import (
	"os"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// startCommand receiver first outup command
type startCommand struct {
	tg      *Teogame
	running bool
}

// newStartCommand return start command receiver
func newStartCommand(tg *Teogame) teocli.StartCommand {
	start := &startCommand{tg, true}
	tg.com = &outputCommands{tg: tg, start: start}
	return start
}

// startCommand command methods
func (p *startCommand) Command(teo *teocli.TeoLNull, pac *teocli.Packet) {
	if p.tg.teo == nil {

		// Redirect standart output to file
		f, _ := os.OpenFile("/tmp/teocli-termloop",
			os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
		os.Stdout = f
		os.Stderr = f

		p.tg.teo = teo
		go p.tg.start(nil)
		return
	}
	p.tg.teo = teo
	teoroomcli.RoomRequest(p.tg.teo, p.tg.peer, nil)
}

// Running used inside teocli.Run() function and return running flag
func (p *startCommand) Running() bool { return p.running }

// Disconnected calls when connection lost and start reconnecting
func (p *startCommand) Disconnected() {
	p.tg.game.Screen().SetLevel(p.tg.level[Menu])
	// screen := p.tg.game.Screen()
	// w,h := screen.Size()
	//p.tg.level[Menu].SetOffset(100, 100)
}

// Stop used to stop teocli.Run() and set running flag to false
func (p *startCommand) Stop() {
	// Sleep to get time disconnect from room controller packet go out and
	<-time.After(50 * time.Millisecond)
	p.running = false
	p.tg.teo.Disconnect()
}

// ---------------------------------------------------------------------------

// outputCommands teonet output commands receiver
type outputCommands struct {
	tg    *Teogame
	start teocli.StartCommand
}

// disconnect [out] send disconnect (leave room) command to room controller
func (com *outputCommands) disconnect() {
	if com.tg.rra == nil {
		return
	}
	teoroomcli.Disconnect(com.tg.teo, com.tg.peer, com.tg.rra.clientID)
}

// sendData [out] send data command to room controller
func (com *outputCommands) sendData(i ...interface{}) (num int, err error) {
	return teoroomcli.Data(com.tg.teo, com.tg.peer, i...)
}

// stop teocli.Run()
func (com *outputCommands) stop() {
	com.start.Stop()
}

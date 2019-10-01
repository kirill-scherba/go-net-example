package main

import (
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// startCommand receiver first outup command
type startCommandD struct {
	tg      *Teogame
	running bool
}

// startCommand return start command receiver
func startCommand(tg *Teogame) teocli.StartCommand {
	start := &startCommandD{tg, true}
	tg.com = &outputCommands{tg: tg, start: start}
	return start
}

// startCommand command methods
func (p *startCommandD) Command(teo *teocli.TeoLNull, pac *teocli.Packet) {
	if p.tg.teo == nil {
		p.tg.teo = teo
		go p.tg.start(nil)
		return
	}
	teoroom.RoomRequest(p.tg.teo, p.tg.peer, nil)
}

// Running used inside teocli.Run() function and return running flag
func (p *startCommandD) Running() bool { return p.running }

// Stop used to stop teocli.Run() and set running flag to false
func (p *startCommandD) Stop() {
	// Sleep to get time disconnect from room controller packet go out and
	<-time.After(50 * time.Millisecond)
	p.running = false
	p.tg.teo.Disconnect()
}

// outputCommands teonet output commands receiver
type outputCommands struct {
	tg    *Teogame
	start teocli.StartCommand
}

// disconnect [out] send disconnect (exit from room) command to room controller
func (com *outputCommands) disconnect() {
	if com.tg.rra == nil {
		return
	}
	teoroom.Disconnect(com.tg.teo, com.tg.peer, com.tg.rra.clientID)
}

// sendData [out] send data command to room controller
func (com *outputCommands) sendData(i ...interface{}) (num int, err error) {
	return teoroom.SendData(com.tg.teo, com.tg.peer, i...)
}

// stop teocli.Run()
func (com *outputCommands) stop() {
	com.start.Stop()
}

package main

import (
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// startCommand receiver first outup command
type startCommandD struct {
	tg      *Teogame
	running bool
}

// startCommand command methods
func (p *startCommandD) Command(teo *teocli.TeoLNull) {
	p.tg.teo = teo

	// Send peers command (just for test, it may be removed)
	fmt.Printf("send peers request\n")
	teo.SendTo(p.tg.peer, teocli.CmdLPeers, nil)

	// Send room request
	fmt.Printf("Send room request\n")
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

// startCommand return start command receiver
func startCommand(tg *Teogame) teocli.StartCommand {
	stcom := &startCommandD{tg, true}
	tg.com = &outputCommands{tg: tg, stcom: stcom}
	return stcom
}

// outputCommands teonet output commands receiver
type outputCommands struct {
	tg    *Teogame
	stcom teocli.StartCommand
}

// disconnect [out] send disconnect (exit from room) command to room controller
func (com *outputCommands) disconnect() {
	teoroom.Disconnect(com.tg.teo, com.tg.peer, com.tg.rra.clientID)
}

// sendData [out] send data command to room controller
func (com *outputCommands) sendData(i ...interface{}) (num int, err error) {
	return teoroom.SendData(com.tg.teo, com.tg.peer, i...)
}

// stop teocli.Run()
func (com *outputCommands) stop() {
	com.stcom.Stop()
}

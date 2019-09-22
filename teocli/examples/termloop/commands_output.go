package main

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// startCommand receiver first outup command
type startCommandD struct{ tg *Teogame }

// startCommand command methods
func (p startCommandD) Command(teo *teocli.TeoLNull) {
	p.tg.teo = teo

	// Send peers command (just for test, it may be removed)
	fmt.Printf("send peers request\n")
	teo.SendTo(p.tg.peer, teocli.CmdLPeers, nil)

	// Send room request
	fmt.Printf("Send room request\n")
	teoroom.RoomRequest(p.tg.teo, p.tg.peer, nil)
}

// startCommand return start command receiver
func startCommand(tg *Teogame) teocli.StartCommand {
	tg.com = &outputCommands{tg: tg}
	return startCommandD{tg}
}

// outputCommands teonet output commands receiver
type outputCommands struct {
	tg *Teogame
}

// disconnect [out] send disconnect (exit from room) command to room controller
func (com *outputCommands) disconnect() {
	teoroom.Disconnect(com.tg.teo, com.tg.peer, nil)
}

// sendData [out] send data command to room controller
func (com *outputCommands) sendData(i ...interface{}) (num int, err error) {
	return teoroom.SendData(com.tg.teo, com.tg.peer, i...)
}

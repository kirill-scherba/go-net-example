package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// inputCommand receivers
type echoAnswerCommand struct{}
type peerAnswerCommand struct{}
type roomRequestAnswerCommand struct{ tg *Teogame }
type roomDataCommand struct{ tg *Teogame }

// inputCommands combine input commands to slice
func inputCommands(tg *Teogame) (com []teocli.Command) {
	com = append(com,
		echoAnswerCommand{},
		peerAnswerCommand{},
		roomRequestAnswerCommand{tg},
		roomDataCommand{tg},
	)
	return
}

// echoAnswer command methods
func (p echoAnswerCommand) Cmd() byte { return teocli.CmdLEchoAnswer }
func (p echoAnswerCommand) Command(packet *teocli.Packet) bool {
	if t, err := packet.TripTime(); err != nil {
		fmt.Println("trip time error:", err)
	} else {
		fmt.Println("trip time (ms):", t)
	}
	return true
}

// peerAnswer command methods
func (p peerAnswerCommand) Cmd() byte { return teocli.CmdLPeersAnswer }
func (p peerAnswerCommand) Command(packet *teocli.Packet) bool {
	ln := strings.Repeat("-", 59)
	fmt.Println("PeerAnswer received\n"+ln, "\n"+packet.Peers()+ln)
	return true
}

// roomRequestAnswer command methods
func (p roomRequestAnswerCommand) Cmd() byte { return teoroom.ComRoomRequestAnswer }
func (p roomRequestAnswerCommand) Command(packet *teocli.Packet) bool {
	go p.tg.startGame()
	return true
}

// roomData command methods
func (p roomDataCommand) Cmd() byte { return teoroom.ComRoomData }
func (p roomDataCommand) Command(packet *teocli.Packet) bool {
	name := string(packet.Data()[2*unsafe.Sizeof(int64(0)):])
	p.tg.addPlayer(name).UnmarshalBinary(packet.Data())
	return true
}

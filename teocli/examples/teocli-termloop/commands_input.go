package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// inputCommand receivers
type echoAnswerCommand struct{}
type peerAnswerCommand struct{}
type roomRequestAnswerCommand struct{ tg *Teogame }
type roomDataCommand struct{ tg *Teogame }
type clientDisconnecCommand struct{ tg *Teogame }

// inputCommands combine input commands to slice
func inputCommands(tg *Teogame) (com []teocli.Command) {
	com = append(com,
		echoAnswerCommand{},
		peerAnswerCommand{},
		roomRequestAnswerCommand{tg},
		roomDataCommand{tg},
		clientDisconnecCommand{tg},
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
	rra := roomRequestAnswerData{}
	rra.UnmarshalBinary(packet.Data())
	fmt.Printf("roomRequestAnswerData.UnmarshalBinary after: %v\n", rra)
	go p.tg.startGame(&rra)
	return true
}

// roomData command methods
func (p roomDataCommand) Cmd() byte { return teoroom.ComRoomData }
func (p roomDataCommand) Command(packet *teocli.Packet) bool {
	id := packet.Data()[0]
	p.tg.addPlayer(id).UnmarshalBinary(packet.Data())
	return true
}

// disconnec (exit from room) command methods
func (p clientDisconnecCommand) Cmd() byte { return teoroom.ComDisconnect }
func (p clientDisconnecCommand) Command(packet *teocli.Packet) bool {
	if packet.Data() == nil || len(packet.Data()) == 0 {
		// Game over
		fmt.Printf("Game over!\n")
		os.Exit(0)
		return true
	}
	id := packet.Data()[0]
	if player, ok := p.tg.player[id]; ok {
		p.tg.level.RemoveEntity(player)
		delete(p.tg.player, id)
	}
	return true
}

// roomRequestAnswerData room request answer data structure
type roomRequestAnswerData struct {
	clientID byte
}

// MarshalBinary marshal roomRequestAnswerData to binary
func (rra *roomRequestAnswerData) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, rra.clientID); err != nil {
		return
	}
	data = buf.Bytes()
	return
}

// UnmarshalBinary unmarshal roomRequestAnswerData from binary
func (rra *roomRequestAnswerData) UnmarshalBinary(data []byte) (err error) {
	fmt.Printf("roomRequestAnswerData.UnmarshalBinary data: %v\n", data)
	if data == nil || len(data) == 0 {
		rra.clientID = byte(rand.Intn(20))
		fmt.Printf("roomRequestAnswerData.UnmarshalBinary set: %v\n", rra)
		return
	}
	var x byte
	buf := bytes.NewReader(data)
	if err = binary.Read(buf, binary.LittleEndian, &x); err != nil {
		return
	}
	rra.clientID = x
	return
}

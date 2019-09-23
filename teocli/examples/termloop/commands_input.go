package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unsafe"

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
	go p.tg.startGame(rra)
	return true
}

// roomData command methods
func (p roomDataCommand) Cmd() byte { return teoroom.ComRoomData }
func (p roomDataCommand) Command(packet *teocli.Packet) bool {
	name := string(packet.Data()[2*unsafe.Sizeof(int64(0)):])
	p.tg.addPlayer(name).UnmarshalBinary(packet.Data())
	return true
}

// disconnec (exit from room) command methods
func (p clientDisconnecCommand) Cmd() byte { return teoroom.ComDisconnect }
func (p clientDisconnecCommand) Command(packet *teocli.Packet) bool {
	if packet.Data() == nil || len(packet.Data()) == 0 {
		return true
	}
	name := string(packet.Data())
	if player, ok := p.tg.player[name]; ok {
		p.tg.level.RemoveEntity(player)
		delete(p.tg.player, name)
	}
	return true
}

type roomRequestAnswerData struct {
	clientID byte
}

func (rra *roomRequestAnswerData) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, rra.clientID)
	data = buf.Bytes()
	return
}
func (rra *roomRequestAnswerData) UnmarshalBinary(data []byte) (err error) {
	fmt.Printf("roomRequestAnswerData.UnmarshalBinary data: %v\n", data)
	if data == nil || len(data) == 0 {
		rra.clientID = byte(rand.Intn(20))
		fmt.Printf("roomRequestAnswerData.UnmarshalBinary set: %v\n", rra)
		return
	}
	var x byte
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.LittleEndian, &x)
	rra.clientID = x
	return
}

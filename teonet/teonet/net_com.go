package teonet

//#include "net_com.h"
import "C"
import (
	"fmt"
)

type command struct {
	teo *Teonet
}

// process processed internal Teonet commands
func (com *command) process(rec *receiveData) (processed bool) {
	com.teo.arp.peerNew(rec)
	processed = true
	cmd := rec.rd.Cmd()
	switch cmd {

	case C.CMD_NONE, C.CMD_CONNECT:
		com.connect(rec)

	case C.CMD_DISCONNECTED:
		com.disconnect(rec)

	case C.CMD_ECHO:
		com.echo(rec)

	default:
		processed = false
	}
	return
}

// connect process 'connect' command and answer with 'connect' command
func (com *command) connect(rec *receiveData) {
	rd := rec.rd
	fmt.Printf(">====-- peer %s connected\n", rd.From())
	//com.teo.SendTo(com.teo.name, C.CMD_NONE, []byte{0}) //C.CMD_CONNECT, nil)
	//com.teo.sendToTcd(rec.tcd, C.CMD_CONNECT, nil)
	com.teo.sendToTcd(rec.tcd, 0, []byte{0})
	// com.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO, []byte{0})
	// \TODO send 'connected' event to user level
}

// disconnect process 'disconnect' comman and close trudp channel and delete
// peer from arp table
func (com *command) disconnect(rec *receiveData) {
	rd := rec.rd
	fmt.Printf(">====-- peer %s disconnected\n", rd.From())
	rec.tcd.CloseChannel()
	delete(com.teo.arp.m, rd.From())
	// \TODO send 'disconnected' event to user level
}

// echo process 'echo' command and answer with 'echo answer' command
func (com *command) echo(rec *receiveData) {
	com.teo.sendToTcd(rec.tcd, C.CMD_ECHO_ANSWER, rec.rd.Data())
}

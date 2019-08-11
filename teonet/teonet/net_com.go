package teonet

//#include "net_com.h"
import "C"
import (
	"github.com/kirill-scherba/net-example-go/teolog/teolog"
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
	com.teo.sendToTcd(rec.tcd, 0, []byte{0})
	// com.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO, []byte{0})
	teolog.DebugV(MODULE, "CMD_CONNECT command processed, from:", rd.From())
	// \TODO send 'connected' event to user level
}

// disconnect process 'disconnect' comman and close trudp channel and delete
// peer from arp table
func (com *command) disconnect(rec *receiveData) {
	rd := rec.rd
	rec.tcd.CloseChannel()
	delete(com.teo.arp.m, rd.From())
	teolog.DebugV(MODULE, "CMD_DISCONNECTED command processed, from:", rd.From())
	// \TODO send 'disconnected' event to user level
}

// echo process 'echo' command and answer with 'echo answer' command
func (com *command) echo(rec *receiveData) {
	com.teo.sendToTcd(rec.tcd, C.CMD_ECHO_ANSWER, rec.rd.Data())
	teolog.DebugV(MODULE, "CMD_ECHO command processed, from:", rec.rd.From())
}

package teonet

//#include "net_com.h"
import "C"
import "fmt"

type command struct {
	teo *Teonet
}

// process processed internal Teonet commands
func (com *command) process(rec *receiveData) (processed bool) {
	processed = true
	cmd := rec.rd.Cmd()
	switch cmd {
	case C.CMD_CONNECT:
		com.connect(rec)
	case C.CMD_ECHO:
		com.echo(rec)
	default:
		processed = false
	}
	return
}

// echo process 'connect' command and answer with 'connect' command
func (com *command) connect(rec *receiveData) {
	rd := rec.rd
	fmt.Printf(">====-- peer %s connected\n", rd.From())
	//com.teo.SendTo(com.teo.name, C.CMD_NONE, []byte{0}) //C.CMD_CONNECT, nil)
	com.teo.SendTo(rd.From(), C.CMD_CONNECT, nil)
}

// echo process 'echo' command and answer with 'echo answer' command
func (com *command) echo(rec *receiveData) {
	rd := rec.rd
	com.teo.SendTo(rd.From(), C.CMD_ECHO_ANSWER, rd.Data())
}

package teonet

//#include "net_com.h"
import "C"
import (
	"errors"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// Teonet commands
const (
	CmdHostInfo = C.CMD_HOST_INFO
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

	case C.CMD_ECHO_ANSWER:
		com.echoAnswer(rec)

	case C.CMD_HOST_INFO:
		com.hostInfo(rec)

	case C.CMD_HOST_INFO_ANSWER:
		com.hostInfoAnswer(rec)

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
	com.teo.arp.delete(rec)
	teolog.DebugV(MODULE, "CMD_DISCONNECTED command processed, from:", rd.From())
	// \TODO send 'disconnected' event to user level
}

// echo process 'echo' command and answer with 'echo answer' command
func (com *command) echo(rec *receiveData) {
	com.teo.sendToTcd(rec.tcd, C.CMD_ECHO_ANSWER, rec.rd.Data())
	teolog.DebugV(MODULE, "CMD_ECHO command processed, from:", rec.rd.From())
}

// echo process 'echoAnswer' command
func (com *command) echoAnswer(rec *receiveData) {
	teolog.Debug(MODULE, "CMD_ECHO_ANSWER command processed, from:", rec.rd.From())
}

// hostInfo process 'hostInfo' command and send host info to peer from
func (com *command) hostInfo(rec *receiveData) (err error) {
	var data []byte

	// Select this host in arp table
	peerArp, ok := com.teo.arp.m[com.teo.param.Name]
	if !ok {
		err = errors.New("host " + com.teo.param.Name + " does not exist in arp table")
		teolog.Error(MODULE, "CMD_HOST_INFO command processed, from:", rec.rd.From())
		return
	}

	// Version
	ver := strings.Split(com.teo.version(), ".")
	for i := 0; i < 3; i++ {
		v, _ := strconv.Atoi(ver[i])
		data = append(data, byte(v))
	}

	// String array length
	stringArLen := len(peerArp.appType)
	data = append(data, byte(stringArLen+1))

	// Name
	data = append(data, append([]byte(com.teo.param.Name), 0)...)

	// Types
	for i := 0; i < stringArLen; i++ {
		data = append(data, append([]byte(peerArp.appType[i]), 0)...)
	}

	// Send answer with host infor data
	com.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO_ANSWER, data)

	teolog.Debug(MODULE, "CMD_HOST_INFO command processed, from:", rec.rd.From(), data)
	return
}

// hostInfoAnswer process 'hostInfoAnswer' command and add host info to the arp table
func (com *command) hostInfoAnswer(rec *receiveData) (err error) {
	var stringAr []string
	data := rec.rd.Data()
	var version string
	//teolog.Debugf(MODULE, "got CMD_HOST_INFO_ANSWER, data_lenngth: %d, data: %v\n", rec.rd.DataLen(), data)

	// Version
	version = strconv.Itoa(int(data[0])) + "." + strconv.Itoa(int(data[1])) + "." + strconv.Itoa(int(data[2]))

	// String array length
	stringArLen := int(data[3])

	// String array
	ptr := 4
	for i := 0; i < stringArLen; i++ {
		charPtr := unsafe.Pointer(&data[ptr])
		stringAr = append(stringAr, C.GoString((*C.char)(charPtr)))
		ptr += len(stringAr[i]) + 1
	}

	// Save to arp Table
	peerArp, ok := com.teo.arp.m[rec.rd.From()]
	if !ok {
		err = errors.New("peer " + rec.rd.From() + " does not exist in arp table")
		teolog.Error(MODULE, "CMD_HOST_INFO_ANSWER command processed, from:", rec.rd.From())
		return
	}
	peerArp.version = version
	peerArp.appType = stringAr[1:]
	com.teo.arp.print()

	//teolog.Debugf(MODULE, "version: %s, string_ar_num: %d, string_ar: %v\n", version, stringArLen, stringAr)
	teolog.Debug(MODULE, "CMD_HOST_INFO_ANSWER command processed, from:", rec.rd.From())
	return
}

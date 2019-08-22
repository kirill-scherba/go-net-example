package teonet

//#include "command.h"
//#include "packet.h"
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// Teonet commands
const (
	CmdNone       = C.CMD_NONE         // Cmd none used as first peers command
	CmdConnectR   = C.CMD_CONNECT_R    // A Peer want connect to r-host
	CmdConnect    = C.CMD_CONNECT      // Inform peer about connected peer
	CmdDisconnect = C.CMD_DISCONNECTED // Send to peers signal about disconnect
	CmdHostInfo   = C.CMD_HOST_INFO    // Request host info, allow JSON in request
)

// JSON data prefix used in teonet requests
var JSON = []byte("JSON")

// BINARY data prefix used in teonet requests
var BINARY = []byte("BINARY")

// command commands module methods holder
type command struct {
	teo *Teonet
}

// process processed internal Teonet commands
func (com *command) process(rec *receiveData) (processed bool) {
	com.teo.arp.peerNew(rec)
	processed = true
	cmd := rec.rd.Cmd()

	// Process kernel commands
	switch cmd {

	case C.CMD_CONNECT_R:
		com.teo.rhost.cmdConnectR(rec)

	case C.CMD_NONE, C.CMD_CONNECT:
		com.connect(rec, cmd)

	case C.CMD_DISCONNECTED:
		com.disconnect(rec)

	case C.CMD_RESET:
		com.reset(rec)

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

	// Process waitFrom commands
	com.teo.wcom.check(rec)

	return
}

// log command processed log message
func (com *command) log(rd *C.ksnCorePacketData, descr string) {
	teolog.DebugVfd(1, MODULE, "got cmd: %d, from: %s, data_len: %d (%s)",
		rd.Cmd(), rd.From(), rd.DataLen(), descr)
}

// error command processed with error log message
func (com *command) error(rd *C.ksnCorePacketData, descr string) {
	teolog.Errorfd(1, MODULE, "got cmd: %d, from: %s, data_len: %d (%s)",
		rd.Cmd(), rd.From(), rd.DataLen(), descr)
}

// connect process 'connect' command and answer with 'connect' command
func (com *command) connect(rec *receiveData, cmd int) {
	if cmd == C.CMD_CONNECT {
		var to string
		if rec.rd != nil && rec.rd.Data() != nil {
			peer, addr, port, err := com.teo.rhost.cmdConnectData(rec)
			if err == nil {
				to = fmt.Sprintf("%s %s:%d", peer, addr, port)
			}
		}
		com.log(rec.rd, "CMD_CONNECT command: "+to)
		com.teo.rhost.cmdConnect(rec)
	} else {
		com.log(rec.rd, "CMD_NONE command")
		//com.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO, []byte{0})
		//com.teo.sendToTcd(rec.tcd, C.CMD_NONE, []byte{0})
	}
	// \TODO ??? send 'connected' event to user level
}

// disconnect process 'disconnect' command and close trudp channel and delete
// peer from arp table
func (com *command) disconnect(rec *receiveData) {
	com.log(rec.rd, fmt.Sprint("CMD_DISCONNECTED command ", rec.rd.Data(), string(rec.rd.Data())))
	com.teo.arp.delete(rec)
	// \TODO send 'disconnected' event to user level
}

// reset process 'reset' command data: <t byte>
//   t = 0 - soft reset
//   t = 1 - hard reset
func (com *command) reset(rec *receiveData) {
	com.log(rec.rd, "CMD_RESET command")
	if rec.rd.DataLen() > 0 {
		b := rec.rd.Data()[0]
		if b == 1 || b == '1' {
			com.teo.Reconnect()
		}
	}
}

// echo process 'echo' command and answer with 'echo answer' command
func (com *command) echo(rec *receiveData) {
	com.log(rec.rd, "CMD_ECHO command, data: "+
		C.GoString((*C.char)(unsafe.Pointer(&rec.rd.Data()[0]))))
	com.teo.sendToTcd(rec.tcd, C.CMD_ECHO_ANSWER, rec.rd.Data())
}

// echo process 'echoAnswer' command
func (com *command) echoAnswer(rec *receiveData) {
	com.log(rec.rd, "CMD_ECHO_ANSWER command, data: "+
		C.GoString((*C.char)(unsafe.Pointer(&rec.rd.Data()[0]))))
}

// hostInfo is th host info json data structure
type hostInfo struct {
	Name        string   `json:"name"`
	Type        []string `json:"type"`
	AppType     []string `json:"appType"`
	Version     string   `json:"version"`
	AppVersion1 string   `json:"app_version"`
	AppVersion2 string   `json:"appVersion"`
}

// hostInfo process 'hostInfo' command and send host info to peer from
func (com *command) hostInfo(rec *receiveData) (err error) {
	var data []byte

	// Select this host in arp table
	peerArp, ok := com.teo.arp.m[com.teo.param.Name]
	if !ok {
		err = errors.New("host " + com.teo.param.Name + " does not exist in arp table")
		com.error(rec.rd, "CMD_HOST_INFO command processed with error: "+err.Error())
		return
	}
	com.log(rec.rd, "CMD_HOST_INFO command")

	// This func convert string Version to byte array
	ver := func(version string) (data []byte) {
		ver := strings.Split(com.teo.Version(), ".")
		for _, vstr := range ver {
			v, _ := strconv.Atoi(vstr)
			data = append(data, byte(v))
		}
		return
	}

	// Create Json or bynary answer depend of input data: JSON - than answer in json
	if l := len(JSON); rec.rd.DataLen() >= l && bytes.Equal(rec.rd.Data()[:l], JSON) {
		data, _ = json.Marshal(hostInfo{com.teo.param.Name, peerArp.appType,
			peerArp.appType, com.teo.Version(), peerArp.appVersion, peerArp.appVersion})
		data = append(data, 0) // add trailing zero (cstring)
	} else {
		typeArLen := len(peerArp.appType)
		name := com.teo.param.Name
		data = ver(com.teo.Version())                   // Version
		data = append(data, byte(typeArLen+1))          // Types array length
		data = append(data, append([]byte(name), 0)...) // Name
		for i := 0; i < typeArLen; i++ {                // Types array
			data = append(data, append([]byte(peerArp.appType[i]), 0)...)
		}
	}

	// Send answer with host infor data
	com.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO_ANSWER, data)

	return
}

// hostInfoAnswer process 'hostInfoAnswer' command and add host info to the arp table
func (com *command) hostInfoAnswer(rec *receiveData) (err error) {
	data := rec.rd.Data()
	var typeAr []string
	var version string

	// Parse json or binary format depend of data.
	// If first char = '{' and last char = '}' than data is in json
	if l := len(data); l > 3 && data[0] == '{' && data[l-2] == '}' && data[l-1] == 0 {
		var j hostInfo
		json.Unmarshal(data, &j)
		version = j.Version
		typeAr = append([]string{j.Name}, j.Type...)
	} else {
		version = strconv.Itoa(int(data[0])) + "." + strconv.Itoa(int(data[1])) + "." + strconv.Itoa(int(data[2]))
		typeArLen := int(data[3])
		ptr := 4
		for i := 0; i < typeArLen; i++ {
			charPtr := unsafe.Pointer(&data[ptr])
			typeAr = append(typeAr, C.GoString((*C.char)(charPtr)))
			ptr += len(typeAr[i]) + 1
		}
	}

	// Save to arp Table
	peerArp, ok := com.teo.arp.m[rec.rd.From()]
	if !ok {
		err = errors.New("peer " + rec.rd.From() + " does not exist in arp table")
		com.error(rec.rd, "CMD_HOST_INFO_ANSWER command processed with error: "+err.Error())
		return
	}
	com.log(rec.rd, "CMD_HOST_INFO_ANSWER command")
	peerArp.version = version
	peerArp.appType = typeAr[1:]
	com.teo.arp.print()

	return
}

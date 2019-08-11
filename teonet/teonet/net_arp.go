package teonet

import (
	"fmt"
	"strings"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

type arpData struct {
	peer string
	mode int
	tcd  *trudp.ChannelData
}

type arp struct {
	teo *Teonet
	m   map[string]*arpData // arp map
}

// peerAdd create new peer in art table map without TCD. Used to create record
// for this host only.
func (arp *arp) peerAdd(peer string) (peerArp *arpData) {
	peerArp, ok := arp.m[peer]
	if ok {
		return
	}
	peerArp = &arpData{peer: peer, mode: -1}
	arp.m[peer] = peerArp
	arp.print()
	return
}

// newPeer create new peer in art table map or select existing
func (arp *arp) peerNew(rec *receiveData) (peerArp *arpData) {
	peer := rec.rd.From()
	peerArp, ok := arp.m[peer]
	if ok {
		//trudp.Log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}
	peerArp = &arpData{peer: peer, tcd: rec.tcd}
	// arp.teo.sendToTcd(rec.tcd, 0, []byte{0})
	arp.teo.sendToTcd(rec.tcd, CmdHostInfo, []byte{0})
	arp.m[peer] = peerArp
	arp.print()
	return
}

// delete remove peer from arp table and close trudp channel (by receiveData)
func (arp *arp) delete(rec *receiveData) (peerArp *arpData) {
	peer := rec.rd.From()
	peerArp, ok := arp.m[peer]
	if !ok {
		return
	}
	if peerArp.tcd != nil {
		peerArp.tcd.CloseChannel()
	}
	delete(arp.m, peer)
	arp.print()
	return
}

// delete remove peer from arp table /*and close trudp channel*/ (by trudp channel key)
func (arp *arp) deleteKey(key string) (peerArp *arpData) {
	for peer, peerArp := range arp.m {
		if peerArp.tcd != nil && peerArp.tcd.GetKey() == key {
			peerArp.tcd.CloseChannel()
			delete(arp.m, peer)
			arp.print()
			break
		}
	}
	return
}

// sprint print teonet arp table
func (arp *arp) print() {
	if arp.teo.param.ShowPeersStatF {
		fmt.Print(arp.sprint())
	}
}

// sprint return teonet arp table string
func (arp *arp) sprint() (str string) {
	var div = "\033[2K" + strings.Repeat("-", 83) + "\n"
	str = div +
		"  # Peer          | Mod | IP              | Port |  Trip time | TR-UDP trip time\n" +
		div
	num := 0
	for peer, peerArp := range arp.m {
		num++
		var ip string
		var port int
		if peerArp.mode == -1 {
			// \TODO get connected IP and Port
			port = arp.teo.param.Port
		} else {
			addr := peerArp.tcd.GetAddr()
			ip = addr.IP.String()
			port = addr.Port
		}
		str += fmt.Sprintf("%3d %s%-15s%s %3d   %-15s  %5d   %7s %s  %s%s%s\n",
			num, // num
			teokeys.ANSIGreen,
			peer, // peer name,
			teokeys.ANSINone,
			peerArp.mode, // mod
			ip,           // ip
			port,         // port
			"", "",       // triptime
			"", "", "", // trudp triptime
		)
	}
	str += div

	return
}

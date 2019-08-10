package teonet

import (
	"fmt"
	"strings"

	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

type arpData struct {
	peer string
	tcd  *trudp.ChannelData
}

type arp struct {
	teo *Teonet
	m   map[string]*arpData // arp map
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
	// arp.teo.sendToTcd(rec.tcd, C.CMD_HOST_INFO, []byte{0})
	arp.m[peer] = peerArp
	arp.print()
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
	for peer, _ := range arp.m {
		num++
		str += fmt.Sprintf("%3d %s%-15s%s %3d   %-15s  %5d   %7s %s  %s%s%s\n",
			num, // num
			"",
			peer, // peer name,
			"",
			0,      // mod
			"",     // ip
			0,      // port
			"", "", // triptime
			"", "", "", // trudp triptime
		)
	}
	str += div

	return
}

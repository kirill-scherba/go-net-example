package teonet

import "github.com/kirill-scherba/net-example-go/trudp/trudp"

type arpData struct {
	peer string
	tcd  *trudp.ChannelData
}

type arp struct {
	//teo *Teonet
	m map[string]*arpData // arp map
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
	return
}

package teonet

type arpData struct {
	peer string
	addr string
	port int
	ch   int
}

type arp struct {
	//teo *Teonet
	m map[string]*arpData // arp map
}

// newPeer create new peer in art table map or select existing
func (arp *arp) peerNew(addr string, port int, rec *receiveData) (peerArp *arpData) {
	rd := rec.rd
	peer := rd.From()
	peerArp, ok := arp.m[peer]
	if ok {
		//trudp.Log(DEBUGvv, "the ChannelData with key", key, "selected")
		return
	}
	peerArp = &arpData{peer: peer, addr: addr, port: port, ch: 0}
	arp.m[peer] = peerArp
	return
}

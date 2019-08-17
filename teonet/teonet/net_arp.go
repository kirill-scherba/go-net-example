package teonet

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

type arpData struct {
	peer    string             // peer name
	mode    int                // mode (-1 - this host; 1 - r-host; 0 - all other host)
	version string             // teonet version
	appType []string           // application types array
	tcd     *trudp.ChannelData // trudp channel connection
}

type arp struct {
	teo *Teonet
	m   map[string]*arpData // arp map
}

// peerAdd create new peer in art table map without TCD. Used to create record
// for this host only.
func (arp *arp) peerAdd(peer, version string) (peerArp *arpData) {
	peerArp, ok := arp.m[peer]
	if ok {
		return
	}
	peerArp = &arpData{peer: peer, mode: -1, version: version}
	arp.m[peer] = peerArp
	arp.print()
	return
}

// newPeer create new peer in art table map or select existing
func (arp *arp) peerNew(rec *receiveData) (peerArp *arpData) {
	peer := rec.rd.From()
	peerArp, ok := arp.m[peer]
	if ok {
		// Close discovery channel (if there is new channel with same teone name)
		// \TODO: Check if this channel really closed (issue #15)
		if rec.tcd != rec.tcd {
			rec.tcd.CloseChannel()
		}
		return
	}
	peerArp = &arpData{peer: peer, tcd: rec.tcd}
	if arp.teo.rhost.isrhost(rec.tcd) {
		peerArp.mode = 1
	}
	arp.m[peer] = peerArp
	arp.print()
	arp.teo.sendToTcd(rec.tcd, CmdHostInfo, []byte{0})
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

// peer return peer name (find by tcd)
func (arp *arp) peer(tcd *trudp.ChannelData) (string, error) {
	for peer, peerArp := range arp.m {
		if peerArp.tcd != nil && peerArp.tcd == tcd {
			return peer, nil
		}
	}
	return "unknown", errors.New("not found")
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

// deleteAll remove all peers from arp table
func (arp *arp) deleteAll() {
	for peer, arpData := range arp.m {
		if arpData.tcd != nil {
			if arpData.mode == 1 {
				arp.teo.rhost.reconnect(arpData.tcd)
			}
			if arpData.mode != -1 {
				arp.teo.sendToTcdUnsafe(arpData.tcd, CmdDisconnect, nil)
				fmt.Printf("send disconnect to %s\n", arpData.peer)
				arpData.tcd.CloseChannel()
			}
		}
		delete(arp.m, peer)
		arp.print()
	}
}

// sprint print teonet arp table
func (arp *arp) print() {
	if arp.teo.param.ShowPeersStatF {
		fmt.Print(arp.sprint())
	}
}

// sprint return teonet arp table string
func (arp *arp) sprint() (str string) {

	var num = 0              // number of body line
	const numadd = 6         // add lines to scroll aria
	const clearl = "\033[2K" // clear line terminal code
	var line = clearl + strings.Repeat("-", 80) + "\n"

	// Sort peers table:
	// read peers arp map keys to slice and sort it
	keys := make([]string, len(arp.m))
	for key := range arp.m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Header
	str = "\0337" + // save cursor position
		"\033[0;0H" + // set cursor to the top
		//"\033[?7l" + // does not wrap
		line +
		clearl + "  # Peer          | Mod | Version | IP              |  Port | Triptime / midle\n" +
		line

	// Body
	for _, peer := range keys {
		peerArp, ok := arp.m[peer]
		if !ok {
			continue
		}
		num++
		var port int
		var ip string
		var triptime, triptimeMidle float32
		if peerArp.mode == -1 {
			// \TODO get connected IP and Port
			port = arp.teo.param.Port
		} else {
			triptime, triptimeMidle = peerArp.tcd.GetTriptime()
			addr := peerArp.tcd.GetAddr()
			ip = addr.IP.String()
			port = addr.Port
		}
		str += fmt.Sprintf(clearl+"%3d %s%-15s%s %3d %9s   %-15s %7d   %8.3f / %-8.3f\n",
			num,               // num
			teokeys.ANSIGreen, // (color begin)
			peer,              // peer name
			teokeys.ANSINone,  // (color end)
			peerArp.mode,      // mod
			peerArp.version,   // teonet version
			ip,                // ip
			port,              // port
			triptime,          // triptime
			triptimeMidle,     // triptime midle
		)
	}

	// Footer
	str += line + fmt.Sprintf(""+
		clearl+"\n"+ // clear line
		//clearl+"\n"+ // clear line
		"\033[%d;r"+ // setscroll mode
		"\0338", // restore cursor position
		num+numadd,
	)

	return
}

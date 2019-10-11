// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet arp module.

package teonet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

// arpData arp map record data structure
type arpData struct {
	peer       string             // peer name
	mode       int                // mode (-1 - this host; 1 - r-host; 0 - all other host)
	version    string             // teonet version
	appVersion string             // application version
	appType    []string           // application types array
	tcd        *trudp.ChannelData // trudp channel connection
}

// arp teonet module structure
type arp struct {
	teo *Teonet             // ponter to Teonet
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

	//peerArp, ok := arp.m[peer]
	var ok bool
	if peerArp, ok = arp.find(peer); ok {
		if rec.tcd != peerArp.tcd {
			if peerArp.tcd != nil {
				teolog.DebugVf(MODULE, "the peer %s is already connected at channel %s, "+
					"now it try connect at channel %s\n",
					peer, peerArp.tcd.GetKey(), rec.tcd.GetKey())
			}
			rec.tcd.Close()
		}
		return
	}

	if peerArp, ok = arp.find(rec); ok {
		teolog.DebugVf(MODULE, "the connection %s already associated with peer %s",
			rec.tcd.GetKey(), peer)
		return
	}

	peerArp = &arpData{peer: peer, tcd: rec.tcd}
	if arp.teo.rhost.isrhost(rec.tcd) {
		peerArp.mode = 1
	}
	arp.m[peer] = peerArp
	arp.print()
	arp.teo.sendToTcd(rec.tcd, CmdNone, []byte{0})
	arp.teo.sendToTcd(rec.tcd, CmdHostInfo, []byte{0})
	go func() {
		r := <-arp.teo.WaitFrom(peer, CmdHostInfoAnswer)
		if r.Err == nil {
			arp.teo.ev.send(EventConnected, arp.teo.packetCreateNew(peer, 0, nil))
		}
	}()
	return
}

// find finds peer in teonet peer arp table
// function uses diferent tarameters:
//  - find by peer name: <peer string>
//  - find by tcd: <tcd *trudp.ChannelData>
//  - find by addr, port (channel = 0): <addr string, port int>
//  - find by addr, port and channel: <addr string, port int, channel int>
func (arp *arp) find(i ...interface{}) (peerArp *arpData, ok bool) {
	switch len(i) {
	case 1:
		switch p := i[0].(type) {

		// Find by peer name
		case string:
			peerArp, ok = arp.m[p]
			return

		// Find by tcd
		case *trudp.ChannelData:
			for _, peerArp = range arp.m {
				if peerArp.tcd != nil && peerArp.tcd == p {
					ok = true
					return
				}
			}
		}

	// Find by address and port and channel (may be ommited)
	case 2, 3:
		var addr = i[0].(string)
		var port = i[1].(int)
		var ch = 0
		if len(i) == 3 {
			ch = i[2].(int)
		}
		for _, peerArp = range arp.m {
			if peerArp.tcd != nil &&
				peerArp.tcd.GetAddr().IP.String() == addr &&
				peerArp.tcd.GetAddr().Port == port &&
				peerArp.tcd.GetCh() == ch {
				ok = true
				return
			}
		}
	}
	return
}

// deletePeer remove peer from arp table
func (arp *arp) deletePeer(peer string) {
	if peerArp, ok := arp.m[peer]; ok {
		if peerArp.mode != -1 {
			arp.teo.ev.send(EventDisconnected, arp.teo.packetCreateNew(peer, 0, nil))
		}
		delete(arp.m, peer)
		arp.print()
	}
}

// delete remove peer from arp table and close trudp channel (by receiveData)
func (arp *arp) delete(rec *receiveData) (peerArp *arpData) {
	peer := rec.rd.From()
	peerArp, ok := arp.m[peer]
	if !ok {
		return
	}
	if peerArp.tcd != nil {
		peerArp.tcd.Close()
	}
	arp.deletePeer(peer)
	return
}

// peer return peer name (find by tcd)
func (arp *arp) peer(tcd *trudp.ChannelData) (string, error) {
	for peer, peerArp := range arp.m {
		if peerArp.tcd == tcd {
			return peer, nil
		}
	}
	return "unknown", errors.New("not found")
}

// delete remove peer from arp table /*and close trudp channel*/ (by trudp channel key)
func (arp *arp) deleteKey(key string) (peerArp *arpData) {
	for peer, peerArp := range arp.m {
		if peerArp.tcd != nil && peerArp.tcd.GetKey() == key {
			peerArp.tcd.Close()
			arp.deletePeer(peer)
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
				arp.teo.rhost.stop(arpData.tcd)
			}
			if arpData.mode != -1 {
				teolog.DebugVvf(MODULE, "send disconnect to %s\n", arpData.peer)
				// \TODO: Very strange!!! Teont C applications send disconnect without
				// data. If we send disconect withou data it dose not processed correctly.
				// --- It works correctly if packet enctryption enable
				//arp.teo.sendToTcdUnsafe(arpData.tcd, CmdDisconnect, arp.teo.Host())
				if arp.teo.param.DisallowEncrypt {
					arp.teo.sendToTcdUnsafe(arpData.tcd, CmdDisconnect, []byte{0})
				} else {
					arp.teo.sendToTcdUnsafe(arpData.tcd, CmdDisconnect, nil)
				}
				arpData.tcd.Close()
			}
		}
		// delete(arp.m, peer)
		// arp.print()
		arp.deletePeer(peer)
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

	// Header
	str = "\0337" + // save cursor position
		"\033[0;0H" + // set cursor to the top
		//"\033[?7l" + // does not wrap
		line +
		clearl + "  # Peer          | Mod | Version | IP              |  Port | Triptime / midle\n" +
		line

	// Body
	keys := arp.sort()
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
		"\033[%d;r"+ // set scroll mode
		"\0338", // restore cursor position
		num+numadd,
	)

	return
}

// sort Sorts peers table
func (arp *arp) sort() (keys []string) {
	for key := range arp.m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}

// binary creates binary peers array
func (arp *arp) binary() (peersDataAr []byte, peersDataArLen int) {
	// Sort peers table and create binary peer buffer
	keys := arp.sort()
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(keys))) // Number of peers
	for _, peer := range keys {
		peerArp, ok := arp.m[peer]
		if !ok {
			continue
		}
		var port int
		var addr string
		var triptime float32
		if peerArp.tcd != nil {
			addr = peerArp.tcd.GetAddr().IP.String()
			port = peerArp.tcd.GetAddr().Port
			triptime = peerArp.tcd.TripTime()
		} else {
			addr = localhostIP
			port = arp.teo.param.Port
		}
		binary.Write(buf, binary.LittleEndian, teocli.PeerData(peerArp.mode, peer,
			addr, port, triptime),
		)
		peersDataArLen++
	}
	peersDataAr = buf.Bytes()
	return
}

type peersDataJSON struct {
	Name     string  `json:"name"`
	Mode     int     `json:"mode"`
	Addr     string  `json:"addr"`
	Port     int     `json:"port"`
	Triptime float32 `json:"triptime"`
	Uptime   float32 `json:"uptime"`
}

type peersDataArJSON struct {
	Length  int             `json:"length"`
	PeersAr []peersDataJSON `json:"arp_data_ar"`
}

// binaryToJSON convert binary peers array to JSON format
func (arp *arp) binaryToJSON(indata []byte) (data []byte, peersDataArLen int) {
	peersDataAr := peersDataArJSON{}
	buf := bytes.NewReader(indata)
	le := binary.LittleEndian
	var numOfPeers uint32
	binary.Read(buf, le, &numOfPeers) // Number of peers
	for i := 0; i < int(numOfPeers); i++ {
		peerData := make([]byte, teocli.PeerDataLength())
		binary.Read(buf, le, peerData)
		var peersData peersDataJSON
		peersData.Mode, peersData.Name, peersData.Addr, peersData.Port, peersData.Triptime = teocli.ParsePeerData(peerData)
		peersDataAr.PeersAr = append(peersDataAr.PeersAr, peersData)
	}
	peersDataAr.Length = int(numOfPeers)
	peersDataArLen = int(numOfPeers)
	var err error
	data, err = json.Marshal(peersDataAr)
	if err != nil {
		//
	}
	//data = append(data, 0) // add trailing zero (cstring)
	//fmt.Printf("binaryToJSON: %s\n", string(data))
	return
}

// json creates json peers array
func (arp *arp) json() (data []byte, peersDataArLen int) {

	peersDataAr := peersDataArJSON{}

	// Sort peers table and create binary peer buffer
	keys := arp.sort()
	peersDataAr.Length = len(keys)
	peersDataArLen = peersDataAr.Length
	for _, peer := range keys {
		peerArp, ok := arp.m[peer]
		if !ok {
			continue
		}
		var peersData peersDataJSON
		peersData.Name = peer
		peersData.Mode = peerArp.mode
		if peerArp.tcd != nil {
			peersData.Addr = peerArp.tcd.GetAddr().IP.String()
			peersData.Port = peerArp.tcd.GetAddr().Port
			peersData.Triptime = peerArp.tcd.TripTime()
		} else {
			peersData.Addr = localhostIP
			peersData.Port = arp.teo.param.Port
		}
		peersDataAr.PeersAr = append(peersDataAr.PeersAr, peersData)
	}
	var err error
	data, err = json.Marshal(peersDataAr)
	if err != nil {
		//
	}
	//data = append(data, 0) // add trailing zero (cstring)

	return
}

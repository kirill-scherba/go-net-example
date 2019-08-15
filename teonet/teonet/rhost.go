package teonet

//#include <stdint.h>
//#include <string.h>
//#include "net_com.h"
/*
uint32_t getPort(void *data, size_t data_len) {
  return *((uint32_t*)(data + data_len - sizeof(uint32_t)));
}
void setPort(void *data, size_t ptr, uint32_t port) {
  *((uint32_t *)(data + ptr)) = port;
}
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"sync"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// rhostData r-host data
type rhostData struct {
	teo *Teonet            // Teonet connection
	tcd *trudp.ChannelData // TRUDP channel data
	wg  sync.WaitGroup     // Reconnect wait group
}

// connect process command CMD_CONNECT_R - a peer want connect to r-host
func (rhost *rhostData) connect(rec *receiveData) {

	// Replay to address we got from peer
	//ksnCoreSendto(kco->kc, rd->addr, rd->port, CMD_NONE, NULL_STR, 1);
	rhost.teo.sendToTcd(rec.tcd, C.CMD_NONE, []byte{0})

	ptr := 1              // pointer to first IP
	from := rec.rd.From() // from
	data := rec.rd.Data() // received data
	numIP := data[0]      // number of received IPs
	port := int(C.getPort(unsafe.Pointer(&data[0]), C.size_t(len(data))))

	// Create data to resend to peers
	makeData := func(from, addr string, port int) []byte {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, []byte(from))
		binary.Write(buf, binary.LittleEndian, byte(0))
		binary.Write(buf, binary.LittleEndian, []byte(addr))
		binary.Write(buf, binary.LittleEndian, byte(0))
		binary.Write(buf, binary.LittleEndian, C.uint32_t(port))
		return buf.Bytes()
	}

	//fmt.Printf("from: %s\nport: %d\nnumber of IPs: %d\n", rec.rd.From(), port, numIP)
	// Send received IPs to this peer child(connected peers)
	for i := 0; i <= int(numIP); i++ {
		var caddr *C.char
		if i == 0 {
			clocalhost := append([]byte("127.0.0.1"), 0)
			caddr = (*C.char)(unsafe.Pointer(&clocalhost[0]))
		} else {
			caddr = (*C.char)(unsafe.Pointer(&data[ptr]))
			ptr += int(C.strlen(caddr)) + 1
		}
		addr := C.GoString(caddr)

		// Send connected(who send this command) peer local IP address and port to all this host child
		for peer, arp := range rhost.teo.arp.m {
			if arp.mode != -1 && peer != from {
				//fmt.Printf("send %s %s:%d ==> to peer: %s\n", from, addr, port, peer)
				rhost.teo.SendTo(peer, C.CMD_CONNECT, makeData(from, addr, port))
			}
		}
	}

	// Send connected(who send this command) peer IP address and port(defined by this host) to all this host child
	for peer, arp := range rhost.teo.arp.m {
		if arp.mode != -1 && peer != from {
			//fmt.Printf("send %s %s:%d ==> to peer: %s\n", from, rec.tcd.GetAddr().IP.String(), rec.tcd.GetAddr().Port, peer)
			rhost.teo.SendTo(peer, C.CMD_CONNECT, makeData(from, rec.tcd.GetAddr().IP.String(), rec.tcd.GetAddr().Port))
		}
	}

	// Send all child IP address and port to connected(who send this command) peer
	for peer, arp := range rhost.teo.arp.m {
		if arp.mode != -1 && peer != from {
			//fmt.Printf("send %s %s:%d ==> to peer: %s\n", peer, arp.tcd.GetAddr().IP.String(), arp.tcd.GetAddr().Port, from)
			rhost.teo.sendToTcd(rec.tcd, C.CMD_CONNECT, makeData(peer, arp.tcd.GetAddr().IP.String(), arp.tcd.GetAddr().Port))
		}
	}

	teolog.DebugV(MODULE, "CMD_CONNECT_R command processed, from:", rec.rd.From())
}

// reconnect reconnect to r-host if selected in function parameters channel is
// r-host trudp channel
func (rhost *rhostData) reconnect(tcd *trudp.ChannelData) {
	if rhost.tcd == tcd {
		rhost.wg.Done()
	}
}

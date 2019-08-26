package teonet

//#include <stdint.h>
//#include <string.h>
//#include "command.h"
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
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
	"github.com/kirill-scherba/net-example-go/trudp/trudp"
)

// rhostData r-host data
type rhostData struct {
	teo     *Teonet            // Teonet connection
	tcd     *trudp.ChannelData // TRUDP channel data
	running bool               // R-Host module running (connected or try to connect)
	wg      sync.WaitGroup     // Reconnect wait group
}

// cmdConnectData parse cmd connect data
func (rhost *rhostData) cmdConnectData(rec *receiveData) (peer, addr string, port int, err error) {

	// Check data
	if rec == nil || rec.rd == nil || rec.rd.Data() == nil {
		err = errors.New("data not present")
		return
	}

	// Parse data
	data := rec.rd.Data()
	var portu uint32

	// \TODO: fix it!!! From time to time teonet-go application send this data
	// with 255, 255 at the end. This code replace it to 0, 0. Neet to understand
	// what going on and fix it!
	l := len(data)
	if data[l-1] == 255 {
		data[l-1] = 0
	}
	if data[l-2] == 255 {
		data[l-2] = 0
	}

	buf := bytes.NewBuffer(data)
	peer, _ = buf.ReadString(0)
	addr, _ = buf.ReadString(0)
	binary.Read(buf, binary.LittleEndian, &portu)
	peer = strings.TrimSuffix(peer, "\x00") // remove leading 0
	addr = strings.TrimSuffix(addr, "\x00") // remove leading 0
	port = int(portu)
	//fmt.Printf("%v\npeer: %s, addr: %s, port: %d\n", data, peer, addr, port)
	return
}

// cmdConnect process command CMD_CONNECT received from r-host
// command data structure: <peer *C.char> <addr *C.char> <port uint32>
func (rhost *rhostData) cmdConnect(rec *receiveData) {

	// Parse cmd connect data
	peer, addr, port, err := rhost.cmdConnectData(rec)
	if err != nil {
		return
	}

	// Does not process this command if peer already connected
	if _, ok := rhost.teo.arp.find(peer); ok {
		teolog.DebugVv(MODULE, "peer", peer, "already connected, suggests address",
			addr, "port", port)
		return
	}

	// Does not create connection if connection with this address an port
	// already exists
	if _, ok := rhost.teo.arp.find(addr, int(port), 0); ok {
		teolog.DebugVv(MODULE, "connection", addr, int(port), 0, "already exsists")
		return
	}

	go func() {
		rhost.teo.wg.Add(1)
		defer rhost.teo.wg.Done()
		// Create new connection
		tcd := rhost.teo.td.ConnectChannel(addr, int(port), 0)

		// Replay to address received in command data
		rhost.teo.sendToTcd(tcd, CmdNone, []byte{0})

		// Disconnect this connection if it does not added to peers arp table during timeout
		//go func(tcd *trudp.ChannelData) {
		time.Sleep(1500 * time.Millisecond)
		if !rhost.running {
			teolog.DebugVv(MODULE, "channel discovery task finished...")
			return
		}
		if _, ok := rhost.teo.arp.find(tcd); !ok {
			teolog.DebugVv(MODULE, "connection", addr, int(port), 0,
				"with peer does not established during timeout")
			tcd.CloseChannel()
			return
		}
	}()
}

// cmdConnectR process command CMD_CONNECT_R - a peer want connect to r-host
// command data structure: <n byte> <addr *C.char> ... <port uint32>
//   n - number of IPs
//   addr - IP address 0
//   ... - next IP address 1..n-1
//   port - port number
func (rhost *rhostData) cmdConnectR(rec *receiveData) {

	// Replay to address we got from peer
	rhost.teo.sendToTcd(rec.tcd, CmdNone, []byte{0})

	ptr := 1              // pointer to first IP
	from := rec.rd.From() // from
	data := rec.rd.Data() // received data
	numIP := data[0]      // number of received IPs
	port := int(C.getPort(unsafe.Pointer(&data[0]), C.size_t(len(data))))

	// Create data buffer to resend to peers
	// data structure: <from []byte> <0 byte> <addr []byte> <0 byte> <port uint32>
	makeData := func(from, addr string, port int) []byte {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, []byte(from))
		binary.Write(buf, binary.LittleEndian, byte(0))
		binary.Write(buf, binary.LittleEndian, []byte(addr))
		binary.Write(buf, binary.LittleEndian, byte(0))
		binary.Write(buf, binary.LittleEndian, uint32(port))
		return buf.Bytes()
	}

	// Send received IPs to this peer child(connected peers)
	for i := 0; i <= int(numIP); i++ {
		var caddr *C.char
		if i == 0 {
			clocalhost := append([]byte(localhostIP), 0)
			caddr = (*C.char)(unsafe.Pointer(&clocalhost[0]))
		} else {
			caddr = (*C.char)(unsafe.Pointer(&data[ptr]))
			ptr += int(C.strlen(caddr)) + 1
		}
		addr := C.GoString(caddr)

		// Send connected(who send this command) peer local IP address and port to
		// all this host child
		for peer, arp := range rhost.teo.arp.m {
			if arp.mode != -1 && peer != from {
				rhost.teo.SendTo(peer, CmdConnect, makeData(from, addr, port))
			}
		}
	}

	// Send connected(who send this command) peer IP address and port(defined by
	// this host) to all this host child
	for peer, arp := range rhost.teo.arp.m {
		if arp.mode != -1 && peer != from {
			rhost.teo.SendTo(peer, CmdConnect,
				makeData(from, rec.tcd.GetAddr().IP.String(), rec.tcd.GetAddr().Port))
			// \TODO: the discovery channel created here (issue #15)
		}
	}

	// Send all child IP address and port to connected(who send this command) peer
	for peer, arp := range rhost.teo.arp.m {
		if arp.mode != -1 && peer != from {
			rhost.teo.sendToTcd(rec.tcd, CmdConnect,
				makeData(peer, arp.tcd.GetAddr().IP.String(), arp.tcd.GetAddr().Port))
		}
	}
	//teolog.Debug(MODULE, "CMD_CONNECT_R command processed, from:", rec.rd.From())
	rhost.teo.com.log(rec.rd, "CMD_CONNECT_R command processed")
}

// connect send CMD_CONNECT_R command to r-host (connect to remote host)
// see command data format in 'connect' function description
func (rhost *rhostData) connect() {

	// Get local IP list
	ips, _ := rhost.getIPs()

	// Create command buffer
	buf := new(bytes.Buffer)
	_, port := rhost.teo.td.GetAddr()
	binary.Write(buf, binary.LittleEndian, byte(len(ips)))
	for _, addr := range ips {
		binary.Write(buf, binary.LittleEndian, []byte(addr))
		binary.Write(buf, binary.LittleEndian, byte(0))
	}
	binary.Write(buf, binary.LittleEndian, uint32(port))
	data := buf.Bytes()
	fmt.Printf("Connect to r-host, send local IPs\nip: %v\nport: %d\n", ips, port)

	// Send command to r-host
	rhost.teo.sendToTcd(rhost.tcd, CmdConnectR, data)
}

// reconnect reconnect to r-host if selected in function parameters channel is
// r-host trudp channel
func (rhost *rhostData) reconnect(tcd *trudp.ChannelData) {
	if rhost.isrhost(tcd) && rhost.running {
		rhost.wg.Done()
	}
}

// stop stops r-host reconnection
func (rhost *rhostData) stop(tcd *trudp.ChannelData) {
	rhost.reconnect(tcd)
}

// isrhost check if selected trudp channel is channel of r-host
func (rhost *rhostData) isrhost(tcd *trudp.ChannelData) (isRhost bool) {
	if rhost.tcd == tcd {
		isRhost = true
	}
	return
}

// getIPs return string slice with local IP address of this host
func (rhost *rhostData) getIPs() (ips []string, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			a := ip.String()
			// Check ipv6 address add [] if ipv6 allowed and
			// skip this address if ipv6 not allowed
			if strings.IndexByte(a, ':') >= 0 {
				if !rhost.teo.param.IPv6Allow {
					continue
				}
				a = "[" + a + "]"
			}
			ips = append(ips, a)
		}
	}
	return
}

// run starts connection and reconnection to r-host
func (rhost *rhostData) run() {
	if rhost.teo.param.RPort > 0 {
		rhost.running = true
		go func() {
			reconnect := 0
			rhost.teo.wg.Add(1)
			for rhost.teo.running {
				if reconnect > 0 {
					time.Sleep(2 * time.Second)
				}
				addr := rhost.teo.param.RAddr
				port := rhost.teo.param.RPort
				teolog.Connectf(MODULE, "connecting to r-host %s:%d:%d\n", addr, port, 0)
				rhost.tcd = rhost.teo.td.ConnectChannel(addr, port, 0)
				rhost.connect()
				rhost.wg.Add(1)
				rhost.wg.Wait()
				reconnect++
			}
			rhost.teo.wg.Done()
			rhost.running = false
		}()
	}
}

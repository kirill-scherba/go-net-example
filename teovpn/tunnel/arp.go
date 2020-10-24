// Tunnels arp table module

package tunnel

import (
	"fmt"
	"net"
	"sync"
)

// Arp is tunnel arp table receiver which hold tunnels arp map
type Arp struct {
	m   map[string]*net.UDPAddr
	mux sync.RWMutex
}

// newArp creates new arp table
func (t *Tunnel) newArp() {
	m := make(map[string]*net.UDPAddr)
	t.arp = &Arp{m: m}
}

// set add or update new UDP address
func (a *Arp) set(haddr net.HardwareAddr, raddr *net.UDPAddr) {
	a.mux.Lock()
	a.m[haddr.String()] = raddr
	a.mux.Unlock()
}

// get read UDP address by hardware address
func (a *Arp) get(haddr net.HardwareAddr) (raddr *net.UDPAddr, ok bool) {
	a.mux.RLock()
	raddr, ok = a.m[haddr.String()]
	a.mux.RUnlock()
	return
}

// String return harbare and udp address table in string
func (a *Arp) String() (s string) {
	a.mux.RLock()
	s += fmt.Sprintf("ARP Table:\n")
	s += fmt.Sprintf("----------\n")
	for i, v := range a.m {
		s += fmt.Sprintf("%s => %s\n", i, v)
	}
	s += fmt.Sprintf("----------\n")
	a.mux.RUnlock()
	return
}

// foreach calls callback function for each map element
func (a *Arp) foreach(f func(haddr string, raddr *net.UDPAddr)) {
	for i, v := range a.m {
		f(i, v)
	}
}

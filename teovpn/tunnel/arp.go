package tunnel

import (
	"net"
	"sync"
)

// Arp is tunnel arp table receiver which hold tunnels arp map
type Arp struct {
	m   map[string]*net.UDPAddr
	mux sync.RWMutex
}

func (t *Tunnel) newArp() {
	m := make(map[string]*net.UDPAddr)
	t.arp = &Arp{m: m}
}

func (a *Arp) set(haddr net.HardwareAddr, raddr *net.UDPAddr) {
	a.mux.Lock()
	a.m[haddr.String()] = raddr
	a.mux.Unlock()
}

func (a *Arp) get(haddr net.HardwareAddr) (raddr *net.UDPAddr, ok bool) {
	a.mux.RLock()
	raddr, ok = a.m[haddr.String()]
	a.mux.RUnlock()
	return
}

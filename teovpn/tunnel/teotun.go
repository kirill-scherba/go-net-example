package tunnel

import (
	"log"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/songgao/water"
)

// DefautPort is default port value for remote and local host
const (
	// DefaultInterface is default tap device interface name
	DefaultInterface = "tap0"

	// DefaultMtu is default interface device mtu
	DefaultMtu = 1500

	// DefautPort is default port value for remote and local host
	DefautPort = 9669

	// DefaultMask is default ip adress mask
	DefaultMask = 24
)

// SetDefault set sefault parameters
func (p *Params) SetDefault() *Params {
	p.Intrface = DefaultInterface
	p.Mtu = DefaultMtu
	p.Rport = DefautPort
	p.Lport = DefautPort
	p.Lmask = DefaultMask
	return p
}

// Params is application parameters
type Params struct {
	Intrface  string // interface name
	Mtu       int    // interface mtu
	Laddr     string // interface local address
	Lmask     int    // interface local address mask
	Raddr     string // remote host address
	Rport     int    // remote host port
	Lport     int    // this host port
	ShowHelpF bool   // show help flag
}

// Tunnel define teonet tunnel data structure
type Tunnel struct {
	p     *Params          // tunnel parameters
	sock  *net.UDPConn     // udp socket
	iface *water.Interface // tap interface
}

// New create new tunnel
func New(p *Params) (t *Tunnel) {
	t = &Tunnel{p: p}
	return t
}

// Run start and process teonet tunnel
func (t *Tunnel) Run() {
	t.newInterface()
	t.listner()
	// for {
	// 	time.Sleep(1 * time.Second)
	// }

}

// newInterface create new tap interface, set ip address and up it
func (t *Tunnel) newInterface() {
	// Create a new tunnel device (requires root privileges).
	conf := water.Config{DeviceType: water.TAP}
	var err error
	t.iface, err = water.New(conf)
	if err != nil {
		log.Fatalf("error creating tap device: %v", err)
	}

	// Setup IP properties.
	switch runtime.GOOS {
	case "linux":
		if err := exec.Command("/sbin/ip", "link", "set", "dev", t.iface.Name(),
			"mtu", strconv.Itoa(t.p.Mtu)).Run(); err != nil {
			log.Fatalf("ip link error: %v", err)
		}
		if err := exec.Command("/sbin/ip", "addr", "add", t.p.Laddr+"/"+strconv.Itoa(t.p.Lmask),
			"dev", t.iface.Name()).Run(); err != nil {
			log.Fatalf("ip addr error: %v", err)
		}
		if err := exec.Command("/sbin/ip", "link", "set", "dev", t.iface.Name(),
			"up").Run(); err != nil {
			log.Fatalf("ip link error: %v", err)
		}
	case "darwin":
		if err := exec.Command("/sbin/ifconfig", t.iface.Name(),
			strconv.Itoa(t.p.Mtu), "1300", t.p.Laddr, t.p.Raddr, "up").Run(); err != nil {
			log.Fatalf("ifconfig error: %v", err)
		}
	default:
		log.Fatalf("no tap support for: %v", runtime.GOOS)
	}
	log.Printf("interface %s sucessfully created\n", t.iface.Name())
}

// listner create new UDP socket and start udp listner
func (t *Tunnel) listner() {

	// Create a new UDP socket.
	//_, port, _ := net.SplitHostPort(t.netAddr)
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("", strconv.Itoa(t.p.Lport)))
	if err != nil {
		log.Fatalf("error resolving address: %v", err)
	}
	t.sock, err = net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalf("error listening on socket: %v", err)
	}
	log.Printf("UDP listner started at port %d\n", t.p.Lport)
	b := make([]byte, 1<<16)
	for {
		n, raddr, err := t.sock.ReadFromUDP(b)
		if err != nil {
			// if isDone(ctx) {
			// 	return
			// }
			log.Printf("net read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
	}
	defer t.sock.Close()
}

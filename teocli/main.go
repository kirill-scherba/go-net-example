package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/kirill-scherba/net-example-go/teocli/teocli"
)

func main() {
	fmt.Println("Teocli test application ver 1.0.0")

	// Flags variables
	var name string      // this client name
	var peer string      // remote server name (to send commands to)
	var raddr string     // remote host address
	var rport, rchan int // remote host port and channel (for TRUDP)
	var timeout int      // send echo timeout (in microsecond)

	// Flags
	flag.StringVar(&name, "n", "teocli-main-test-01", "this application name")
	flag.StringVar(&peer, "peer", "ps-server", "remote server name (to send commands to)")
	flag.StringVar(&raddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&rchan, "c", 0, "remote host channel (to connect to remote host TRUDP channel)")
	flag.IntVar(&rport, "r", 9010, "remote host port (to connect to remote host)")
	flag.IntVar(&timeout, "t", 1000000, "send echo timeout (in microsecond)")
	flag.Parse()

	for {
		running := true
		// Connect to L0 server
		fmt.Printf("try connecting to %s:%d ...\n", raddr, rport)
		teo, err := teocli.Connect(raddr, rport, false)
		if err != nil {
			panic(err)
		}
		// Send L0 login (requered after connect)
		fmt.Printf("send login\n")
		if _, err := teo.SendLogin(name); err != nil {
			panic(err)
		}
		// Send peers command
		fmt.Printf("send peers request\n")
		teo.Send(teocli.CmdLPeers, peer, nil)
		// Sender (send echo in loop)
		go func() {
			for i := 0; running; i++ {
				switch {

				// Send peers command
				case i%9 == 1:
					fmt.Printf("send peers request (%d,%d)\n", i, i%9)
					teo.Send(teocli.CmdLPeers, peer, nil)

				// Send large data packet with cmd 129
				case i%19 == 1:
					fmt.Printf("send large data packet with cmd 129 (%d,%d)\n", i, i%19)
					data := append([]byte(strings.Repeat("Q", 2000)), 0)
					teo.Send(129, peer, data)

				// Send echo
				default:
					fmt.Printf("send echo %d\n", i)
					teo.SendEcho(peer, fmt.Sprintf("Hello from Go(No %d)!", i))

				}
				time.Sleep(time.Duration(timeout) * time.Microsecond)
			}
		}()
		// Reader (read data and display it)
		for {
			packet, err := teo.Read()
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Printf("got cmd %d from %s, data len: %d, data: %v\n",
				packet.Command(), packet.From(), len(packet.Data()), packet.Data())
			switch packet.Command() {
			// Echo answer
			case teocli.CmdLEchoAnswer:
				if t, err := packet.TripTime(); err != nil {
					fmt.Println("trip time error:", err)
				} else {
					fmt.Println("trip time (ms):", t)
				}
			// Peers answer
			case teocli.CmdLPeersAnswer:
				ln := strings.Repeat("-", 59)
				fmt.Println("PeerAnswer received\n"+ln, "\n"+packet.Peers()+ln)
			}
		}
		teo.Disconnect()
		running = false
		time.Sleep(5 * time.Second)
	}
}

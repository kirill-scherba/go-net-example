package main

import (
	"flag"
	"fmt"
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
		// Send peers command (for this test)
		fmt.Printf("send peers request\n")
		teo.Send(72, peer, nil)
		// Sender (send echo in loop)
		go func() {
			for {
				fmt.Printf("send echo\n")
				teo.SendEcho(peer, "Hello from go!")
				time.Sleep(time.Duration(timeout) * time.Microsecond)
			}
		}()
		// Reader (read data and display it)
		for {
			packet, _ := teo.Read()
			if len(packet) == 0 {
				fmt.Println("disconnected...")
				break
			}
			fmt.Println("got packet", len(packet), packet)
			if packet[0] == 66 {
				fmt.Println("trip time (ms):", teo.ProccessEchoAnswer(packet))
			}
		}

		break
	}
}

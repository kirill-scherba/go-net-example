package teocli

import (
	"fmt"
	"time"
)

// StartCommand teocli first command interface
type StartCommand interface {
	Command(teo *TeoLNull)
}

// Command teocli input command interface
type Command interface {
	Cmd() byte
	Command(pac *Packet) bool
}

// Run cnnect and run
func Run(peer, name, raddr string, rport int, tcp bool, reconnectAfter time.Duration,
	startCommand StartCommand, commands ...Command) {

	var err error
	var teo *TeoLNull
	//var connected bool

	network := func(tcp bool) string {
		if tcp {
			return "TCP"
		}
		return "TRUDP"
	}

	// Reconnect loop, reconnect if disconnected afer reconnectAfter time (in sec)
	for {
		// Connect to L0 server
		//connected = false
		fmt.Printf("try %s connecting to %s:%d ...\n", network(tcp), raddr, rport)
		teo, err = Connect(raddr, rport, tcp)
		if err != nil {
			fmt.Println(err)
			time.Sleep(reconnectAfter)
			continue
		}
		//connected = true

		// Send Teonet L0 login (requered after connect)
		fmt.Printf("send login\n")
		if _, err := teo.SendLogin(name); err != nil {
			panic(err)
		}

		// Send peers command (just for test, it may be removed)
		fmt.Printf("send peers request\n")
		teo.SendTo(peer, CmdLPeers, nil)

		// Send Start game request to the teo-room
		startCommand.Command(teo)

		// Reader (receive data and process it)
		for {
			packet, err := teo.Read()
			if err != nil {
				fmt.Println(err)
				break
			}
			// Process commands
			for _, com := range commands {
				if cmd := com.Cmd(); cmd == packet.Command() {
					if com.Command(packet) {
						break
					}
				}
			}
		}
		// Disconnect
		teo.Disconnect()
		//connected = false
		time.Sleep(reconnectAfter)
	}
}

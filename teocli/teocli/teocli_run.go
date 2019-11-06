package teocli

import (
	"fmt"
	"time"
)

// StartCommand teocli first command interface
type StartCommand interface {
	Command(teo *TeoLNull, pac *Packet)
	Disconnected()
	Running() bool
	Stop()
}

// Command teocli input command interface
type Command interface {
	Cmd() byte
	Command(pac *Packet) bool
}

// Run conect and run
func Run(name, raddr string, rport int, tcp bool, timeout time.Duration,
	startCommand StartCommand, commands ...Command) {

	var err error
	var teo *TeoLNull

	network := func(tcp bool) string {
		if tcp {
			return "tcp"
		}
		return "trudp"
	}

	// Reconnect loop, reconnect if disconnected afer timeout time (in sec)
	for {
		// Connect to L0 server
		fmt.Printf("Try %s connecting to %s:%d ...\n", network(tcp), raddr, rport)
		teo, err = Connect(raddr, rport, tcp)
		if err != nil {
			fmt.Println(err)
			time.Sleep(timeout)
			continue
		}

		// Send Teonet L0 login (requered after connect)
		fmt.Printf("send login: '%s'\n", name)
		if _, err := teo.SendLogin(name); err != nil {
			fmt.Println(err)
			time.Sleep(timeout)
			continue
		}

		// Execute start command
		startCommand.Command(teo, nil)

		// Reader (receive data and process it)
		for {
			packet, err := teo.Read()
			if err != nil {
				fmt.Println(err)
				startCommand.Disconnected()
				break
			}
			// Process loadded commands
			for _, com := range commands {
				if cmd := com.Cmd(); cmd == packet.Command() {
					if com.Command(packet) {
						break
					}
				}
			}
		}

		// Stop running if ganning flag set to false
		if !startCommand.Running() {
			break
		}

		// Disconnect
		teo.Disconnect()
		time.Sleep(timeout)
	}
}

func readСookie() {
}

func writeСookie() {
}

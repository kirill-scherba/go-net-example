package main

import (
	"flag"
	"fmt"

	"github.com/kirill-scherba/go-net-example/udp/trudp"
)

func main() {
	fmt.Println("UDP teset application ver 1.0.0")

	var (
		rhost string
		rport int
		port  int
	)

	flag.IntVar(&rport, "r", 9010, "remote host port (to connect to remote host)")
	flag.StringVar(&rhost, "a", "", "remove host address (to connect to remote host)")
	flag.IntVar(&port, "p", 9000, "this host port (to remote hosts connect to this host)")
	flag.Parse()

	conn := trudp.Init(port)
	if rhost != "" && rport != 0 {
		conn.Connect(rhost, rport)
	}
	conn.Run()
}

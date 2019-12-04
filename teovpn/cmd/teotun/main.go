// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teotun is a simple udp tunnel which up tap interface beatven two node
//
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kirill-scherba/teonet-go/teovpn/tunnel"
)

// Version this teonet application version
const Version = "0.0.1"

// main parse aplication parameters and connect to Teonet. When teonet connected
// the game started
func main() {
	fmt.Println("Teotun ver " + Version)

	// Flags variables
	p := new(tunnel.Params).SetDefault()
	flag.StringVar(&p.Intrface, "i", p.Intrface, "interface name")
	flag.IntVar(&p.Mtu, "mtu", p.Mtu, "interface mtu")
	flag.StringVar(&p.Laddr, "l", p.Laddr, "local host ip address to set to interface")
	flag.IntVar(&p.Lmask, "mask", p.Lmask, "local host ip address mask")
	flag.StringVar(&p.Raddr, "a", p.Raddr, "remote host address to connect to remote host")
	flag.IntVar(&p.Rport, "r", p.Rport, "remote host port (to connect to remote host)")

	flag.BoolVar(&p.ShowHelpF, "h", false, "show this help message")
	flag.Parse()

	// Check flags and show usage
	if p.ShowHelpF {
		flag.Usage()
		os.Exit(0)
	}
	if p.Laddr == "" {
		fmt.Printf("the 'local host ip address' not defined, use -l to define it\n")
		flag.Usage()
		os.Exit(0)
	}
	// if p.Raddr == "" {
	// 	fmt.Printf("the 'remote host address' not defined, use -a to define it\n")
	// 	flag.Usage()
	// 	os.Exit(0)
	// }
	// if p.Rport == 0 {
	// 	fmt.Printf("wrong 'remote host port' value, use -r to set it\n")
	// 	flag.Usage()
	// 	os.Exit(0)
	// }

	t := tunnel.New(p)
	t.Run()
}

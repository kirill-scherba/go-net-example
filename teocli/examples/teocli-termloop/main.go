// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet client using linux terminal game engine
//
// This is simple terminal game with teonet client connected to teonet l0
// server and teonet room controller. This application connect to l0 server
// first and than request room in room controller. When room controller answer
// with room request answer this game application can send its hero position and
// show position of other players entered to the same room.
//
// Install client and server:
//
//  go get github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go get github.com/kirill-scherba/teonet-go/teonet
//
// Run server applications:
//
//  # run teonet l0 server
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teonet
//  go run . -p 7050 -l0-allow teo-l0
//
//  # run teonet room controller
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teonet/app/teoroom
//  go run . -r 7050 teo-room
//
// Run this game client application:
//
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go run . -r 7050 -peer teo-room -n game-01
//
//  cd $GOPATH/src/github.com/kirill-scherba/teonet-go/teocli/examples/teocli-termloop
//  go run . -r 7050 -peer teo-room -n game-02
//
// To exit from this game type Ctrl+C twice. When you start play next time
// you'll be connected to another room.
//
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
)

// Version this teonet application version
const Version = "0.0.1"

// main parse aplication parameters and connect to Teonet. When teonet connected
// the game started
func main() {
	fmt.Println("Teocli termloop application ver " + Version)

	// Flags variables
	var name string  // this client name
	var peer string  // room controller peer name
	var raddr string // l0 server address
	var rport int    // l0 server port
	var timeout int  // reconnect timeout (in seconds)
	var tcp bool     // connect by TCP flag

	// Flags
	flag.StringVar(&name, "n", "teocli-go-main-test-01", "this application name")
	flag.StringVar(&peer, "peer", "teo-room", "teo-room peer name (to send commands to)")
	flag.StringVar(&raddr, "a", "localhost", "remote host address (to connect to remote host)")
	flag.IntVar(&rport, "r", 9010, "l0 server port (to connect to l0 server)")
	flag.BoolVar(&tcp, "tcp", false, "connect by TCP")
	flag.IntVar(&timeout, "t", 5, "reconnect after timeout (in second)")
	flag.Parse()

	// Run teonet game (connect to teonet, start game and process received commands)
	run(name, peer, raddr, rport, tcp, time.Duration(timeout)*time.Second)
}

// Run connect to teonet, start game and process received commands
func run(name, peer, raddr string, rport int, tcp bool, timeout time.Duration) (tg *Teogame) {
	tg = &Teogame{peer: peer, player: make(map[byte]*Player)}
	teocli.Run(name, raddr, rport, tcp, timeout, startCommand(tg), inputCommands(tg)...)
	return
}

// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server statistic module.
//
// Hold statictic data, printf statistic table to terminal, return statictics
// string, process teonet l0 client statistic commands (for remote consumers).

package teonet

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
)

// stat teonet l0 server staistic
type l0Stat struct {
	l0        *l0Conn
	isUpdated bool
}

// statNew sreates new statistic data struct and method receiver
func (l0 *l0Conn) statNew() (stat *l0Stat) {
	stat = &l0Stat{l0: l0, isUpdated: true}
	if l0.teo.param.ShowClientsStatF {
		stat.process()
	}
	return
}

// update set update l0Stat value to true
func (stat *l0Stat) updated() {
	stat.isUpdated = true
}

// process print statistic continuously
func (stat *l0Stat) process() {
	go func() {
		var str string
		stat.updated()
		stat.l0.teo.wg.Add(1)
		for stat.l0.teo.running && stat.l0.teo.param.ShowClientsStatF {
			if stat.isUpdated {
				str = stat.sprint()
			}
			fmt.Print(str)
			time.Sleep(250 * time.Millisecond)
		}
		stat.l0.teo.wg.Done()
	}()
}

// print print teonet arp table
func (stat *l0Stat) print() {
	if stat.l0.allow && stat.l0.teo.param.ShowClientsStatF {
		fmt.Print(stat.sprint())
	}
}

// sprint return teonet l0 server clients table string
func (stat *l0Stat) sprint() (str string) {
	if !(stat.l0.allow && stat.l0.teo.param.ShowClientsStatF) {
		return
	}

	var line = "\033[2K" + strings.Repeat("-", 77) + "\n"
	var length, lenadd = 0, 7
	stat.isUpdated = false

	// Sort clients table:
	// read clients map keys to slice and sort it
	keys := make([]string, len(stat.l0.mn))
	for name := range stat.l0.mn {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	// Header
	str = fmt.Sprintf("" +
		"\0337" + // Save cursor
		"\033[0;0H" + // Set cursor to the top
		"\033[?7l" + // Does not wrap
		"\033[2K" + // Clear line
		"L0 clients statistic:\n" +
		line +
		"\033[2K" + "  # Name                    Net    Address                    Send     Recv\n" +
		line,
	)

	// Body
	for _, name := range keys {
		cli, ok := stat.l0.mn[name]
		if !ok {
			continue
		}
		length++
		str += fmt.Sprintf("\033[2K"+ // Clear line
			"%3d %s%-22.*s%s  %-6s %s%-22.*s%s %8d %8d\n",
			length,               // Number of line
			teokeys.ANSIGreen,    //
			22,                   // Length of client name
			name,                 // Client name
			teokeys.ANSINone,     //
			stat.l0.network(cli), // Network type: tcp, trudp, etc.
			teokeys.ANSIYellow,   //
			22,                   // Length of client address
			cli.addr,             // Client address
			teokeys.ANSINone,     //
			cli.stat.send,        // Number of send packets
			cli.stat.receive,     // Number of receive packets
		)
	}

	// Footer
	str += fmt.Sprintf(line+
		"\033[2K\n"+ // Clear line
		"\033[%d;r"+ // Set scroll mode
		"\0338", // Restore cursor
		length+lenadd,
	)
	return
}

// send set send operation in statistic
func (stat *l0Stat) send(client *client, packet []byte) {
	client.stat.send++
	// client.stat.sendRT.Calculate(len(packet))
	stat.updated()
}

// receive set receive operation in statistic
func (stat *l0Stat) receive(client *client, packet []byte) {
	client.stat.receive++
	// client.stat.receiveRT.Calculate(len(packet))
	stat.updated()
}

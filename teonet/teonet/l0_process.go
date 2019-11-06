// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server received data packet processing module

package teonet

import (
	"net"
	"strings"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

// check checks that received trudp packet is l0 packet and process it so.
// Return satatus 1 if not processed(if it is not teocli packet), or 0 if
// processed and send, or -1 if part of packet received and we wait next
// subpacket
func (l0 *l0Conn) check(tcd *trudp.ChannelData, packet []byte) (p []byte, status int) {
	if !l0.allow {
		return nil, 1
	}
	// Find this trudp key (connection) in cliens table and get cli, or create
	// new if not fount
	key := tcd.GetKey()
	var cli *teocli.TeoLNull
	client, ok := l0.findAddr(key)
	if !ok {
		cli, _ = teocli.Init(false) // create new client
	} else {
		cli = client.cli
	}
	return l0.packetCheck(cli, key, tcd, packet)
}

// packetCheck check that received tcp packet is l0 packet and process it so.
// Return satatus 1 if not processed(if it is not teocli packet), or 0 if
// processed and send, or -1 if part of packet received and we wait next subpacket
func (l0 *l0Conn) packetCheck(cli *teocli.TeoLNull, addr string, conn conn, data []byte) (p []byte, status int) {
check:
	p, status = cli.PacketCheck(data)
	switch status {
	case 0:
		l0.toprocess(p, cli, addr, conn)
		data = nil
		goto check // check next packet in read buffer
	case -1:
		if data != nil {
			teolog.DebugVv(MODULE, "packet not received yet (got part of packet)")
		}
	case 1:
		teolog.DebugVvf(MODULE, "wrong packet received (drop it): %d, data: %v\n", len(p), p)
	}
	return
}

// toprocess send packet to packet processing
func (l0 *l0Conn) toprocess(p []byte, cli *teocli.TeoLNull, addr string, conn conn) {
	l0.ch <- &packet{packet: p, client: &client{cli: cli, addr: addr, conn: conn}}
	return
}

// Process received packets
func (l0 *l0Conn) process() {
	teolog.DebugVvf(MODULE, "l0 packet process started\n")
	l0.ch = make(chan *packet)
	l0.teo.wg.Add(1)
	go func() {
	packetGet:
		for pac := range l0.ch {
			teolog.DebugVvf(MODULE,
				"valid packet received from client %s, length: %d\n",
				pac.client.addr, len(pac.packet),
			)
			p := pac.client.cli.NewPacket(pac.packet)

			// Find address in clients map and add if it absent, or send packet to
			// peer if client already exsists
			client, ok := l0.findAddr(pac.client.addr)
			if !ok {
				// First message from client should contain login command. Loging
				// command parameters: cmd = 0 and name = "" and data = 'client name'.
				// When we got valid data in l0 login we add clien to cliens map and all
				// next commands form this client resends to peers with this name (name
				// from login command data)
				if d := p.Data(); p.Command() == 0 && p.Name() == "" {
					pac.client.name = string(d[:len(d)-1])
					l0.stat.receive(pac.client, d)
					l0.add(pac.client)

					// Send to users registrar.
					// Check l0 config and find valid prefixes and if prefix
					// find than send login command to users registrar service.
					// Users registrar return answer [1] if login valid and user
					// my continue or data structure:
					// UserNew{user_id gocql.UUID,access_tocken gocql.UUID,prefix string}
					prefix := l0.param.Value().(*param).Prefix
					for _, p := range prefix {
						if strings.HasPrefix(pac.client.name, p) {
							go func() {
								_, err := l0.sendToRegistrar([]byte(pac.client.name))
								if err != nil {
									// TODO: disconnect this client?
									// send to registrar error: %s
									return
								}
								// TODO: send answer to client and set valid client id
								// to it name with `l0.rename`.
								// It alread set inside sendToRegistrar...
								// Move that code here?
							}()
							continue packetGet
						}
					}

					// Send login command to auth service
					l0.sendToAuth(d)
					continue
				}

				// Incorrect login packet received
				teolog.Errorf(MODULE,
					"incorrect login packet received from client %s, disconnect...\n",
					pac.client.addr)
				//fmt.Printf("cmd: %d, to: %s, data: %v\n", p.Command(), p.Name(), p.Data())
				// Send http answer to tcp request
				switch pac.client.conn.(type) {
				case net.Conn:
					pac.client.conn.Write([]byte("HTTP/1.1 200 OK\n" +
						"Content-Type: text/html\n\n" +
						"<html><body>Hello!</body></html>\n"))
				}
				pac.client.conn.Close()
				continue
			}

			// if client exists: send it command to Client connected to this server
			// or to Peer for exising client
			l0.stat.receive(client, p.Data())
			if _, ok := l0.findName(p.Name()); ok {
				l0.sendTo(client.name, p.Name(), p.Command(), p.Data())
			} else {
				l0.sendToPeer(p.Name(), client.name, p.Command(), p.Data())
			}
		}
		l0.closeAll()
		teolog.DebugVv(MODULE, "l0 packet process stopped")
		l0.teo.wg.Done()
	}()
}

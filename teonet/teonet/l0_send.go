// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server send commands module

package teonet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teousers"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// packetCreate creates packet data for sendToPeer and sendToL0
func (l0 *l0Conn) packetCreate(client string, cmd byte, data []byte) []byte {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, cmd)                       // Command
	binary.Write(buf, le, byte(len(client)+1))       // Client name length (include trailing zero)
	binary.Write(buf, le, uint16(len(data)))         // Packet data length
	binary.Write(buf, le, append([]byte(client), 0)) // Client name (include trailing zero)
	binary.Write(buf, le, []byte(data))              // Packet data
	return buf.Bytes()
}

// packetParse parse command data
func (l0 *l0Conn) packetParse(d []byte) (name string, cmd byte, data []byte) {
	buf := bytes.NewReader(d)
	le := binary.LittleEndian
	var fromLen byte
	var dataLen uint16
	binary.Read(buf, le, &cmd)     // Command
	binary.Read(buf, le, &fromLen) // From client name length (include trailing zero)
	binary.Read(buf, le, &dataLen) // Packet data length
	name = func() string {
		nameBuf := make([]byte, fromLen)
		binary.Read(buf, le, nameBuf)
		return string(nameBuf[:len(nameBuf)-1])
	}()
	data = func() []byte {
		d := make([]byte, dataLen)
		binary.Read(buf, le, d)
		return d
	}()
	return
}

// sendToPeer (send from L0 server to peer) send packet received from client to peer
func (l0 *l0Conn) sendToPeer(peer string, client string, cmd byte, data []byte) {
	teolog.DebugVf(MODULE,
		"send cmd: %d, %d bytes data packet to peer %s, from client: %s",
		cmd, len(data), peer, client,
	)
	l0.teo.SendTo(peer, CmdL0, l0.packetCreate(client, cmd, data)) // Send to peer
}

// sendToL0 (send from peer to L0 server) send packet from peer to client
func (l0 *l0Conn) sendToL0(peer string, client string, cmd byte, data []byte) (length int, err error) {
	teolog.DebugVf(MODULE,
		"send cmd: %d, %d bytes data packet to l0 %s, from client: %s",
		cmd, len(data), peer, client,
	)
	return l0.teo.SendTo(peer, CmdL0To, l0.packetCreate(client, cmd, data)) // Send to L0
}

// sendTo send command from peer or client to L0 client connected to this server
func (l0 *l0Conn) sendTo(from string, toClient string, cmd byte, data []byte) (length int, err error) {

	// Get client data from name map
	l0.mux.Lock()
	client, ok := l0.mn[toClient]
	l0.mux.Unlock()
	if !ok {
		err = fmt.Errorf("send to client: can't find client '%s' in clients map",
			toClient)
		teolog.Error(MODULE, err.Error())
		return
	}

	teolog.DebugVf(MODULE,
		"got cmd: %d, %d bytes data packet from peer %s, to client: %s\n",
		cmd, len(data), from, toClient,
	)

	packet, err := client.cli.PacketCreate(uint8(cmd), from, data)
	if err != nil {
		teolog.Error(MODULE, err.Error())
		return
	}

	teolog.DebugVf(MODULE,
		"send cmd: %d, %d bytes data packet, to %s l0 client: %s\n",
		cmd, len(data), l0.network(client), client.name)

	l0.stat.send(client, packet)
	return client.conn.Write(packet)
}

// sendToRegistrar sends login commands to users registrar got answer and send
// answer to client
func (l0 *l0Conn) sendToRegistrar(d []byte) (data []byte, err error) {
	teoCDB := "teo-cdb"
	CmdAuth := byte(133)
	teolog.Debugf(MODULE,
		"login command, send to users registrar (teo-cdb): %s, data: %v\n",
		teoCDB, d)
	l0.teo.SendTo(teoCDB, CmdAuth, d)
	r := <-l0.teo.WaitFrom(teoCDB, CmdAuth, 1*time.Second)
	if r.Err != nil {
		err = r.Err
		teolog.Errorf(MODULE,
			"does not receive answer from users registrar (teo-cdb): %s\n",
			teoCDB)
		return
	}
	data = r.Data
	teolog.Debugf(MODULE,
		"got answer from users registrar (teo-cdb): %s, %v\n",
		teoCDB, r.Data)

	// Check answer
	res := &teousers.UserResponce{}
	err = res.UnmarshalBinary(data)
	if err != nil {
		// can't create new user
		return
	}
	fmt.Printf("\nresult: %v\n", res)

	// Set client name in system
	client := res.Prefix + "-" + res.ID.String()
	cookies := res.Prefix + "-" + res.AccessToken.String()
	l0.rename(string(d), client)

	// Send command to client
	l0.sendTo("", client, 129, []byte(cookies))

	return
}

// sendToRegistrar sends login commands to users registrar
func (l0 *l0Conn) sendToAuth(d []byte) (length int, err error) {
	teoAuth := "teo-auth"
	teolog.Debugf(MODULE, "login command, send to auth: %s, data: %v\n", teoAuth, d)
	l0.teo.SendTo(teoAuth, CmdUser, d)
	return
}

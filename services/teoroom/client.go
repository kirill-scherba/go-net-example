// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teoroom

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Client data
type Client struct {
	tr                   *Teoroom // Pointer to Teoroom receiver
	name                 string   // Client name
	data                 [][]byte // Client data (which sends to new room clients)
	*teonet.L0PacketData          // Client L0 address
}

// newClient creates new Client (or add Client to Room controller).
func (tr *Teoroom) newClient(c *teonet.Packet) (cli *Client) {
	l0 := c.GetL0()
	client := c.From()
	cli = &Client{tr: tr, name: client, L0PacketData: l0}
	tr.mcli[client] = cli
	return
}

// roomRequest finds room for client or create new room and adds client to this
// room. It returns roomID and cliID or error if not found. The RoomID is an
// unical room number since this application started. The cliID is a client
// number (and position) in this room.
func (cli *Client) roomRequest() (roomID uint32, cliID int, err error) {
	for _, rid := range cli.tr.creating {
		if r, ok := cli.tr.mroom.find(rid); ok &&
			r.state != RoomClosed && r.state != RoomStopped &&
			func() bool { _, ok := r.cliwas[cli.name]; return !ok }() &&
			len(r.client) < r.gparam.MaxClientsInRoom {
			return r.id, r.addClient(cli), nil
		}
	}
	r := cli.tr.newRoom()
	// send disconnect to rooms clients if can't find clients during WaitForMinClients
	go func() {
		<-time.After(time.Duration(r.gparam.WaitForMinClients) * time.Millisecond)
		if !(r.state == RoomCreating && r.numClients() < r.gparam.MinClientsToStart) {
			return
		}
		fmt.Printf("Room id %d stopped. Can't find players during start "+
			"timeout.\n", r.id)
		r.setState(RoomStopped)
		r.funcToClients(func(l0 *teonet.L0PacketData, client string) {
			r.tr.teo.SendToClientAddr(l0, client, teoroomcli.ComDisconnect, nil)
			r.tr.Process.ComDisconnect(client)
		})
	}()
	return r.id, r.addClient(cli), nil
}

// roomClientID finds client in rooms and returns roomID and cliID or error if
// not found. The RoomID is an unical room number since this application
// started. The cliID is a client number (and position) in this room.
func (cli *Client) roomClientID() (roomID uint32, cliID int, err error) {
	var r *Room
	cli.tr.mroom.mx.RLock()
	defer cli.tr.mroom.mx.RUnlock()
	for roomID, r = range cli.tr.mroom.m {
		for id, c := range r.client {
			if c == cli {
				cliID = id
				return
			}
		}
	}
	err = fmt.Errorf("can't find client %s in room structure", cli.name)
	return
}

// sendState set client state and send it to cdb room statistic
func (cli *Client) sendState(state byte, roomID gocql.UUID) (err error) {

	cliID := strings.SplitN(cli.name, "-", 2)
	if len(cliID) != 2 {
		err = errors.New("wrong name")
		return
	}
	fmt.Printf("Send client state %d, id: %s, game: %s, roomID: %s\n",
		state, cliID[1], cliID[0], roomID.String())
	clientID, err := gocql.ParseUUID(cliID[1])
	if err != nil {
		return
	}
	stats.SendClientState(cli.tr.teo, state, roomID, clientID)
	return
}

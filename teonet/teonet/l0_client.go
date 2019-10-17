// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server client module

package teonet

import (
	"errors"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// client data structure
type client struct {
	name string           // name
	addr string           // address (ip:port:ch)
	conn conn             // Connection tcp (net.Conn), websocket or trudp (*trudp.ChannelData)
	stat clientStat       // Statistic
	cli  *teocli.TeoLNull // teocli connection to use readBuffer
}

// clientStat client statistic
type clientStat struct {
	send    int // send packes to client counter
	receive int // receive packes from client counter
	// sendRT    trudp.RealTimeSpeed // send packes to client real time counter
	// receiveRT trudp.RealTimeSpeed // receive packes to client real time counter
}

// add new client
func (l0 *l0Conn) add(client *client) {
	teolog.Connectf(MODULE, "client %s (%s) connected\n", client.name, client.addr)
	l0.closeName(client.name)
	l0.mux.Lock()
	l0.ma[client.addr] = client
	l0.mn[client.name] = client
	l0.mux.Unlock()
	l0.stat.updated()
}

// rename client
func (l0 *l0Conn) rename(name, newname string) {
	if name == newname {
		return
	}
	cli, ok := l0.findName(name)
	if !ok {
		return
	}
	teolog.Connectf(MODULE, "client %s renamed to %s\n", name, newname)
	l0.mux.Lock()
	//delete(l0.ma, cli.addr)
	delete(l0.mn, cli.name)
	cli.name = newname
	//l0.ma[cli.addr] = cli
	l0.mn[cli.name] = cli
	l0.mux.Unlock()
}

// close disconnect connected client
func (l0 *l0Conn) close(client *client) (err error) {
	if client == nil {
		err = errors.New("client is nil")
		teolog.Error(MODULE, err.Error())
		return
	}
	if client.conn != nil {
		teolog.Connectf(MODULE, "client %s (%s) disconnected\n", client.name, client.addr)
		client.conn.Close()
		client.conn = nil
	}
	l0.mux.Lock()
	delete(l0.ma, client.addr)
	delete(l0.mn, client.name)
	l0.mux.Unlock()
	l0.stat.updated()
	return
}

// closeAddr disconnect connected client by address
func (l0 *l0Conn) closeAddr(addr string) (done bool) {
	if client, ok := l0.findAddr(addr); ok {
		l0.close(client)
		done = true
	}
	return
}

// closeName disconnect connected client by name
func (l0 *l0Conn) closeName(name string) (done bool) {
	if client, ok := l0.findName(name); ok {
		l0.close(client)
		done = true
	}
	return
}

// closeAll disconnect all connected clients
func (l0 *l0Conn) closeAll() {
	for _, client := range l0.mn {
		l0.close(client)
	}
}

// findAddr finds client in clients map by address
func (l0 *l0Conn) findAddr(addr string) (client *client, ok bool) {
	l0.mux.Lock()
	client, ok = l0.ma[addr]
	l0.mux.Unlock()
	return
}

// findName finds client in clients map by name
func (l0 *l0Conn) findName(name string) (client *client, ok bool) {
	l0.mux.Lock()
	client, ok = l0.mn[name]
	l0.mux.Unlock()
	return
}

// exists return true if client exists in clients map
func (l0 *l0Conn) exists(client *client) (ok bool) {
	for _, cli := range l0.mn {
		if cli == client {
			ok = true
			return
		}
	}
	return
}

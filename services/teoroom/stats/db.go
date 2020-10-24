// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom stats database module.
//
// This service uses Scylla database and gocql and gocqlx packages to work with
// db. Usually teoroomcdb package used in Teonet teocdb service application to
// provide rooms database functions to other teonet network applications.

package stats

import (
	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teocdb"
)

// db data structure and methods receiver.
type db struct {
	*Rooms
	*rooms
	*clients
	session *gocql.Session
}

// newCdb creates new cdb structure.
func newDb(hosts ...string) (d *db, err error) {
	d = &db{}
	d.rooms = d.newRooms()
	d.clients = d.newClients()
	err = d.connect(hosts...)
	return
}

// connect to database.
func (d *db) connect(hosts ...string) (err error) {
	keyspace := "teoroom"
	cluster := gocql.NewCluster(func() (h []string) {
		if h = hosts; len(h) > 0 {
			keyspace = h[0]
			h = h[1:]
		}
		if len(h) == 0 {
			h = teocdb.HostsDefault
		}
		return
	}()...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	d.session, err = cluster.CreateSession()
	d.rooms.session = d.session
	d.clients.session = d.session
	return
}

// close database connection.
func (d *db) close() {
	d.session.Close()
}

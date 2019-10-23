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
			h = []string{"172.17.0.2", "172.17.0.3", "172.17.0.4"}
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

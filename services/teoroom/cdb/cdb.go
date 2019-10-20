// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cdb (teoroomcdb) is the Teonet room cdb functions and datbase themes.
//
// Teoroomcdb provide teroom database functions executed in cdb.
//
// Install this go package:
//
//   go get github.com/kirill-scherba/teonet-go/services/teoroom/cdb
//
// Data base organisation
//
// To store database we use ScyllaDB. Run Scylla in Docker:
//   https://www.scylladb.com/download/open-source/#docker
//
// Install database schemas. Before you execute application which used this
// service, launch `cqlsh`:
//
//   docker exec -it scylla cqlsh
//
// and execute content of cql/teoroom.cql file.
//
//
package cdb

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/table"
)

// Room data structure
type Room struct {
	ID      gocql.UUID // Room ID
	RoomNum int        // Room number
	Started time.Time  // Time when room start
	Stopped time.Time  // Time when room stop
	State   int        // Current rooms state
}

// cdb data structure and methods receiver.
type cdb struct {
	//*Rooms
	session       *gocql.Session
	roomsTable    *table.Table
	roomsMetadata table.Metadata
}

// newCdb creates new cdb structure.
func newDb(hosts ...string) (d *cdb, err error) {
	d = &cdb{
		roomsMetadata: table.Metadata{
			Name: "rooms",
			Columns: []string{
				"id",
				"room_num",
				"started",
				"stopped",
				"state",
			},
			PartKey: []string{"id"},
			SortKey: []string{}, //"room_num", "started", "stopped", "state"},
		},
	}
	// roomsTable allows for simple CRUD operations based on personMetadata.
	d.roomsTable = table.New(d.roomsMetadata)
	err = d.connect(hosts...)
	return
}

// connect to cdb.
func (d *cdb) connect(hosts ...string) (err error) {
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
	return
}

// close cdb
func (d *cdb) close() {
	d.session.Close()
}

// set add new room or update existing. First input parameter is structure with
// filled RoomID, and all other fields from Room structure needs to set
// (usaly it may be Room structure with all fields filled). Next parameters
// is column names which will be set to database, it may be ommited and than
// all columns sets.
func (d *cdb) set(r interface{}, columns ...string) (err error) {
	var stmt string
	var names []string
	if len(columns) == 0 {
		stmt, names = d.roomsTable.Insert()
	} else {
		stmt, names = d.roomsTable.Update(columns...)
	}
	q := gocqlx.Query(d.session.Query(stmt), names).BindStruct(r)
	fmt.Println(q.String())
	return q.ExecRelease()
}

func (d *cdb) setCreating(roomNum int) (roomID gocql.UUID, err error) {
	roomID = gocql.TimeUUID()
	room := &Room{
		ID:      roomID,
		RoomNum: roomNum,
		Started: time.Now(),
		State:   teoroom.RoomCreating,
	}
	return roomID, d.set(room)
}

func (d *cdb) setRunning(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:    roomID,
		State: teoroom.RoomRunning,
	}
	return d.set(room, d.roomsMetadata.Columns[4])
}

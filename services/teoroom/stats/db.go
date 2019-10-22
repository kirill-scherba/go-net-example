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
	RoomNum uint32     // Room number
	Created time.Time  // Time when room created
	Started time.Time  // Time when room started
	Closed  time.Time  // Time when room closed to add players
	Stopped time.Time  // Time when room stopped
	State   int        // Current rooms state
}

// db data structure and methods receiver.
type db struct {
	*Rooms
	session       *gocql.Session
	roomsTable    *table.Table
	roomsMetadata table.Metadata
}

// Column numbers
const (
	colID       = iota // 0
	colRoomNum         // 1
	colCreated         // 2
	colStartded        // 3
	colClosed          // 4
	colStopped         // 5
	colState           // 6
)

// newCdb creates new cdb structure.
func newDb(hosts ...string) (d *db, err error) {
	d = &db{
		roomsMetadata: table.Metadata{
			Name: "rooms",
			Columns: []string{
				"id",       // 0
				"room_num", // 1
				"created",  // 2
				"started",  // 3
				"closed",   // 4
				"stopped",  // 5
				"state",    // 6
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
	return
}

// close database connection.
func (d *db) close() {
	d.session.Close()
}

// set add new room or update existing. First input parameter is structure with
// filled RoomID, and all other fields from Room structure needs to set
// (usaly it may be Room structure with all fields filled). Next parameters
// is column names which will be set to database, it may be ommited and than
// all columns sets.
func (d *db) set(r interface{}, columns ...string) (err error) {
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

// set using column numbers from roomsMetadata.Columns structure
func (d *db) setByColumnsNumber(r interface{}, columnsNum ...int) (err error) {
	var columns []string
	for _, column := range columnsNum {
		columns = append(columns, d.roomsMetadata.Columns[column])
	}
	return d.set(r, columns...)
}

// set creating state (create new rooms record)
func (d *db) setCreating(roomID gocql.UUID, roomNum uint32) (err error) {
	room := &Room{
		ID:      roomID,
		RoomNum: roomNum,
		Created: time.Now(),
		State:   teoroom.RoomCreating,
	}
	return d.set(room)
}

// set running state (started)
func (d *db) setRunning(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:      roomID,
		Started: time.Now(),
		State:   teoroom.RoomRunning,
	}
	return d.setByColumnsNumber(room, colStartded, colState)
}

// set closed state
func (d *db) setClosed(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:     roomID,
		Closed: time.Now(),
		State:  teoroom.RoomClosed,
	}
	return d.setByColumnsNumber(room, colClosed, colState)
}

// set stopped state
func (d *db) setStopped(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:      roomID,
		Stopped: time.Now(),
		State:   teoroom.RoomStopped,
	}
	return d.setByColumnsNumber(room, colStopped, colState)
}

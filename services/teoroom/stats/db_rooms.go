// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom stats database table rooms module.
//
// This service uses Scylla database and gocql and gocqlx packages to work with
// database. This module contain table rooms definition and functions.

package stats

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoroom"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
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

// Column numbers
const (
	rColID       = iota // 0
	rColRoomNum         // 1
	rColCreated         // 2
	rColStartded        // 3
	rColClosed          // 4
	rColStopped         // 5
	rColState           // 6
)

type rooms struct {
	roomsTable    *table.Table
	roomsMetadata table.Metadata
	session       *gocql.Session
}

// newRooms creates rooms receiver
func (d *db) newRooms() *rooms {
	r := &rooms{
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
	r.roomsTable = table.New(r.roomsMetadata)
	return r
}

// set add new room or update existing. First input parameter is structure with
// filled RoomID, and all other fields from Room structure needs to set
// (usaly it may be Room structure with all fields filled). Next parameters
// is column names which will be set to database, it may be ommited and than
// all columns sets.
func (d *rooms) set(r interface{}, columns ...string) (err error) {
	var stmt string
	var names []string
	if len(columns) == 0 {
		stmt, names = d.roomsTable.Insert()
	} else {
		stmt, names = d.roomsTable.Update(columns...)
	}
	q := gocqlx.Query(d.session.Query(stmt), names).BindStruct(r)
	teolog.Debug(MODULE, q.String())
	return q.ExecRelease()
}

// setByColumnsNumber set using column numbers from roomsMetadata.Columns structure
func (d *rooms) setByColumnsNumber(r interface{}, columnsNum ...int) (err error) {
	var columns []string
	for _, column := range columnsNum {
		columns = append(columns, d.roomsMetadata.Columns[column])
	}
	return d.set(r, columns...)
}

// setCreating set creating state (create new rooms record)
func (d *rooms) setCreating(roomID gocql.UUID, roomNum uint32) (err error) {
	room := &Room{
		ID:      roomID,
		RoomNum: roomNum,
		Created: time.Now(),
		State:   teoroom.RoomCreating,
	}
	return d.set(room)
}

// setRunning set running state (started)
func (d *rooms) setRunning(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:      roomID,
		Started: time.Now(),
		State:   teoroom.RoomRunning,
	}
	return d.setByColumnsNumber(room, rColStartded, rColState)
}

// setClosed set closed state
func (d *rooms) setClosed(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:     roomID,
		Closed: time.Now(),
		State:  teoroom.RoomClosed,
	}
	return d.setByColumnsNumber(room, rColClosed, rColState)
}

// setStopped set stopped state
func (d *rooms) setStopped(roomID gocql.UUID) (err error) {
	room := &Room{
		ID:      roomID,
		Stopped: time.Now(),
		State:   teoroom.RoomStopped,
	}
	return d.setByColumnsNumber(room, rColStopped, rColState)
}

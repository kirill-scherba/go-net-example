// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom stats database table clients module.
//
// This service uses Scylla database and gocql and gocqlx packages to work with
// database. This module contain table clients definition and functions.

package stats

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/table"
)

// Client data structure
type Client struct {
	RoomID       gocql.UUID // Room ID
	ID           gocql.UUID // ID
	Added        time.Time  // Time added
	Loadded      time.Time  // Time loadded
	Started      time.Time  // Time started
	Leave        time.Time  // Time leave
	Disconnected time.Time  // Time disconnected
	GameStat     []byte     // Game statistic
}

// Column numbers
const (
	cColRoomID       = iota // 0
	cColID                  // 1
	cColAdded               // 2
	cColLoadded             // 3
	cColStarted             // 4
	cColLeave               // 5
	cColDisconnected        // 6
	cColGameStat            // 7
)

type clients struct {
	clientsTable    *table.Table
	clientsMetadata table.Metadata
	session         *gocql.Session
}

// newRooms creates rooms receiver
func (d *db) newClients() *clients {
	c := &clients{
		clientsMetadata: table.Metadata{
			Name: "clients",
			Columns: []string{
				"room_id",      // 0
				"id",           // 1
				"added",        // 2
				"loadded",      // 3
				"started",      // 4
				"leave",        // 5
				"disconnected", // 6
				"game_stat",    // 7
			},
			PartKey: []string{"room_id", "id"},
			SortKey: []string{},
		},
	}
	// clientsTable allows for simple CRUD operations based on personMetadata.
	c.clientsTable = table.New(c.clientsMetadata)
	return c
}

// set add new client or update existing. First input parameter is structure with
// filled RoomID, ID, and all other fields from Client structure needs to set
// (usaly it may be Client structure with all fields filled). Next parameters
// is column names which will be set to database, it may be ommited and than
// all columns sets.
func (d *clients) set(r interface{}, columns ...string) (err error) {
	var stmt string
	var names []string
	if len(columns) == 0 {
		stmt, names = d.clientsTable.Insert()
	} else {
		stmt, names = d.clientsTable.Update(columns...)
	}
	q := gocqlx.Query(d.session.Query(stmt), names).BindStruct(r)
	teolog.DebugV(MODULE, q.String())
	return q.ExecRelease()
}

// setByColumnsNumber set using column numbers from clientsMetadata.Columns structure
func (d *clients) setByColumnsNumber(r interface{}, columnsNum ...int) (err error) {
	var columns []string
	for _, column := range columnsNum {
		columns = append(columns, d.clientsMetadata.Columns[column])
	}
	return d.set(r, columns...)
}

// setAdded client added to room (create new client and set time)
func (d *clients) setAdded(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID: roomID,
		ID:     id,
		Added:  time.Now(),
	}
	return d.set(client)
}

// setLoadded client loaded to room
func (d *clients) setLoadded(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID:  roomID,
		ID:      id,
		Loadded: time.Now(),
	}
	return d.setByColumnsNumber(client, cColLoadded)
}

// setStarted client started play in room
func (d *clients) setStarted(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID:  roomID,
		ID:      id,
		Started: time.Now(),
	}
	return d.setByColumnsNumber(client, cColStarted)
}

// setLeave set client leave from rom
func (d *clients) setLeave(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID: roomID,
		ID:     id,
		Leave:  time.Now(),
	}
	return d.setByColumnsNumber(client, cColLeave)
}

// setDisconnected client loaded to room
func (d *clients) setDisconnected(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID:       roomID,
		ID:           id,
		Disconnected: time.Now(),
	}
	return d.setByColumnsNumber(client, cColDisconnected)
}

// setGameStat set clients game statistic
func (d *clients) setGameStat(roomID gocql.UUID, id gocql.UUID, gameStat []byte) (err error) {
	client := &Client{
		RoomID:   roomID,
		ID:       id,
		GameStat: gameStat,
	}
	return d.setByColumnsNumber(client, cColGameStat)
}

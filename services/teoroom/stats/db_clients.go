// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom stats database table clients module.
//
// This service uses Scylla database and gocql and gocqlx packages to work with
// database. This module contain table clients definition and functions.

package stats

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/table"
)

// Client data structure
type Client struct {
	RoomID   gocql.UUID // Room ID
	ID       gocql.UUID // ID
	Added    time.Time  // Time added
	Leave    time.Time  // Time leave
	GameStat []byte     // Game statistic
}

// Column numbers
const (
	cColRoomID   = iota // 0
	cColID              // 1
	cColAdded           // 2
	cColLeave           // 3
	cColGameStat        // 4
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
				"room_id",   // 0
				"id",        // 1
				"added",     // 2
				"leave",     // 3
				"game_stat", // 4
			},
			PartKey: []string{"room_id", "id"},
			SortKey: []string{}, //"added", "leave"
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
	fmt.Println(q.String())
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

// setLeave set client leave time
func (d *clients) setLeave(roomID gocql.UUID, id gocql.UUID) (err error) {
	client := &Client{
		RoomID: roomID,
		ID:     id,
		Leave:  time.Now(),
	}
	return d.setByColumnsNumber(client, cColLeave)
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

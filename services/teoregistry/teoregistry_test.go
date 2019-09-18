// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teoregistry test module

// Data base organisation (we use ScyllaDB):
//
// Run Scylla in Docker: https://www.scylladb.com/download/open-source/#docker
/* Before you execute this tests Launch `cqlsh` and execute:
//
// Keyspace 'teoregistry'
// Atention: this test use another default keyspace than main package code. So
// you need set this keyspace and tables again.
//
CREATE KEYSPACE teoregistry_test WITH replication = { 'class': 'SimpleStrategy', 'replication_factor' : 3 };
USE teoregistry_test;
//
// And exequte 'Tables' from teoregistry.go (see line 25)
//
*/

package teoregistry

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
)

func TestTeoregistry(t *testing.T) {

	const AppName = "teotest-7755"
	var tre *Teoregistry
	var uuid gocql.UUID
	var app Application
	var com Command
	var err error

	t.Run("Connect", func(t *testing.T) {
		tre, err = Connect("teoregistry_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer tre.Close()

	app = Application{
		Uuid:   gocql.TimeUUID(),
		Name:   AppName,
		Descr:  "Teonet test application",
		Author: "Kirill Scherba <kirill@scherba.ru>",
		Git:    "https://github.com/kirill-scherba/teonet-go/",
		Goget:  "github.com/kirill-scherba/teonet-go/services/teoregistry",
	}

	t.Run("Application Set(insert new application)", func(t *testing.T) {
		uuid, err = tre.app.Set(&app)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was set to database\n", uuid)
	})

	t.Run("Application Set(update application)", func(t *testing.T) {
		app.License = "MIT"
		uuid, err = tre.app.Set(&app)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was updated in database\n", uuid)
	})

	t.Run("Application Get", func(t *testing.T) {
		a, err := tre.app.Get(uuid)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was get from database:\n%v\n", uuid, a)
	})

	t.Run("Application Set(insert new application)", func(t *testing.T) {
		app.Uuid = gocql.TimeUUID()
		app.Descr = "Teonet supper test application"
		uuid, err := tre.app.Set(&app)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was set to database\n", uuid)
	})

	t.Run("Application Num", func(t *testing.T) {
		numApp, err := tre.app.Num()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Database contain %d applications\n", numApp)
	})

	t.Run("Application List", func(t *testing.T) {
		listApp, err := tre.app.List()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application list was get from database: %v\n", listApp)
		for _, a := range listApp {
			if a.Uuid != uuid && a.Name == AppName {
				tre.app.Remove(a.Uuid)
				fmt.Printf("Application with uuid %s was removed from database\n", a.Uuid)
			}
		}
	})

	com = Command{
		AppId:       uuid,
		Cmd:         129,
		Type:        0,
		Descr:       "Command #129 (input)",
		TxtF:        true,
		TxtNum:      1,
		TxtDescr:    "<number int>",
		JsonF:       false,
		Json:        "",
		BinaryF:     false,
		BinaryDescr: "",
	}

	t.Run("Command Set(insert new command)", func(t *testing.T) {
		err = tre.com.Set(&com)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Command for app %s with cmd %d was set to database\n", uuid, com.Cmd)

		com.Type = 2
		com.Descr = "Command #129 (output)"
		err = tre.com.Set(&com)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Command for app %s with cmd %d was set to database\n", uuid, com.Cmd)
	})

	t.Run("Application Get", func(t *testing.T) {
		a, err := tre.app.Get(uuid)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was get from database:\n%v\n", uuid, a)
	})

	t.Run("Application Remove", func(t *testing.T) {
		err = tre.app.Remove(uuid)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Application with uuid %s was removed from database\n", uuid)
	})

	t.Run("Close", func(t *testing.T) {
		tre.Close()
	})
}

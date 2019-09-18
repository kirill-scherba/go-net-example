// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teoregistry (teo-registry: teonet registry service) package
//
// The teonet registry story teonet applications(services) and its command api
// definition.
// This package contain registry database schemas, service and clients functions
//
//
/* Install:
   go get github.com/kirill-scherba/teonet-go/services/teoregistry
*/
//
// Data base organisation (we use ScyllaDB):
//
// Run Scylla in Docker: https://www.scylladb.com/download/open-source/#docker
/* Before you execute the program running this services, Launch `cqlsh` and execute:
//
// Keyspace 'teoregistry'
CREATE KEYSPACE teoregistry WITH replication = { 'class': 'SimpleStrategy', 'replication_factor' : 3 };
USE teoregistry;
//
// Tables
// Table 'applications': Teonet applications (services) description
CREATE TABLE applications(
  uuid        TIMEUUID,
  name        TEXT,
  descr       TEXT,
  author      TEXT,
  license     TEXT,
  goget    		TEXT,
  git      		TEXT,
  PRIMARY KEY(uuid)
);
CREATE INDEX ON applications (name);
//
// Table 'commands': Teonet applications commands description
// - cmdType values:  0 - input; 1 - input/output (same parameters); 2 - output
CREATE TABLE commands(
  app_id       TIMEUUID,
  cmd          INT,
  type     		 TINYINT,
	descr  	     TEXT,
  txt_f        BOOLEAN,
  txt_num      TINYINT,
  txt_descr    TEXT,
  jsonf        BOOLEAN,
  json         TEXT,
  binary_f     BOOLEAN,
  binary_descr TEXT,
  PRIMARY KEY(app_id, cmd, type)
);

*/

package teoregistry

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

// Teoregistry is the teoregistry packet receiver
type Teoregistry struct {
	session *gocql.Session
	app     *App
	com     *Com
}

// App is application functions receiver
type App struct {
	tre *Teoregistry
}

// Com is command functions receiver
type Com struct {
	tre *Teoregistry
}

// Application is the Table 'applications': Teonet applications (services) description
type Application struct {
	UUID    gocql.UUID
	Name    string
	Descr   string
	Author  string
	License string
	Goget   string
	Git     string
	Com     []Command
}

// Applist is short application representation used in list function and commands
type Applist struct {
	UUID gocql.UUID
	Name string
}

// Command is the Table 'commands': Teonet applications commands description
// - cmdType values:  0 - input; 1 - input/output (same parameters); 2 - output
type Command struct {
	AppID       gocql.UUID
	Cmd         int
	Type        uint8
	Descr       string
	TxtF        bool
	TxtNum      uint8
	TxtDescr    string
	JSONF       bool
	JSON        string
	BinaryF     bool
	BinaryDescr string
}

// Connect to the cql cluster and create teoregistry receiver.
// First parameter is keyspace, next parameters is hosts name (usualy it should
// be 3 hosts - 3 ScyllaDB nodes)
func Connect(hosts ...string) (tre *Teoregistry, err error) {
	tre = &Teoregistry{}
	tre.app = &App{tre}
	tre.com = &Com{tre}
	var keyspace = "teoregistry"
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
	tre.session, _ = cluster.CreateSession()
	return
}

// Close closes cql connection and destroy teoregistry receiver
func (tre *Teoregistry) Close() {
	tre.session.Close()
}

// Set inserts or updates application info
func (app *App) Set(a *Application) (uuid gocql.UUID, err error) {
	stmt, names := qb.Update("applications").Set(
		"name", "descr", "author", "license", "goget", "git",
	).Where(qb.Eq("uuid")).ToCql()
	q := gocqlx.Query(app.tre.session.Query(stmt), names).BindStruct(a)
	if err = q.ExecRelease(); err != nil {
		fmt.Printf("List Error: %s\n", err.Error())
		return
	}
	uuid = a.UUID
	return
}

// Get gets application info
func (app *App) Get(uuid gocql.UUID) (a *Application, err error) {
	a = &Application{UUID: uuid}
	stmt, names := qb.Select("applications").Where(qb.Eq("uuid")).Limit(1).ToCql()
	q := gocqlx.Query(app.tre.session.Query(stmt), names).BindMap(qb.M{
		"uuid": uuid,
	})
	if err = q.GetRelease(a); err != nil {
		fmt.Printf("Get Error: %s\n", err.Error())
		return
	}
	a.Com, err = app.tre.com.List(uuid)
	return
}

// Remove removes application info and all its commands
func (app *App) Remove(uuid gocql.UUID) (err error) {
	if err = app.tre.session.Query(`DELETE FROM applications WHERE uuid = ?`,
		uuid).Exec(); err != nil {
		fmt.Printf("Remove Error: %s\n", err.Error())
		return
	}
	err = app.tre.com.RemoveAll(uuid)
	return
}

// Num gets number of applications
func (app *App) Num() (numApp int, err error) {
	if err := app.tre.session.Query(`SELECT count(*) FROM applications LIMIT 1`).
		Consistency(gocql.One).Scan(&numApp); err != nil {
		fmt.Printf("Num Error: %s\n", err.Error())
	}
	return
}

// List gets list of applications. Returns list of applications which contain
// apllication uuid and name
func (app *App) List() (listApp []Applist, err error) {
	stmt, names := qb.Select("applications").Columns("uuid", "name").ToCql()
	q := gocqlx.Query(app.tre.session.Query(stmt), names)
	if err = q.SelectRelease(&listApp); err != nil {
		fmt.Printf("List Commands Error: %s\n", err.Error())
	}
	return
}

// Set inserts or updates command info
func (com *Com) Set(c *Command) (err error) {
	stmt, names := qb.Update("commands").Set(
		"descr", "txt_f", "txt_num", "txt_descr", "jsonf", "json",
		"binary_f", "binary_descr",
	).Where(qb.Eq("app_id"), qb.Eq("cmd"), qb.Eq("type")).ToCql()
	q := gocqlx.Query(com.tre.session.Query(stmt), names).BindStruct(c)
	if err = q.ExecRelease(); err != nil {
		fmt.Printf("List Commands Error: %s\n", err.Error())
	}
	return
}

// Remove removes command
func (com *Com) Remove(appid gocql.UUID, cmd int, cmdtype int) (err error) {
	if err = com.tre.session.Query(`DELETE FROM commands
		WHERE app_id= ?, cmd = ?, type = ?`,
		appid, cmd, cmdtype).Exec(); err != nil {
		fmt.Printf("Remove Command Error: %s\n", err.Error())
	}
	return
}

// RemoveAll removes all application commands
func (com *Com) RemoveAll(appid gocql.UUID) (err error) {
	if err = com.tre.session.Query(`DELETE FROM commands WHERE app_id= ?`,
		appid).Exec(); err != nil {
		fmt.Printf("Remove All Commands Error: %s\n", err.Error())
	}
	return
}

// List gets list of commands
func (com *Com) List(appid gocql.UUID) (listCom []Command, err error) {
	stmt, names := qb.Select("commands").Where(qb.Eq("app_id")).ToCql()
	q := gocqlx.Query(com.tre.session.Query(stmt), names).BindMap(qb.M{
		"app_id": appid,
	})
	if err = q.SelectRelease(&listCom); err != nil {
		fmt.Printf("List Commands Error: %s\n", err.Error())
	}
	return
}

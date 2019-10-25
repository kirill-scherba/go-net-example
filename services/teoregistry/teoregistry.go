// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoregistry (teo-registry) is the Teonet registry service package
//
// Teoregistry store teonet applications(services) and its commands api
// description in teonet database. This package contain teonet registry
// database schemas, services and clients functions.
//
// Install this go package:
//   go get github.com/kirill-scherba/teonet-go/services/teoregistry
//
// Data base organisation
//
// To store database we use ScyllaDB. Run Scylla in Docker:
//   https://www.scylladb.com/download/open-source/#docker
//
// Install database schemas. Before you execute application which used this
// service, launch `cqlsh`:
//   docker exec -it scylla cqlsh
// and execute next commands:
/*
   // Keyspace 'teoregistry'
   CREATE KEYSPACE teoregistry WITH replication = { 'class': 'SimpleStrategy', 'replication_factor' : 3 };
   USE teoregistry;

   // Table 'applications': Teonet applications (services) description
   CREATE TABLE IF NOT EXISTS applications(
   uuid        TIMEUUID,
   name        TEXT,
   version     TEXT,
   descr       TEXT,
   author      TEXT,
   license     TEXT,
   goget       TEXT,
   git         TEXT,
   PRIMARY KEY(uuid)
   );
   CREATE INDEX IF NOT EXISTS ON applications (name);

   // Table 'commands': Teonet applications commands description
   // - cmdType values:  0 - input; 1 - input/output (same parameters); 2 - output
   CREATE TABLE IF NOT EXISTS commands(
   app_id       TIMEUUID,
   cmd          INT,
   type         TINYINT,
   descr        TEXT,
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
	"github.com/kirill-scherba/teonet-go/services/teoapi"
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

// Applist is short application representation used in list function and commands
type Applist struct {
	UUID gocql.UUID
	Name string
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
func (app *App) Set(a *teoapi.Application) (uuid gocql.UUID, err error) {
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
func (app *App) Get(uuid gocql.UUID) (a *teoapi.Application, err error) {
	a = &teoapi.Application{UUID: uuid}
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
func (com *Com) Set(c *teoapi.Command) (err error) {
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
func (com *Com) List(appid gocql.UUID) (listCom []teoapi.Command, err error) {
	stmt, names := qb.Select("commands").Where(qb.Eq("app_id")).ToCql()
	q := gocqlx.Query(com.tre.session.Query(stmt), names).BindMap(qb.M{
		"app_id": appid,
	})
	if err = q.SelectRelease(&listCom); err != nil {
		fmt.Printf("List Commands Error: %s\n", err.Error())
	}
	return
}

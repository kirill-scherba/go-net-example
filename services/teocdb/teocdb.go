// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdb (teo-cdb) is the Teonet database service package
//
// Install this go package:
//   go get github.com/kirill-scherba/teonet-go/services/teoregistry
//
// Data base organisation
//
// This service uses ScyllaDB. If you install this service manually you need
// install ScyllaDB. Run Scylla in Docker:
//
//   https://www.scylladb.com/download/open-source/#docker
//
// Before you execute application which used this package you need install
// database schemas. Launch `cqlsh`:
//   docker exec -it scylla cqlsh
// and execute next commands:
/*
	create keyspace teocdb with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };
	create table teocdb.map(key text, data blob, PRIMARY KEY(key));
*/
//
package teocdb

import (
	"fmt"

	"github.com/gocql/gocql"
)

// Teocdb is teocdb packet receiver
type Teocdb struct {
	Session *gocql.Session
}

// Connect to the cql cluster and return teocdb receiver
func Connect(hosts ...string) (tdb *Teocdb, err error) {
	tdb = &Teocdb{}
	cluster := gocql.NewCluster(func() (h []string) {
		if h = hosts; len(h) == 0 {
			h = []string{"172.17.0.2", "172.17.0.3", "172.17.0.4"}
		}
		return
	}()...)
	cluster.Keyspace = "teocdb"
	cluster.Consistency = gocql.Quorum
	tdb.Session, _ = cluster.CreateSession()

	// Create keyspace and table
	const mapSchema = `
		// create KEYSPACE IF NOT EXISTS teocdb WITH replication = {
		// 	'class' : 'SimpleStrategy',
		// 	'replication_factor' : 3
		// };
		create TABLE IF NOT EXISTS teocdb.map(
			key text,
			data blob,
			PRIMARY KEY(key)
		)`
	if err = tdb.execStmt(tdb.Session, mapSchema); err != nil {
		//t.Fatal("create table:", err)
	}
	return
}

// ExecStmt executes a statement string.
func (tdb *Teocdb) execStmt(s *gocql.Session, stmt string) error {
	q := s.Query(stmt).RetryPolicy(nil)
	defer q.Release()
	return q.Exec()
}

// Close teocdb connection
func (tdb *Teocdb) Close() {
	tdb.Session.Close()
}

// Update key value
func (tdb *Teocdb) Update(key string, value []byte) (err error) {
	if err = tdb.Session.Query(`UPDATE map SET data = ? WHERE key = ?`,
		value, key).Exec(); err != nil {
		fmt.Printf("Insert Error: %s\n", err.Error())
	}
	return
}

// Get value by key
func (tdb *Teocdb) Get(key string) (data []byte, err error) {
	if err := tdb.Session.Query(`SELECT data FROM map WHERE key = ? LIMIT 1`,
		key).Consistency(gocql.One).Scan(&data); err != nil {
		fmt.Printf("Get Error: %s\n", err.Error())
	}
	return
}

// List read and return array of all keys connected to selected key
func (tdb *Teocdb) List(key string) (keyAr []string, err error) {
	var keyOut string
	iter := tdb.Session.Query(`
		SELECT key FROM map WHERE key >= ? and key < ?
		ALLOW FILTERING`,
		key, key+"a").Iter()
	for iter.Scan(&keyOut) {
		fmt.Println("key:", keyOut)
		keyAr = append(keyAr, keyOut)
	}
	return
}

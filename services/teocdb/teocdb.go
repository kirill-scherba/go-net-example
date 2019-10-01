// Copyright 2019 teonet-go authors.  All rights reserved.
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
	"encoding/binary"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

type Teocdb struct {
	Session *gocql.Session
}

type JData struct {
	Key   string      `json:"key"`
	Id    interface{} `json:"id"`
	Value interface{} `json:"value"`
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

func (tdb *Teocdb) Close() {
	tdb.Session.Close()
}

func Send(teo *teonet.Teonet, key string, value []byte) {
	//fmt.Println("Marshal(key, value):", Marshal(key, value))
	teo.SendTo("teo-cdb", 129, Marshal(key, value))
}

// Marshal pack key value to data byte array
func Marshal(key string, value []byte) (data []byte) {
	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, uint32(len(key)))
	data = append(append(l, []byte(key)...), value...)
	return
}

// Unmarshal unpack key value from data byte array
func Unmarshal(data []byte) (key string, value []byte) {
	l := binary.LittleEndian.Uint32(data)
	key = string(data[:l])
	value = data[l:]
	return
}

func (tdb *Teocdb) Update(key string, value []byte) (err error) {
	if err = tdb.Session.Query(`UPDATE map SET data = ? WHERE key = ?`,
		value, key).Exec(); err != nil {
		fmt.Printf("Insert Error: %s\n", err.Error())
	}
	return
}

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

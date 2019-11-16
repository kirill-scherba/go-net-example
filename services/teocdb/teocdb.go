// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdb (teo-cdb) is the Teonet key-value database service package
//
// Install this go package:
//
//   go get github.com/kirill-scherba/teonet-go/services/teocdb
//
// Data base organisation
//
// This service uses ScyllaDB. If you install this service manually you need
// install ScyllaDB. Run Scylla in Docker:
//
//   https://www.scylladb.com/download/open-source/#docker
//
// You may check version of your existing running scylla docker container with
// command:
//
//   docker exec -it scylla scylla --version
//
// Before you execute application which used this package you need install
// database schemas. Launch `cqlsh`:
//
//   docker exec -it scylla cqlsh
//
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
	"github.com/kirill-scherba/teonet-go/services/teoapi"
	cdb "github.com/kirill-scherba/teonet-go/services/teocdbcli"
)

// HostsDefault is default hosts IPs
var HostsDefault = []string{"172.19.0.2", "172.19.0.3", "172.19.0.4"}

// Teocdb is teocdb packet receiver
type Teocdb struct {
	session *gocql.Session
	*Process
	con cdb.TeoConnector
}

// Connect to the cql cluster and return teocdb receiver
func Connect(con cdb.TeoConnector, hosts ...string) (tcdb *Teocdb, err error) {
	keyspace := "teocdb"
	tcdb = &Teocdb{con: con}
	tcdb.Process = &Process{tcdb}
	cluster := gocql.NewCluster(func() (h []string) {
		if h = hosts; len(h) > 0 {
			keyspace = h[0]
			h = h[1:]
		}
		return
	}()...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	tcdb.session, _ = cluster.CreateSession()

	// Create table
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
	if err = tcdb.execStmt(tcdb.session, mapSchema); err != nil {
		//t.Fatal("create table:", err)
	}
	return
}

// ExecStmt executes a statement string.
func (tcdb *Teocdb) execStmt(s *gocql.Session, stmt string) error {
	q := s.Query(stmt).RetryPolicy(nil)
	defer q.Release()
	return q.Exec()
}

// Close teocdb connection
func (tcdb *Teocdb) Close() {
	tcdb.session.Close()
}

// Set key value
func (tcdb *Teocdb) Set(key string, value []byte) (err error) {
	if err = tcdb.session.Query(`UPDATE map SET data = ? WHERE key = ?`,
		value, key).Exec(); err != nil {
		fmt.Printf("Insert(or update) key '%s' Error: %s\n", key, err.Error())
	}
	return
}

// Get value by key, returns key value or empty data if key not found
func (tcdb *Teocdb) Get(key string) (data []byte, err error) {
	// Does not return err of tcdb.session.Query function
	if err := tcdb.session.Query(`SELECT data FROM map WHERE key = ? LIMIT 1`,
		key).Consistency(gocql.One).Scan(&data); err != nil {
		fmt.Printf("Get key '%s' Error: %s\n", key, err.Error())
	}
	return
}

// List read and return array of all keys starts from selected key
func (tcdb *Teocdb) List(key string) (keyList cdb.KeyList, err error) {
	var keyOut string
	iter := tcdb.session.Query(`
		SELECT key FROM map WHERE key >= ? and key < ?
		ALLOW FILTERING`,
		key, key+"a").Iter()
	for iter.Scan(&keyOut) {
		keyList.Append(keyOut)
	}
	return
}

// Process receiver to process teocdb commands
type Process struct{ tcdb *Teocdb }

// CmdBinary process CmdBinary command
func (p *Process) CmdBinary(pac teoapi.Packet) (err error) {
	var request, responce cdb.KeyValue
	err = request.UnmarshalBinary(pac.Data())
	if err != nil {
		return
	}
	responce = request
	switch request.Cmd {
	case cdb.CmdSet:
		if err = p.tcdb.Set(request.Key, request.Value); err != nil {
			return
		}
		responce.Value = nil

	case cdb.CmdGet:
		if responce.Value, err = p.tcdb.Get(request.Key); err != nil {
			return
		}

	case cdb.CmdList:
		var keys cdb.KeyList
		if keys, err = p.tcdb.List(request.Key); err != nil {
			return
		}
		responce.Value, _ = keys.MarshalBinary()
	}

	if retdata, err := responce.MarshalBinary(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// CmdSet process CmdSet command
func (p *Process) CmdSet(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if err = p.tcdb.Set(request.Key, request.Value); err != nil {
		return
	}
	// Return only Value for text requests and all fields for json
	responce := request
	responce.Value = nil
	if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// CmdGet process CmdGet command
func (p *Process) CmdGet(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	}
	// Return only Value for text requests and all fields for json
	responce := request
	if responce.Value, err = p.tcdb.Get(request.Key); err != nil {
		return
	} else if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// CmdList process CmdList command
func (p *Process) CmdList(pac teoapi.Packet) (err error) {
	var keys cdb.KeyList
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if keys, err = p.tcdb.List(request.Key); err != nil {
		return
	}
	// Return only Value for text requests and all fields for json
	responce := request
	responce.Value, err = keys.MarshalJSON()
	if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

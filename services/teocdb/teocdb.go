// Copyright 2019,2021 Teonet-go authors.  All rights reserved.
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
	create table teocdb.ids(id_name text, next_id int, PRIMARY KEY(id_name));
	create table teocdb.queue(key text, time timestamp, lock text, data blob, PRIMARY KEY(key, time));
*/
//
package teocdb

import (
	"encoding/json"
	"fmt"
	"log"
	"plugin"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoapi"
	cdb "github.com/kirill-scherba/teonet-go/services/teocdbcli"
)

// HostsDefault is default hosts IPs
var HostsDefault = []string{
	"172.17.0.2", "172.17.0.3", "172.17.0.4",
	"172.18.0.2", "172.18.0.3", "172.18.0.4",
}

// KeyspaceDefault is default keyspace bane
var KeyspaceDefault = "teocdb"

// Teocdb is teocdb packet receiver
type Teocdb struct {
	session *gocql.Session
	*Process
	con cdb.TeoConnector
}

// Connect to the cql cluster and return teocdb receiver
// func Connect(con cdb.TeoConnector, hosts ...string) (tcdb *Teocdb, err error) {
// 	keyspace := "teocdb"
// 	tcdb = &Teocdb{con: con}
// 	tcdb.Process = &Process{tcdb}
// 	cluster := gocql.NewCluster(func() (h []string) {
// 		if h = hosts; len(h) > 0 {
// 			keyspace = h[0]
// 			h = h[1:]
// 		}
// 		return
// 	}()...)
// 	cluster.Keyspace = keyspace
// 	cluster.Consistency = gocql.Quorum
// 	tcdb.session, _ = cluster.CreateSession()

// 	// Create table
// 	const mapSchema = `
// 		// create KEYSPACE IF NOT EXISTS teocdb WITH replication = {
// 		// 	'class' : 'SimpleStrategy',
// 		// 	'replication_factor' : 3
// 		// };
// 		create TABLE IF NOT EXISTS teocdb.map(
// 			key text,
// 			data blob,
// 			PRIMARY KEY(key)
// 		)`
// 	if err = tcdb.execStmt(tcdb.session, mapSchema); err != nil {
// 		//t.Fatal("create table:", err)
// 	}
// 	return
// }

// Connect to the cql cluster and return teocdb receiver
func Connect(con cdb.TeoConnector, keyAndHosts ...string) (tcdb *Teocdb, err error) {
	keyspace := KeyspaceDefault
	tcdb = &Teocdb{con: con}
	tcdb.Process = &Process{tcdb}
	cluster := gocql.NewCluster(func() (h []string) {
		// Get keyspace from keyAndHosts input array
		if len(keyAndHosts) > 0 {
			keyspace = keyAndHosts[0]
		}
		// Get hosts name from keyAndHosts input array or set default hosts name if input array empty
		if len(keyAndHosts) > 1 {
			h = keyAndHosts[1:]
		} else {
			h = HostsDefault
		}
		return
	}()...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	if tcdb.session, err = cluster.CreateSession(); err != nil {
		return
	}

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
	}
	return
}

// Get value by key, returns key value or empty data if key not found
func (tcdb *Teocdb) Get(key string) (data []byte, err error) {
	// Does not return err of tcdb.session.Query function
	if err = tcdb.session.Query(`SELECT data FROM map WHERE key = ? LIMIT 1`,
		key).Consistency(gocql.One).Scan(&data); err != nil {
	}
	return
}

// Delete record from database by key, returns
func (tcdb *Teocdb) Delete(key string) (err error) {
	// Does not return err of tcdb.session.Query function
	if err = tcdb.session.Query(`DELETE data FROM map WHERE key = ?`,
		key).Exec(); err != nil {
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

// ListBody read and return array of all keys data starts from selected key
func (tcdb *Teocdb) ListBody(key string) (dataList []string, err error) {
	var dataOut string
	iter := tcdb.session.Query(`
		SELECT data FROM map WHERE key >= ? and key < ?
		ALLOW FILTERING`,
		key, key+"a").Iter()
	for iter.Scan(&dataOut) {
		dataList = append(dataList, dataOut)
	}
	return
}

// Func execute plugin function and return data
func (tcdb *Teocdb) Func(key string, value []byte) (data []byte, err error) {
	return tcdb.PluginFunc(key, value)
}

// SetID set keys next ID value
func (tcdb *Teocdb) SetID(key string, value []byte) (err error) {
	nextID, err := strconv.Atoi(string(value))
	if err != nil {
		return
	}
	return tcdb.session.Query(`UPDATE ids SET next_id = ? WHERE id_name = ?`,
		nextID, key).Exec()
}

// GetID returns new diginal ID for key, ID just increments
func (tcdb *Teocdb) GetID(key string) (data []byte, err error) {
	var nextID int
	// Read current counter value with id_name
	if err = tcdb.session.Query(`SELECT next_id FROM ids WHERE id_name = ? LIMIT 1`,
		key).Consistency(gocql.One).Scan(&nextID); err != nil {

		// Check error
		if err != gocql.ErrNotFound {
			log.Println("Read current counter error:", err)
			return
		}

		// Create new record if counter with id_name does not exists
		nextID = 1
		if err = tcdb.SetID(key, []byte(strconv.Itoa(nextID))); err != nil {
			return
		}
	}

	// Set new next_id and return current next_id
	var ok bool
	for {
		// Increment nextID in database
		if err = tcdb.session.Query(`UPDATE ids SET next_id = ? WHERE id_name = ? IF next_id = ?`,
			nextID+1, key, nextID).Scan(&ok, &nextID); err != nil {
			log.Println("Increment current counter error:", err)
			return
		}
		// log.Println("Update result:", ok, nextID)
		if ok {
			break
		}
	}

	// Return received nextID in text
	data = []byte(fmt.Sprintf("%d", nextID))
	return
}

// DeleteID delete counter from database by key
func (tcdb *Teocdb) DeleteID(key string) (err error) {
	return tcdb.session.Query(`DELETE FROM ids WHERE id_name = ?`,
		key).Exec()
}

// TODO: SetQueue
// UPDATE queue SET data = textAsBlob('Hello19') WHERE key = 'queue/first' AND time = toTimestamp(now());

// TODO: GetQueue
// SELECT time FROM queue WHERE key = 'queue/first' AND lock = '' LIMIT 1  ALLOW FILTERING ;
// UPDATE queue SET lock = 'locked' WHERE key = 'queue/first' AND time = '2021-12-12 22:01:52.301' IF lock = null;
// DELETE FROM queue WHERE key = 'queue/first' AND time = '2021-12-12 22:01:52.301';

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
			const notFound = "not found"
			if err.Error() != notFound {
				return
			}
			responce.Err = err.Error()
		}

	case cdb.CmdList:
		var keys cdb.KeyList
		if keys, err = p.tcdb.List(request.Key); err != nil {
			return
		}
		responce.Value, _ = keys.MarshalBinary()

	case cdb.CmdListBody:
		var datas []string
		if datas, err = p.tcdb.ListBody(request.Key); err != nil {
			return
		}
		responce.Value, _ = json.Marshal(datas)

	case cdb.CmdDelete:
		if err = p.tcdb.Delete(request.Key); err != nil {
			return
		}
		responce.Value = nil

	case cdb.CmdFunc:
		if responce.Value, err = p.tcdb.PluginFunc(request.Key, request.Value); err != nil {
			return
		}

	case cdb.CmdGetID:
		if responce.Value, err = p.tcdb.GetID(request.Key); err != nil {
			return
		}

	case cdb.CmdSetID:
		if err = p.tcdb.SetID(request.Key, request.Value); err != nil {
			return
		}
		responce.Value = nil

	case cdb.CmdDeleteID:
		if err = p.tcdb.DeleteID(request.Key); err != nil {
			return
		}
		responce.Value = nil
	}

	retdata, err := responce.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
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

// CmdDelete process CmdDelete command
func (p *Process) CmdDelete(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if err = p.tcdb.Delete(request.Key); err != nil {
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

// CmdListBody process CmdListBody command
func (p *Process) CmdListBody(pac teoapi.Packet) (err error) {
	var datas []string
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if datas, err = p.tcdb.ListBody(request.Key); err != nil {
		return
	}
	// Return all fields in json
	responce := request
	responce.Value, err = json.Marshal(datas)

	if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// CmdFunc process CmdFunc command
func (p *Process) CmdFunc(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	}
	// Return only Value for text requests and all fields for json
	responce := request
	if responce.Value, err = p.tcdb.Func(request.Key, request.Value); err != nil {
		return
	} else if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// PluginFuncType define plugin function type
// type PluginFuncType func(params ...string) (data []byte, err error)

// PluginFunc process plugin function: plugin_name.func(parameters ...string)
func (tcdb *Teocdb) PluginFunc(fff string, value []byte) (data []byte, err error) {

	d := cdb.Plugin{}
	d.UnmarshalBinary(value)

	p, err := plugin.Open("/root/plugin/" + d.Name + ".so")
	if err != nil {
		return
	}

	f, err := p.Lookup(d.Func)
	if err != nil {
		return
	}

	return f.(func(*Teocdb, ...string) ([]byte, error))(tcdb, d.Params...)
}

// CmdGetID process CmdGetID command; next id number created for
func (p *Process) CmdGetID(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	}
	// Return only Value for text requests and all fields for json
	responce := request
	if responce.Value, err = p.tcdb.GetID(request.Key); err != nil {
		return
	} else if !request.RequestInJSON {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), responce.Value)
	} else if retdata, err := responce.MarshalText(); err == nil {
		_, err = p.tcdb.con.SendAnswer(pac, pac.Cmd(), retdata)
	}
	return
}

// CmdSetID process CmdSetID command; next id number created for
func (p *Process) CmdSetID(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if err = p.tcdb.SetID(request.Key, request.Value); err != nil {
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

// CmdDeleteID process CmdDeleteID command
func (p *Process) CmdDeleteID(pac teoapi.Packet) (err error) {
	data := pac.RemoveTrailingZero(pac.Data())
	request := cdb.KeyValue{Cmd: pac.Cmd()}
	if err = request.UnmarshalText(data); err != nil {
		return
	} else if err = p.tcdb.DeleteID(request.Key); err != nil {
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

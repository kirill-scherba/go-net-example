// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teousers (teo-users) is the Teonet users service package.
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
// Before you execute application which used this package you need install
// database schemas. Launch `cqlsh`:
//
//   docker exec -it scylla cqlsh
//
// and execute next commands:
/*
 CREATE KEYSPACE IF NOT EXISTS teousers with replication = { 'class' : 'SimpleStrategy',
 'replication_factor' : 3 };
 USE teousers;
 CREATE TABLE IF NOT EXISTS  users (
   id uuid,
   access_token uuid,
   prefix text,
   name text,
   avatar_id uuid,
   gravatar_id text,
   online boolean,
   last_online timestamp,
   state int,
   PRIMARY KEY (id)
 );
 CREATE INDEX IF NOT EXISTS ON users (prefix);
 CREATE INDEX IF NOT EXISTS ON users (name);
 CREATE INDEX IF NOT EXISTS ON users (online);
*/
// To run db tests repeat the same with teousers_test in first string.
//
package teousers

// Users is the teousers data structure and methods receiver
type Users struct {
	*db
	*Process
	TeoConnector
}

// Connect to the cql cluster and create teousers receiver.
// First parameter is keyspace, next parameters is hosts name (usualy it should
// be 3 hosts - 3 ScyllaDB nodes)
func Connect(con TeoConnector, hosts ...string) (u *Users, err error) {
	u = &Users{TeoConnector: con}
	u.db, err = newDb(hosts...)
	u.Process = &Process{u}
	return
}

// Close closes cql connection and destroy teoregistry receiver
func (u *Users) Close() {
	u.db.close()
}

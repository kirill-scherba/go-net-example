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
 CREATE TABLE IF NOT EXISTS users (
   user_id uuid,
   access_token uuid,
   user_name text,
   avatar_id uuid,
   gravatar_id text,
   online boolean,
   last_online timestamp,
   PRIMARY KEY (user_id)
 );
 CREATE INDEX IF NOT EXISTS ON users (online);
*/
// To run db tests repeat the same with teousers_test in first string.
//
package teousers

import "github.com/gocql/gocql"

// Users is the teousers data structure and methods receiver
type Users struct {
	session *gocql.Session
	*db
}

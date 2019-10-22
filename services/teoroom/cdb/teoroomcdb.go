// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cdb (teoroomcdb) is the Teonet room cdb functions and datbase themes.
//
// Teoroomcdb provide teroom database functions executed in cdb.
//
// Install this go package:
//
//   go get github.com/kirill-scherba/teonet-go/services/teoroom/cdb
//
// Data base organisation
//
// To store database we use ScyllaDB. Run Scylla in Docker:
//   https://www.scylladb.com/download/open-source/#docker
//
// Install database schemas. Before you execute application which used this
// service, launch `cqlsh`:
//
//   docker exec -it scylla cqlsh
//
// and execute content of cql/teoroom.cql file.
//
// For tests repeat the same instructions but use teoroom_test keyspace.
//
package cdb

import "github.com/kirill-scherba/teonet-go/services/teoroomcli/cdb"

// Rooms is the teoroomcdb data structure and methods receiver
type Rooms struct {
	*db
	*Process
	cdb.TeoConnector
}

// Connect to the cql cluster and create teoroomcdb receiver.
// First parameter is keyspace, next parameters is hosts name (usualy it should
// be 3 hosts - 3 ScyllaDB nodes)
func Connect(con cdb.TeoConnector, hosts ...string) (r *Rooms, err error) {
	r = &Rooms{TeoConnector: con}
	r.db, err = newDb(hosts...)
	r.Process = &Process{r}
	return
}

// Close closes cql connection and destroy teoregistry receiver
func (r *Rooms) Close() {
	r.db.close()
}

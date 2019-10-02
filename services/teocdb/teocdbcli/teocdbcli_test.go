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
package teocdbcli

import "testing"

func TestTeocdb(t *testing.T) {

	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {

		key := "test.key.123"
		id := 1
		value := []byte("Hello world!")

		bd := &BinaryData{key, id, value}
		data, err := bd.MarshalBinary()
		if err != nil {
			t.Error(err)
			return
		}
		if err = bd.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}
		if bd.Key != key || bd.ID != id || string(value) != string(bd.Value) {
			t.Errorf(
				"unmarshalled structure fields values not equal to input values: %s, %d, %s",
				bd.Key, bd.ID, bd.Value,
			)
		}
	})

}

// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers db tests module.
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
 CREATE KEYSPACE IF NOT EXISTS teousers_test with replication = { 'class' : 'SimpleStrategy',
 'replication_factor' : 3 };
 USE teousers_test;
 CREATE TABLE IF NOT EXISTS  users (
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

package teousers

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func TestDB(t *testing.T) {

	const AppName = "teotest-7755-2"
	userID := gocql.TimeUUID()
	var err error
	var u *Users

	t.Run("Connect", func(t *testing.T) {
		u, err = Connect(nil, "teousers_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer u.Close()

	t.Run("Set", func(t *testing.T) {
		accessToken := gocql.TimeUUID()
		user := &User{
			UserID:      userID,
			AccessToken: accessToken,
			UserName:    "Test-1",
			LastOnline:  time.Now(),
		}
		fmt.Println("set user_id:", user.UserID)
		err = u.db.set(user)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Get", func(t *testing.T) {
		user := &User{UserID: userID}
		fmt.Println("get user_id:", user.UserID)
		err = u.db.get(user)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(user)
	})

	t.Run("Delete", func(t *testing.T) {
		user := &User{UserID: userID}
		fmt.Println("delete user_id:", user.UserID)
		err = u.db.delete(user)
		if err != nil {
			t.Error(err)
			return
		}
	})

}

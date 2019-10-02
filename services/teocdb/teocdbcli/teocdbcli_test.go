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

import (
	"fmt"
	"testing"

	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func TestTeocdbcli(t *testing.T) {

	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {

		cmd := byte(1)
		id := uint16(2)
		key := "test.key.123"
		value := []byte("Hello world!")

		bdInput := &BinaryData{cmd, id, key, value}
		bdOutput := &BinaryData{}
		data, err := bdInput.MarshalBinary()
		//fmt.Println(data)
		if err != nil {
			t.Error(err)
			return
		}
		if err = bdOutput.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}
		if bdOutput.Cmd != cmd || bdOutput.ID != id || bdOutput.Key != key ||
			string(value) != string(bdOutput.Value) {
			t.Errorf("unmarshalled structure fields values"+
				" not equal to input values:\n%d, %d, '%s', '%s'\n"+
				"data: %v",
				bdOutput.Cmd, bdOutput.ID, bdOutput.Key, string(bdOutput.Value),
				data,
			)
			return
		}
		fmt.Printf(
			"data: %v\n"+"cmd: %d\n"+"id: %d\n"+"key: %s\n"+"value: %s\n",
			data, bdOutput.Cmd, bdOutput.ID, bdOutput.Key, string(bdOutput.Value))
	})

}

// Connect to teo-cdb and test Send function.
// To execute this test use command line parameters: go test -args teo-test
func TestTeocdbcliSend(t *testing.T) {
	ch := make(chan bool)
	version := "1.0.0"
	param := teonet.Params()
	param.Name = "teo-test"
	param.LogLevel = "NONE"
	teo := teonet.Connect(param, []string{"teo-go", "teo-test"}, version)
	go teo.Run(func(teo *teonet.Teonet) {
		for ev := range teo.Event() {
			switch ev.Event {
			case teonet.EventConnected:
				pac := ev.Data
				//fmt.Printf("Event Connected from: %s\n", pac.From())
				if pac.From() == "teo-cdb" {
					go func() {
						cdb := NewTeocdbcli(teo)
						fmt.Printf("send data\n")
						data, err := cdb.Send(CmdSet, "test.key.919", []byte("Hello!"))
						if err != nil {
							fmt.Printf("got error: %s\n", err.Error())
							ch <- false
							return
						}
						fmt.Printf("got data: %v\n", data)
						ch <- true
					}()
				}
			}
		}
	})
	retval := <-ch
	fmt.Printf("got retval %v\n", retval)
	//os.Exit(0)
	//teo.Close()
}

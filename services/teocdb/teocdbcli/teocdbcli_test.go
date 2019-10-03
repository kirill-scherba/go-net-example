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
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func TestKeyValueBinary(t *testing.T) {

	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {

		cmd := byte(1)
		id := uint16(2)
		key := "test.key.123"
		value := []byte("Hello world!")

		bdInput := &KeyValue{cmd, id, key, value, false}
		bdOutput := &KeyValue{}
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
func TestKeyValueText(t *testing.T) {
	t.Run("MarshalUnmarshalText", func(t *testing.T) {

		id := uint16(77)
		key := "test.key.5678"
		value := []byte("Hello!")

		type fields struct {
			Cmd           byte
			ID            uint16
			Key           string
			Value         []byte
			requestInJSON bool
		}
		type args struct {
			text []byte
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			// TODO: Add test cases.
			{"text-key", fields{Key: key}, args{[]byte(key)}, false},
			{"text-key-id-", fields{Key: key, ID: id}, args{[]byte(key + "," + strconv.Itoa(int(id)) + ",")}, false},
			{"text-key-value", fields{Key: key, Value: value}, args{[]byte(key + "," + string(value))}, false},
			{"text-key-id-value", fields{Key: key, ID: id, Value: value}, args{[]byte(key + "," + strconv.Itoa(int(id)) + "," + string(value))}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				kv := &KeyValue{}

				// Cmd:           tt.fields.Cmd,
				// ID:            tt.fields.ID,
				// Key:           tt.fields.Key,
				// Value:         tt.fields.Value,
				// requestInJSON: tt.fields.requestInJSON,

				if err := kv.UnmarshalText(tt.args.text); (err != nil) != tt.wantErr {
					t.Errorf("KeyValue.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				}
				if isErr := (kv.Key != tt.fields.Key); isErr != tt.wantErr {
					t.Errorf("Key not right ")
				}
				if isErr := (string(kv.Value) != string(tt.fields.Value)); isErr != tt.wantErr {
					t.Errorf("Value not right ")
				}
				if isErr := (kv.ID != tt.fields.ID); isErr != tt.wantErr {
					t.Errorf("Id not right ")
				}
				fmt.Printf("text: %s ->  KeyValue%v\n", string(tt.args.text), kv)
			})
		}
	})
}

// Connect to teo-cdb and test Send function.
// To execute this test use command line parameters: go test -args teo-test
func TestSend(t *testing.T) {
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
				// Start cdb test after peer teo-cdb connected to this test
				if pac.From() == "teo-cdb" {
					go func() {
						// Create new teonet teocdb client
						cdb := NewTeocdbCli(teo)

						fmt.Printf("connected to %s\n", pac.From())

						// Key value which will be used in test
						key := "test.key.919"
						value := "Hello world-919-23! - " + fmt.Sprint(time.Now())

						// Error processing function
						error := func(err error) {
							fmt.Printf("got error: %s\n", err.Error())
							t.Error(err)
							ch <- false
						}

						// Set data - Save {key,value} data to DB
						fmt.Printf("Set data: %s\n", value)
						data, err := cdb.Send(CmdSet, key, []byte(value))
						if err != nil {
							error(err)
							return
						}
						fmt.Printf("got answer data: %v\n", data)

						// Get data - Read data from DB by key
						fmt.Printf("Get data\n")
						data, err = cdb.Send(CmdGet, key, nil)
						if err != nil {
							error(err)
							return
						}
						fmt.Printf("got answer data: %s\n", string(data))
						if string(data) != value {
							error(errors.New("got wrong answer data"))
							return
						}

						// Get list of keys - Read array of keys with common prefix
						fmt.Printf("Get data\n")
						data, err = cdb.Send(CmdList, "test.key.", nil)
						if err != nil {
							error(err)
							return
						}
						//keylist, err := cdb.Keys(data)
						var keylist KeyList
						keylist.UnmarshalBinary(data)
						fmt.Printf("got answer keys: %v\n%v\n", keylist, data)

						// Get values of all keys received in previous example
						for _, key := range keylist.Keys() {
							data, err = cdb.Send(CmdGet, key, nil)
							if err != nil {
								error(err)
								return
							}
							fmt.Printf("got answer key, value: %s -> %s\n", key,
								string(data))
						}

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

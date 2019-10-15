// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main is integration test for Send function of the teocdbcli package.
//
// **Note: This test can't be implemented as part of teocdbcli package because
// this test uses teonet and teonet uses the teocdbcli package. And when we
// trying run this test under the teocdbcli package this error hapends:
// Import cycle not allowed.
//
package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teocdb/teocdbcli"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Connect to teo-cdb and test Send function.
// To execute this test use command line parameters: go test -args teo-test
func TestSend(t *testing.T) {
	ch := make(chan bool)
	version := "1.0.0"
	param := teonet.Params()
	param.Name = "teo-test"
	param.LogLevel = "NONE"
	param.L0wsAllow = false
	teo := teonet.Connect(param, []string{"teo-go", "teo-test"}, version)
	testSend := func(pac *teonet.Packet) {
		// Create new teonet teocdb client
		cdb := teocdbcli.New(teo)

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
		data, err := cdb.Send(teocdbcli.CmdSet, key, []byte(value))
		if err != nil {
			error(err)
			return
		}
		fmt.Printf("got answer data: %v\n", data)

		// Get data - Read data from DB by key
		fmt.Printf("Get data\n")
		data, err = cdb.Send(teocdbcli.CmdGet, key)
		if err != nil {
			error(err)
			return
		}
		fmt.Printf("got answer data: %s\n", string(data))
		if string(data) != value {
			error(errors.New("got wrong answer data"))
			return
		}

		// Get not existing key
		fmt.Printf("Get not existing data\n")
		data, err = cdb.Send(teocdbcli.CmdGet, "a.not.existing.key.test")
		if err != nil {
			error(err)
			return
		}
		fmt.Printf("got answer data: %v\n", data)

		// Get list of keys - Read array of keys with common prefix
		fmt.Printf("Get list\n")
		data, err = cdb.Send(teocdbcli.CmdList, "test.key.")
		if err != nil {
			error(err)
			return
		}
		//keylist, err := cdb.Keys(data)
		var keylist teocdbcli.KeyList
		keylist.UnmarshalBinary(data)
		fmt.Printf("got answer keys: %v\n%v\n", keylist, data)

		// Get values of all keys received in previous example
		for _, key := range keylist.Keys() {
			data, err = cdb.Send(teocdbcli.CmdGet, key)
			if err != nil {
				error(err)
				return
			}
			fmt.Printf("got answer key, value: %s -> %s\n", key,
				string(data))
		}

		ch <- true
	}
	go teo.Run(func(teo *teonet.Teonet) {
		for ev := range teo.Event() {
			switch ev.Event {
			case teonet.EventConnected:
				pac := ev.Data
				// Start cdb test after peer teo-cdb connected to this app
				if pac.From() == "teo-cdb" {
					go testSend(pac)
				}
			}
		}
	})
	retval := <-ch
	fmt.Printf("got retval %v\n", retval)
	//os.Exit(0)
	//teo.Close()
}

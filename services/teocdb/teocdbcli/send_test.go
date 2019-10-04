// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocdbcli

import (
	"errors"
	"fmt"
	"testing"
	"time"

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

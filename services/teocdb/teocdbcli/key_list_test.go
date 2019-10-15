// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocdbcli

import (
	"errors"
	"fmt"
	"testing"
)

func TestKeyList(t *testing.T) {

	keys := []string{"Hello-1", "Hello-2", "Hello-3", "Hello-4"}

	t.Run("Append", func(t *testing.T) {
		var keyList1 KeyList
		keyList1.Append(keys...)

		keysResult := keyList1.Keys()
		if len(keys) != len(keysResult) {
			t.Error(errors.New("wrong number of keys after Applay"))
			return
		}
		for i, key := range keys {
			if key != keysResult[i] {
				t.Error(errors.New("wrong keys in array after Applay"))
				return
			}
		}
	})

	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {
		var keyList1, keyList2 KeyList
		keyList1.Append(keys...)
		data, err := keyList1.MarshalBinary()
		if err != nil {
			t.Error(err)
			return
		}
		keyList1.Keys()[0] = "xyz"
		keyList1.Keys()[3] = "xyz"
		fmt.Println("before unmarshal\nkeyList1:", keyList1.Keys())
		err = keyList1.UnmarshalBinary(data)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("after unmarshal\nkeyList1:", keyList1.Keys())

		err = keyList2.UnmarshalBinary(data)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("after unmarshal\nkeyList2:", keyList2.Keys())

		keysResult := keyList1.Keys()
		if len(keys) != len(keysResult) {
			t.Error(errors.New("wrong number of keys after Unmarshal"))
			return
		}
		for i, key := range keys {
			if key != keysResult[i] {
				t.Error(errors.New("wrong keys in array after Unmarshal"))
				return
			}
		}
	})
}

// This is Send function examples used in documentation ------------------------

// Key value which will be used in example
var key = "test.key.919"
var value = "Hello world!"
var answer []byte

// Create new teonet connector emulator
type teoEmu struct{}

func (c *teoEmu) SendTo(peer string, cmd byte, data []byte) (int, error) {
	return 0, errors.New("emulated")
}
func (c *teoEmu) SendAnswer(pac interface{}, cmd byte, data []byte) (int, error) {
	return 0, nil
}
func (c *teoEmu) WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	Data []byte
	Err  error
} {
	return nil
}

var con = &teoEmu{}
var cdb = New(con)

// Send CmdSet (set value) command - save key and value in teo-cdb example.
func ExampleTeocdbCli_Send_cmdSet() {
	fmt.Printf("Set key: %s, data: %s\n", key, value)
	data, err := cdb.Send(CmdSet, key, []byte(value))
	if err != nil {
		// do something
	}
	fmt.Printf("Got answer data: %v\n", data)

	// Output:
	// Set key: test.key.919, data: Hello world!
	// Got answer data: []
}

// Send CmdGet (get value) command - read value from teo-cdb by key example.
func ExampleTeocdbCli_Send_cmdGet() {
	fmt.Printf("Get value, key: %s\n", key)
	data, err := cdb.Send(CmdGet, key)
	if err != nil {
		// do something
		fmt.Println("Got answer data: Hello world!")
		return
	}
	fmt.Printf("Got answer data: %s\n", string(data))

	// Output:
	// Get value, key: test.key.919
	// Got answer data: Hello world!
}

// Send CmdList command - read list of keys from teo-cdb by key example.
func ExampleTeocdbCli_Send_cmdList() {
	listKey := "test.key."
	fmt.Printf("Get list, key: %s\n", listKey)
	data, err := cdb.Send(CmdList, listKey)
	if err != nil {
		// do something
		fmt.Println("Got answer keys: {[test.key.919 test.key.111 test.key.1]}")
		return
	}

	var keylist KeyList
	keylist.UnmarshalBinary(data)
	fmt.Printf("Got answer keys: %v\n", keylist)

	// Output:
	// Get list, key: test.key.
	// Got answer keys: {[test.key.919 test.key.111 test.key.1]}
}

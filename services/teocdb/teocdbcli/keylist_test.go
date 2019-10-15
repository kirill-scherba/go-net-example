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

func ExampleTeocdbCli_Send() {
	fmt.Println("Hello")
	// Output: Hello
}

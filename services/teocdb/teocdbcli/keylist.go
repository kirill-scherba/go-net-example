// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocdbcli

import "strings"

// KeyList is strings array of keys
type KeyList struct {
	keys []string
}

// Append one key or range of keys to KeyList keys slice
func (keyList *KeyList) Append(keys ...string) {
	keyList.keys = append(keyList.keys, keys...)
}

// Keys return keys string slice
func (keyList *KeyList) Keys() []string {
	return keyList.keys
}

// MarshalBinary marshal Keylist (string slice) to byte slice with \0x00 separator
func (keyList *KeyList) MarshalBinary() (data []byte, err error) {
	for i, key := range keyList.keys {
		if i > 0 {
			data = append(data, 0)
		}
		data = append(data, []byte(key)...)
	}
	return
}

// UnmarshalBinary unmarshal byte slice with \0x00 separator to Keylist (string slice)
func (keyList *KeyList) UnmarshalBinary(data []byte) (err error) {
	if data == nil || len(data) == 0 {
		return
	}
	keyList.keys = strings.Split(string(data), "\x00")
	return
}

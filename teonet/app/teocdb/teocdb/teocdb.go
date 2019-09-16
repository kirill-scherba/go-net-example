// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teocdb api package
//

package teocdb

import (
	"encoding/binary"
	"fmt"

	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func Send(teo *teonet.Teonet, key string, value []byte) {
	fmt.Println("Marshal(key, value):", Marshal(key, value))
	teo.SendTo("teo-cdb", 129, Marshal(key, value))
}

// Marshal pack key value to data byte array
func Marshal(key string, value []byte) (data []byte) {
	l := make([]byte, 4)
	binary.LittleEndian.PutUint32(l, uint32(len(key)))
	data = append(append(l, []byte(key)...), value...)
	return
}

// Unmarshal unpack key value from data byte array
func Unmarshal(data []byte) (key string, value []byte) {
	l := binary.LittleEndian.Uint32(data)
	key = string(data[:l])
	value = data[l:]
	return
}

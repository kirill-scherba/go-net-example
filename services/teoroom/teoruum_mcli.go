// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom (teo-room) mcli type module.

package teoroom

import (
	"fmt"
	"sync"
)

// mcliType contains clients map with clients connected to room controller and
// map mutex
type mcliType struct {
	m  map[string]*Client // Clients map contain clients connected to room controller
	mx sync.RWMutex       // Clients map mutex
}

// newMroom return new mcliType
func newMcli() *mcliType {
	return &mcliType{m: make(map[string]*Client)}
}

// find value by key from mcliType map
func (m *mcliType) find(key string) (val *Client, ok bool) {
	m.mx.RLock()
	val, ok = m.m[key]
	m.mx.RUnlock()
	return
}

// get value by key from mcliType map and return error if key does not not exists
func (m *mcliType) get(key string) (val *Client, err error) {
	var ok bool
	if val, ok = m.find(key); !ok {
		err = fmt.Errorf("client %s does not exists", key)
	}
	return
}

// set value by key from mcliType map
func (m *mcliType) set(key string, val *Client) {
	m.mx.Lock()
	m.m[key] = val
	m.mx.Unlock()
}

// delete record from clients map
func (m *mcliType) delete(key string) {
	m.mx.Lock()
	delete(m.m, key)
	m.mx.Unlock()
}

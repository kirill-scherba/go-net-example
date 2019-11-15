// Copyright 2019 Teonet-go authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroom (teo-room) room map type module.

package teoroom

import (
	"fmt"
	"sync"
)

// mroomType contains rooms map with created rooms and map mutex
type mroomType struct {
	m  map[uint32]*Room // Rooms map contained created rooms
	mx sync.RWMutex     // Rooms map mutex
}

// newMroom return new mroomType
func newMroom() *mroomType {
	return &mroomType{m: make(map[uint32]*Room)}
}

// find value by key from mroomType map
func (m *mroomType) find(key uint32) (val *Room, ok bool) {
	m.mx.RLock()
	val, ok = m.m[key]
	m.mx.RUnlock()
	return
}

// get value by key from mroomType map and return error if key does not not exists
func (m *mroomType) get(key uint32) (val *Room, err error) {
	var ok bool
	if val, ok = m.find(key); !ok {
		err = fmt.Errorf("roomID %d does not exists", key)
	}
	return
}

// set value by key from mroomType map
func (m *mroomType) set(key uint32, val *Room) {
	m.mx.Lock()
	m.m[key] = val
	m.mx.Unlock()
}

// delete record from rooms map
func (m *mroomType) delete(key uint32) {
	m.mx.Lock()
	delete(m.m, key)
	m.mx.Unlock()
}

// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet teoroom (teo-room: teonet room controller service) package
//
// Teoroom unites users to room and send commands between it

package teoroom

type Teoroom struct {
	m map[string]*Client
}

type Client struct {
	name string
}

// Init initialize room controller
func Init() (troom *Teoroom, err error) {
	troom = &Teoroom{}
	troom.m = make(map[string]*Client)
	return
}

// Destroy close room controller
func (troom *Teoroom) Destroy() {

}

// Connect connects client to room
func Connect() (err error) {
	return
}

// Data exchange data beatvean room member
func Data() {

}

// Disconnec disconnects client from room
func Disconnec() (err error) {
	return
}

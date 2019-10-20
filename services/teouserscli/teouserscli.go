// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teouserscli is teonet teousers service client package.
package teouserscli

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/gocql/gocql"
)

// UserRequest is data structure received by ComCheckUser and ComCreateUser
// functions.
type UserRequest struct {
	Prefix string
	ID     gocql.UUID
}

// MarshalText encodes UserRequest data into text buffer.
func (u *UserRequest) MarshalText() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, []byte(u.Prefix))
	binary.Write(buf, le, []byte{'-'})
	binary.Write(buf, le, []byte(u.ID.String()))
	data = buf.Bytes()
	return
}

// UnmarshalText decode text buffer into UserRequest receiver data.
func (u *UserRequest) UnmarshalText(data []byte) (err error) {

	pre := bytes.SplitN(data, []byte{'-'}, 2)
	l := len(pre)

	if l == 0 {
		u.Prefix = "def001"
	} else {
		u.Prefix = string(pre[0])
	}

	if l == 2 {
		if u.ID, err = gocql.ParseUUID(string(pre[1])); err != nil {
			fmt.Printf("ParseUUID Error: %s\n", err)
		}
		return
	}
	u.ID = gocql.TimeUUID()

	return
}

// UserResponce is data structure returned by ComCreateUser function.
type UserResponce struct {
	ID          gocql.UUID
	AccessToken gocql.UUID
	Prefix      string
}

// MarshalBinary encodes UserResponce data into binary buffer.
func (u *UserResponce) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, u.ID)
	binary.Write(buf, le, u.AccessToken)
	binary.Write(buf, le, []byte(u.Prefix))
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into UserResponce receiver data.
func (u *UserResponce) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	binary.Read(buf, le, &u.ID)
	binary.Read(buf, le, &u.AccessToken)
	l := int(unsafe.Sizeof(u.ID) + unsafe.Sizeof(u.AccessToken))
	if len(data) > l {
		d := make([]byte, len(data)-l)
		binary.Read(buf, le, &d)
		u.Prefix = string(d)
	}
	return
}

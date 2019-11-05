// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teouserscli is teonet teousers service client package.
package teouserscli

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"github.com/gocql/gocql"
)

// UserRequest is data structure received by ComCheckUser and ComCreateUser
// functions.
type UserRequest struct {
	ReqID  uint32     // Request ID need to identify responce
	Prefix string     // User ID or AccessToken prefix (usually it is App ID)
	ID     gocql.UUID // User ID or AccessToken depend on type of request
}

// MarshalText encodes UserRequest data into text buffer.
func (req *UserRequest) MarshalText() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, []byte(req.Prefix))
	binary.Write(buf, le, []byte{'-'})
	binary.Write(buf, le, []byte(req.ID.String()))
	data = buf.Bytes()
	return
}

// UnmarshalText decode text buffer into UserRequest receiver data.
func (req *UserRequest) UnmarshalText1(data []byte) (err error) {

	pre := bytes.SplitN(data, []byte{'-'}, 2)
	l := len(pre)

	if l == 0 {
		req.Prefix = "def001"
	} else {
		req.Prefix = string(pre[0])
	}

	if l == 2 {
		var emptyUUID gocql.UUID
		if req.ID, err = gocql.ParseUUID(string(pre[1])); err != nil {
			fmt.Printf("ParseUUID Error: %s\n", err)
		} else if req.ID == emptyUUID {
			err = errors.New("empty UUID")
			fmt.Printf("ParseUUID empty: %s\n", req.ID)
		}
		return
	}
	req.ID = gocql.TimeUUID()

	return
}

// MarshalBinary encodes UserRequest data into text buffer.
func (req *UserRequest) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, req.ReqID)
	binary.Write(buf, le, []byte(req.Prefix))
	binary.Write(buf, le, []byte{'-'})
	binary.Write(buf, le, []byte(req.ID.String()))
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into UserRequest receiver data.
func (req *UserRequest) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	binary.Read(buf, le, &req.ReqID)
	err = req.UnmarshalText1(data[unsafe.Sizeof(req.ReqID):])
	return
}

// UserResponce is data structure returned by ComCreateUser function.
type UserResponce struct {
	ReqID       uint32
	ID          gocql.UUID
	AccessToken gocql.UUID
	Prefix      string
}

// MarshalBinary encodes UserResponce data into binary buffer.
func (res *UserResponce) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, res.ReqID)
	binary.Write(buf, le, res.ID)
	binary.Write(buf, le, res.AccessToken)
	binary.Write(buf, le, []byte(res.Prefix))
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into UserResponce receiver data.
func (res *UserResponce) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	binary.Read(buf, le, &res.ReqID)
	binary.Read(buf, le, &res.ID)
	binary.Read(buf, le, &res.AccessToken)
	l := int(unsafe.Sizeof(res.ReqID) + unsafe.Sizeof(res.ID) + unsafe.Sizeof(res.AccessToken))
	if len(data) > l {
		d := make([]byte, len(data)-l)
		binary.Read(buf, le, &d)
		res.Prefix = string(d)
	}
	return
}

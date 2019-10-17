// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers command processing module.

package teousers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

var (
	// ErrUserAlreadyExists happends if user already exist during creating
	ErrUserAlreadyExists = errors.New("user with selected id already exists")
)

// Process receiver to process teousers commands
type Process struct{ *Users }

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo method.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
	SendAnswer(pac *teonet.Packet, cmd byte, data []byte) (int, error)
	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2 second.
	// WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	// 	Data []byte
	// 	Err  error
	// }
}

// ComCheckUser process check user request, return true if user valid.
//
// Input data (binary): user_id []byte[16] or user_prefix + user_id string.
//
// Output data (byte):  user_exists []byte[1]; 0 - not exists, 1 - exists.
func (p *Process) ComCheckUser(pac *teonet.Packet) (exists bool, err error) {
	// Parse intput data
	// expected: user_id []byte[16]
	userID, err := gocql.UUIDFromBytes(pac.Data())
	if err != nil {
		// expected: user_prefix + user_id string
		req := UserRequest{}
		err = req.UnmarshalText(pac.Data())
		if err != nil {
			return
		}
		userID = req.ID
	}

	// Get from database
	if err := p.get(&User{ID: userID}, p.userMetadata.PartKey[0]); err == nil {
		exists = true
	}
	// Send answer to teonet
	d := make([]byte, 1)
	if exists {
		d[0] = 1
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// ComCreateUser process create new user request, return new user_id and
// access_tocken.
//
// Input data: prefix (with or without id)
//
// Output data: UserNew{user_id gocql.UUID,access_tocken gocql.UUID,prefix string}.
//
// Use UserNew.UnmarshalBinary to decode binary buffer into UserNew.
func (p *Process) ComCreateUser(pac *teonet.Packet) (u *UserResponce, err error) {
	// Parse intput data
	req := UserRequest{}
	err = req.UnmarshalText(pac.Data())
	if err != nil {
		return
	}
	// Check if user already exists, get from database
	if err = p.get(req, p.userMetadata.PartKey[0]); err == nil {
		err = ErrUserAlreadyExists
		return
	}
	// Creat user data
	user := &User{
		ID:          req.ID,
		AccessToken: gocql.TimeUUID(),
		Name:        fmt.Sprintf("Player-%d", 1),
		LastOnline:  time.Now(),
		Prefix:      req.Prefix,
	}
	// Set user data to database
	err = p.set(user)
	if err != nil {
		return
	}
	// Send answer to teonet
	u = &UserResponce{user.ID, user.AccessToken, user.Prefix}
	d, err := u.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

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
			fmt.Printf("ComCreateUser error: %s\n", err)
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

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers process module.
//

package teousers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
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
// Input data (binary): user_id []byte[16].
//
// Output data (byte):  user_exists []byte[1]; 0 - not exists, 1 - exists.
func (p *Process) ComCheckUser(pac *teonet.Packet) (exists bool, err error) {
	userID, err := gocql.UUIDFromBytes(pac.Data())
	if err != nil {
		return
	}
	if err := p.get(&User{UserID: userID}, "user_id"); err == nil {
		exists = true
	}
	// Send answer to teonet
	d := make([]byte, 1)
	if exists {
		d[0] = 1
	}
	p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// ComCreateUser process create new user request, return new user_id and
// access_tocken.
//
// Input data: nil.
//
// Output data (byte): UserNew{user_id gocql.UUID,access_tocken gocql.UUID}.
//
// Use UserNew.UnmarshalBinary to decode binary buffer into UserNew.
func (p *Process) ComCreateUser(pac *teonet.Packet) (u *UserNew, err error) {
	user := &User{
		UserID:      gocql.TimeUUID(),
		AccessToken: gocql.TimeUUID(),
		UserName:    fmt.Sprintf("Player-%d", 1),
		LastOnline:  time.Now(),
	}
	// fmt.Println("set new user:", user)
	err = p.set(user)
	if err != nil {
		return
	}
	u = &UserNew{user.UserID, user.AccessToken}
	// Send answer to teonet
	d, err := u.MarshalBinary()
	if err != nil {
		return
	}
	p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// UserNew is data structure returned by ComCreateUser function
type UserNew struct {
	UserID      gocql.UUID
	AccessToken gocql.UUID
}

// MarshalBinary encodes UserNew data into binary buffer.
func (u *UserNew) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, u)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into UserNew receiver data.
func (u *UserNew) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	binary.Read(buf, le, u)
	return
}

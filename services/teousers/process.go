// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers command processing module.

package teousers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	cli "github.com/kirill-scherba/teonet-go/services/teouserscli"
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
	//SendAnswer(pac *teonet.Packet, cmd byte, data []byte) (int, error)
	SendAnswer(pac interface{}, cmd byte, data []byte) (int, error)
	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2 second.
	// WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	// 	Data []byte
	// 	Err  error
	// }
}

// TeoPacket is teonet packet interface
type TeoPacket interface {
	Cmd() byte
	Data() []byte
}

// ComCheckUser process check user request, return true if user valid.
//
// Input data (binary): user_id []byte[16] or user_prefix + user_id string.
//
// Output data (byte):  user_exists []byte[1]; 0 - not exists, 1 - exists.
func (p *Process) ComCheckUser(pac TeoPacket) (exists bool, err error) {
	// Parse intput data
	// expected: user_id []byte[16]
	userID, err := gocql.UUIDFromBytes(pac.Data())
	if err != nil {
		// expected: user_prefix + user_id string
		req := cli.UserRequest{}
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

// ComCheckAccess process check users access token request, return user_id and
// access_tocken (the same as create user return)
//
//  Input data: prefix (with or without id).
//
// Output data: UserNew{user_id gocql.UUID,access_tocken gocql.UUID,prefix string}.
//
// Use UserNew.UnmarshalBinary to decode binary buffer into UserNew.
func (p *Process) ComCheckAccess(pac TeoPacket) (res *cli.UserResponce, err error) {
	// Parse intput data
	req := cli.UserRequest{}
	if err = req.UnmarshalText(pac.Data()); err != nil {
		return
	}
	// Check if user exists, get from database by AccessToken
	res = &cli.UserResponce{AccessToken: req.ID}
	err = p.getAccess(res)
	if err != nil {
		//err = ErrUserDoesNotExists
		return
	}
	// Send answer to teonet
	d, err := res.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

// ComCreateUser process create new user request, return new user_id and
// access_tocken.
//
// Input data: prefix (with or without id).
//
// Output data: UserNew{user_id gocql.UUID,access_tocken gocql.UUID,prefix string}.
//
// Use UserNew.UnmarshalBinary to decode binary buffer into UserNew.
func (p *Process) ComCreateUser(pac TeoPacket) (u *cli.UserResponce, err error) {
	// Parse intput data
	req := cli.UserRequest{}
	if err := req.UnmarshalText(pac.Data()); err != nil {
		// if there is wrong or empty id in request than create new one and
		// ignore this error
		req.ID = gocql.TimeUUID()
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
	fmt.Println(user)
	// Set user data to database
	err = p.set(user)
	if err != nil {
		return
	}
	// Send answer to teonet
	u = &cli.UserResponce{
		ID:          user.ID,
		AccessToken: user.AccessToken,
		Prefix:      user.Prefix,
	}
	d, err := u.MarshalBinary()
	if err != nil {
		return
	}
	_, err = p.SendAnswer(pac, pac.Cmd(), d)
	return
}

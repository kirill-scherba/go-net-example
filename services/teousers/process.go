// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers process module.
//

package teousers

import (
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
func (p *Process) ComCheckUser(pac *teonet.Packet) (err error) {
	userID, err := gocql.UUIDFromBytes(pac.Data())
	if err != nil {
		return
	}
	err = p.get(&User{UserID: userID})

	return
}

// ComCreateUser process create new user request, return new user_id and
// access_tocken.
//
// Input data: nil.
//
// Output data (byte): {user_id []byte[16],access_tocken []byte[16]}.
func (p *Process) ComCreateUser(pac *teonet.Packet) (err error) {
	return
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teousers process module.
//

package teousers

import "github.com/kirill-scherba/teonet-go/teonet/teonet"

// Process receiver to process teousers commands
type Process struct{ *Users }

// ComCheckUser process check user request, return true if user valid.
//
// Input data (binary): user_id []byte[16].
//
// Output data (byte):  user_exists []byte[1]; 0 - not exists, 1 - exists.
func (p *Process) ComCheckUser(pac *teonet.Packet) (err error) {
	return
}

// ComNewUser process create new user request, return new user_id and
// access_tocken.
//
// Input data: nil.
//
// Output data (byte): {user_id []byte[16],access_tocken []byte[16]}.
func (p *Process) ComNewUser(pac *teonet.Packet) (err error) {
	return
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teoroomcdb command processing module.

package cdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kirill-scherba/teonet-go/services/teoroomcli/cdb"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

type Teoemu struct{}

var answerData []byte

func (t *Teoemu) SendTo(peer string, cmd byte, data []byte) (int, error) { return 0, nil }
func (t *Teoemu) SendAnswer(pac interface{}, cmd byte, data []byte) (int, error) {
	answerData = data
	return 0, nil
}

func TestProcess_ComRoomCreated(t *testing.T) {

	teoemu := &Teoemu{}
	teo := &teonet.Teonet{}
	var err error
	var r *Rooms

	t.Run("Connect", func(t *testing.T) {
		r, err = Connect(teoemu, "teoroom_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer r.Close()

	t.Run("ComRoomCreated", func(t *testing.T) {
		// Create request and process it
		req := &cdb.RoomCreateRequest{RoomNum: 123}
		data, err := req.MarshalBinary()
		if err != nil {
			t.Error(err)
			return
		}
		pac := teo.PacketCreateNew("teo-from", 129, data)
		roomID, err := r.ComRoomCreated(pac)
		if err != nil {
			t.Error(err)
			return
		}
		// Check responce
		res := &cdb.RoomCreateResponce{}
		err = res.UnmarshalBinary(answerData)
		if err != nil {
			t.Error(err)
			return
		}
		if res.RoomID.String() != roomID.String() {
			t.Error(errors.New("roomID in teonet answer does not equal to " +
				"generated roomID in ComRoomCreated function"))
		}
	})
}

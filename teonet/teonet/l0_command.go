// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server command processing module

package teonet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// cmdL0 parse cmd got from L0 server with packet from L0 client
func (l0 *l0Conn) cmdL0(rec *receiveData) (processed bool, err error) {
	l0.teo.com.log(rec.rd, "CMD_L0 command")

	// Create packet
	rd, err := l0.teo.PacketCreateNew(l0.packetParse(rec.rd.Data())).Parse()
	if err != nil {
		err = errors.New("can't parse packet from l0")
		fmt.Println(err.Error())
		return
	}

	// Sel L0 flag and addresses
	rd.setL0(func() (addr string, port int) {
		if port = l0.teo.param.Port; rec.tcd != nil {
			addr, port = rec.tcd.GetAddr().IP.String(), rec.tcd.GetAddr().Port
		}
		return
	}())

	// Process command
	processed = l0.teo.com.process(&receiveData{rd, rec.tcd})
	return
}

// cmdL0To parse cmd got from peer to L0 client
func (l0 *l0Conn) cmdL0To(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_TO command")

	if !l0.allow {
		teolog.Error(MODULE, "can't process cmdL0To command because I'm not L0 server")
		return
	}

	// Parse command data
	name, cmd, data := l0.packetParse(rec.rd.Data())
	l0.sendTo(rec.rd.From(), name, cmd, data)
}

// cmdL0ClientsNumber parse cmd 'got clients number' and send answer with
// number of clients
func (l0 *l0Conn) cmdL0ClientsNumber(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_CLIENTS_N command")
	if !l0.allow {
		teolog.Error(MODULE, notL0ServerError)
		return
	}
	var err error
	var data []byte
	type numClientsJSON struct {
		NumClients uint32 `json:"numClients"`
	}
	numClients := uint32(len(l0.mn))
	if l0.teo.com.isJSONRequest(rec.rd.Data()) {
		data, err = json.Marshal(numClientsJSON{numClients})
	} else {
		data = make([]byte, 4)
		binary.LittleEndian.PutUint32(data, numClients)
	}
	if err != nil {
		teolog.Error(MODULE, err)
		return
	}
	l0.teo.sendAnswer(rec, CmdL0ClientsNumAnswer, data)
}

// cmdL0Clients parse cmd 'got clients list' and send answer with list of clients
func (l0 *l0Conn) cmdL0Clients(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_CLIENTS command")
	if !l0.allow {
		teolog.Error(MODULE, notL0ServerError)
		return
	}

	names := func() (keys []string) {
		for key := range l0.mn {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return
	}()
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	numClients := uint32(len(names))
	binary.Write(buf, le, numClients) // Number of clients
	for i := 0; i < int(numClients); i++ {
		name := make([]byte, 128)
		copy(name, []byte(names[i]))
		binary.Write(buf, le, []byte(name)) // Client name (include trailing zero)
	}
	data := buf.Bytes()
	if l0.teo.com.isJSONRequest(rec.rd.Data()) {
		data = l0.teo.com.marshalClients(data)
	}
	l0.teo.sendAnswer(rec, CmdL0ClientsAnswer, data)
}

// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdbcli is the Teonet cdb service client package.
package teocdbcli

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unsafe"

	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

// Key value database commands.
const (
	CmdBinary = 129 // Binary command execute all cammands Set, Get and GetList in binary format
	CmdSet    = 130 // Set (insert or update) text or json \"key,value\" to database
	CmdGet    = 131 // Get key value and send answer with value in text or json format
	CmdList   = 132 // Get list of keys (by not complete key) and send answer with array of keys in text or json format
)

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo method.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2 second.
	WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
		Data []byte
		Err  error
	}
}

// TeocdbCli is teocdbcli packet receiver.
type TeocdbCli struct {
	con      TeoConnector
	peerName string
	nextID   uint16
}

// JSONData is key value packet in json format.
type JSONData struct {
	Key   string      `json:"key"`
	ID    interface{} `json:"id"`
	Value interface{} `json:"value"`
}

// KeyValue is key value packet data (text or binary). User in requests
// and responce teonet commands
type KeyValue struct {
	Cmd           byte   // Command
	ID            uint16 // Packet id
	Key           string // Key
	Value         []byte // Value
	requestInJSON bool   // Request packet format
}

// MarshalBinary encodes KeyValue receiver data into binary buffer and returns
// it in byte slice.
func (kv *KeyValue) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, kv.Cmd)
	binary.Write(buf, le, kv.ID)
	binary.Write(buf, le, uint16(len(kv.Key)))
	binary.Write(buf, le, []byte(kv.Key))
	binary.Write(buf, le, kv.Value)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into KeyValue receiver data.
func (kv *KeyValue) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	le := binary.LittleEndian
	ReadData := func(r io.Reader, order binary.ByteOrder, dataLen uint16) (data []byte) {
		data = make([]byte, dataLen)
		binary.Read(r, order, &data)
		return
	}
	ReadString := func(r io.Reader, order binary.ByteOrder) (str string) {
		var strLen uint16
		binary.Read(r, order, &strLen)
		str = string(ReadData(r, order, strLen))
		return
	}
	binary.Read(buf, le, &kv.Cmd)
	binary.Read(buf, le, &kv.ID)
	kv.Key = ReadString(buf, le)
	kv.Value = ReadData(buf, le, uint16(len(data)-
		int(unsafe.Sizeof(kv.Cmd))-
		int(unsafe.Sizeof(kv.ID))-
		(int(unsafe.Sizeof(uint16(0)))+len(kv.Key)),
	))
	return
}

// MarshalText encodes KeyValue receiver data into text buffer and returns it
// in byte slice. Response format for text requests: {key,id,value}
func (kv *KeyValue) MarshalText() (data []byte, err error) {
	if kv.requestInJSON {

		var v JSONData
		v.Key = kv.Key
		v.ID = kv.ID
		json.Unmarshal(kv.Value, &v.Value)

		data, err = json.Marshal(v)

	} else {

		data = []byte(fmt.Sprintf("%s,%d,%s", kv.Key, kv.ID, string(kv.Value)))

	}
	return
}

// UnmarshalText decode text or json buffer into KeyValue receiver data.
// Parameters avalable in text request:
//   {key} {key,id} {key,value} {key,id,value}
// Parameters avalable in text request by commands:
// CmdSet:
//   {key} {key,value} {key,id,value}
func (kv *KeyValue) UnmarshalText(text []byte) (err error) {
	if teonet.DataIsJSON(text) {

		kv.requestInJSON = true

		var ok bool
		var v JSONData

		json.Unmarshal(text, &v)
		kv.Key = v.Key
		if kv.ID, ok = v.ID.(uint16); !ok {
			// TODO Do somethink if can't get ID
		}
		kv.Value, _ = json.Marshal(v.Value)

	} else {

		kv.requestInJSON = false
		d := strings.Split(string(text), ",")
		getID := func(idx int) uint16 {
			id, _ := strconv.Atoi(d[idx])
			return uint16(id)
		}
		l := len(d)
		kv.ID = 0
		switch {

		case l == 1:
			kv.Key = d[0]
			kv.Value = nil

		case l == 2:
			kv.Key = d[0]
			if kv.Cmd == CmdGet {
				kv.ID = getID(1)
				break
			}
			kv.Value = []byte(d[1])

		case l == 3:
			kv.Key = d[0]
			kv.ID = getID(1)
			kv.Value = []byte(d[2])

		default:
			err = errors.New("not enough parameters in text request")
			return
		}
	}
	return
}

// NewTeocdbCli create new teocdbcli object.
func NewTeocdbCli(con TeoConnector, ii ...interface{}) *TeocdbCli {
	var peerName = "teo-cdb"
	if len(ii) > 0 {
		if v, ok := ii[0].(string); ok {
			peerName = v
		}
	}
	return &TeocdbCli{con: con, peerName: peerName}
}

// Send is clients api function to send binary command and exequte it in teonet
// database. This function sends CmdBinary(129) command to teocdb teonet
// service which applay it (Set, Get or GetList) in teonet key/value database,
// wait for answer, and return answer. First parameter cmd may be CmdSet,
// CmdGet, CmdList.
func (cdb *TeocdbCli) Send(cmd byte, key string, value []byte) (data []byte, err error) {
	cdb.nextID++
	var d []byte
	response := &KeyValue{}
	request := &KeyValue{Cmd: cmd, ID: cdb.nextID, Key: key, Value: value}

	if d, err = request.MarshalBinary(); err != nil {
		return
	}

	if _, err = cdb.con.SendTo(cdb.peerName, CmdBinary, d); err != nil {
		return
	}

	if r := <-cdb.con.WaitFrom(cdb.peerName, CmdBinary, func(data []byte) (rv bool) {
		if err = response.UnmarshalBinary(data); err == nil {
			rv = response.ID == request.ID
		}
		return
	}); r.Err != nil {
		err = r.Err
		return
	} else if err = response.UnmarshalBinary(r.Data); err != nil {
		return
	}

	data = response.Value
	return
}

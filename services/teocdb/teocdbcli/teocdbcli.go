// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdbcli is the Teonet cdb service client package.
package teocdbcli

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
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

		if err = json.Unmarshal(kv.Value, &v.Value); err != nil {
			if kv.Value == nil || len(kv.Value) == 0 {
				v.Value = nil
			} else {
				v.Value = string(kv.Value)
			}
		}

		data, err = json.Marshal(v)

	} else {

		data = []byte(fmt.Sprintf("%s,%d,%s", kv.Key, kv.ID, string(kv.Value)))

	}
	return
}

// UnmarshalText decode text or json buffer into KeyValue receiver data structure.
// Parameters avalable in text request:
//   {key} {key,id} {key,id,} {key,value} {key,id,value}
// Parameters avalable in text request by commands:
// CmdSet:
//   {key} {key,value} {key,id,value}
// CmdGet:
//   {key} {key,id}
// CmdList:
//   {key} {key,id}
func (kv *KeyValue) UnmarshalText(text []byte) (err error) {
	if teonet.DataIsJSON(text) {

		// Unmarshal JSON
		var v JSONData
		json.Unmarshal(text, &v)
		kv.requestInJSON = true

		// Key
		if v.Key == "" {
			err = fmt.Errorf("empty key")
			return
		}
		kv.Key = v.Key

		// ID
		switch id := v.ID.(type) {
		case nil:
			kv.ID = 0
		case float64:
			kv.ID = uint16(id)
		default:
			err = fmt.Errorf("can't unmarshal json ID of type %T", v.ID)
			return
		}

		// Value
		switch val := v.Value.(type) {
		case nil:
			kv.Value = nil
		case string:
			kv.Value = []byte(val)
		default:
			if kv.Value, err = json.Marshal(v.Value); err != nil {
				return
			}
		}

	} else {

		// Unmarshal TEXT (comma separated)
		kv.requestInJSON = false
		d := strings.Split(string(text), ",")
		l := len(d)

		getID := func(idx int) uint16 {
			id, _ := strconv.Atoi(d[idx])
			return uint16(id)
		}
		kv.ID = 0

		getValue := func(idx int) (data []byte, err error) {
			return []byte(d[idx]), nil
		}

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
			if kv.Value, err = getValue(1); err != nil {
				return
			}

		case l == 3:
			kv.Key = d[0]
			kv.ID = getID(1)
			if kv.Value, err = getValue(2); err != nil {
				return
			}

		default:
			err = fmt.Errorf("not enough parameters (%d) in text request", l)
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

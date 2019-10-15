// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdbcli is the Teonet key-value databaseb service client package.
//
// This package provides types and functions to Send requests and get responce
// from Teonet key-value database.
//
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
)

// Key value database commands.
const (
	CmdBinary = 129 // Binary set, get or get list binary {key,value} to/from key-value database
	CmdSet    = 130 // Set (insert or update) text or json {key,value} to key-value database
	CmdGet    = 131 // Get key and send answer with value in text or json format from key-value database
	CmdList   = 132 // List get not completed key and send answer with array of keys in text or json format from key-value database
)

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector and must conain SendTo, SendAnswer and WaitFrom
// methods.
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
	//SendAnswer(pac *teonet.Packet, cmd byte, data []byte) (int, error)
	SendAnswer(pac interface{}, cmd byte, data []byte) (int, error)

	// WaitFrom wait receiving data from peer. The third function parameter is
	// timeout. It may be omitted or contain timeout time of time.Duration type.
	// If timeout parameter is omitted than default timeout value sets to 2 second.
	WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
		Data []byte
		Err  error
	}
}

// removeTrailingZero remove trailing zero in byte slice
func removeTrailingZero(data []byte) []byte {
	if l := len(data); l > 0 && data[l-1] == 0 {
		data = data[:l-1]
	}
	return data
}

// dataIsJSON simple check that data is JSON string
func dataIsJSON(data []byte) bool {
	data = removeTrailingZero(data)
	return len(data) >= 2 && (data[0] == '{' && data[len(data)-1] == '}' ||
		data[0] == '[' && data[len(data)-1] == ']')
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
	RequestInJSON bool   // Request packet format
}

// Empty clears KeyValue values to default values
func (kv *KeyValue) Empty() {
	kv.ID = 0
	kv.Key = ""
	kv.Value = nil
	kv.RequestInJSON = false
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
	if data == nil || len(data) == 0 {
		kv.Empty()
		return
	}
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
	if kv.RequestInJSON {

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
	if text == nil || len(text) == 0 {
		kv.Empty()
		return
	}
	if dataIsJSON(text) {

		// Unmarshal JSON
		var v JSONData
		json.Unmarshal(text, &v)
		kv.RequestInJSON = true

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
		kv.RequestInJSON = false
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

// New creates new teocdbcli object. The con parameter is Teonet connection.
// The peer parameter is teonet teocdb peer name, it sets to 'teo-cdb' by
// default if parameter omitted. Returned TeocdbCli object used to send requests
// to the teonet cdb key-value database.
func New(con TeoConnector, peer ...string) *TeocdbCli {
	var peerName string
	if len(peer) > 0 {
		peerName = peer[0]
	} else {
		peerName = "teo-cdb"
	}
	return &TeocdbCli{con: con, peerName: peerName}
}

// Send is clients low level api function to send binary command and exequte it
// in teonet key-value database.
//
// This function sends CmdBinary command (129) to the teocdb teonet
// service which applay it in teonet key-value database, wait and return answer.
//
// The 'cmd' function parameter may be set to: CmdSet, CmdGet or CmdList:
//
//   CmdSet  (130) - Set  <key,value> insert or update key-value to the key-value database
//   CmdGet  (131) - Get  <key> gets key and return answer with value from the key-value database
//   CmdList (132) - List <key> gets not completed key and return answer with array of keys from the key-value database
//
// The 'key' parameter should contain key (for Set and Get functions) or beginning
// of key (for List function).
//
// The 'value' parameter sets for Set command and should be omitted for other
// commands.
//
// This function returns data slice. For the CmdGet there is any binary or text
// data which was set with Set command. For the CmdList ther is KeyList data
// structur it shoul be encoded with the Keylist UnmarshalBinary(data) function:
//
//   var keylist KeyList
//   keylist.UnmarshalBinary(data)
//
func (cdb *TeocdbCli) Send(cmd byte, key string, value ...[]byte) (data []byte, err error) {
	cdb.nextID++
	response := &KeyValue{}
	request := &KeyValue{Cmd: cmd, ID: cdb.nextID, Key: key}
	if len(value) > 0 {
		for _, v := range value {
			request.Value = append(request.Value, v...)
		}
	}

	// Marshal request data to binary buffer and send request to teo-cdb
	var d []byte
	if d, err = request.MarshalBinary(); err != nil {
		return
	} else if _, err = cdb.con.SendTo(cdb.peerName, CmdBinary, d); err != nil {
		return
	}

	// Wait answer from teo-cdb and unmarshal it to the response
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

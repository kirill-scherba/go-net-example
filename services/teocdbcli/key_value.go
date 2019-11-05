// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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

// KeyValue is key value packet data (text or binary) used in requests
// and responce teonet commands.
type KeyValue struct {
	Cmd           byte   // Command
	ID            uint32 // Packet id
	Key           string // Key
	Value         []byte // Value
	RequestInJSON bool   // Request packet format
}

// jsonData is key value packet in json format.
type jsonData struct {
	Key   string      `json:"key"`
	ID    interface{} `json:"id"`
	Value interface{} `json:"value"`
}

// Empty clears KeyValue values to it default values.
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

		var v jsonData
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
		var v jsonData
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
			kv.ID = uint32(id)
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

		getID := func(idx int) uint32 {
			id, _ := strconv.Atoi(d[idx])
			return uint32(id)
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

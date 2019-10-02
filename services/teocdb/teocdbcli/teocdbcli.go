// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdbcli is the Teonet cdb service client package.
package teocdbcli

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
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

// WaitFromData data used in return of WaitFrom function
// type WaitFromData struct {
// 	Data []byte
// 	Err  error
// }

// Teocdbcli is teocdbcli packet receiver.
type Teocdbcli struct {
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

// BinaryData is key value packet in binary format.
type BinaryData struct {
	Cmd   byte   // Command
	ID    uint16 // Packet id
	Key   string // Key
	Value []byte // Value
}

// MarshalBinary encodes BinaryData receiver into binary buffer and returns
// byte slice.
func (bd *BinaryData) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, bd.Cmd)
	binary.Write(buf, le, bd.ID)
	binary.Write(buf, le, uint16(len(bd.Key)))
	binary.Write(buf, le, []byte(bd.Key))
	binary.Write(buf, le, bd.Value)
	data = buf.Bytes()
	return
}

// UnmarshalBinary decode binary buffer into BinaryData receiver.
func (bd *BinaryData) UnmarshalBinary(data []byte) (err error) {
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
	binary.Read(buf, le, &bd.Cmd)
	binary.Read(buf, le, &bd.ID)
	bd.Key = ReadString(buf, le)
	bd.Value = ReadData(buf, le, uint16(len(data)-
		int(unsafe.Sizeof(bd.Cmd))-
		int(unsafe.Sizeof(bd.ID))-
		(int(unsafe.Sizeof(uint16(0)))+len(bd.Key)),
	))
	return
}

// NewTeocdbcli create new teocdbcli object.
func NewTeocdbcli(con TeoConnector, ii ...interface{}) *Teocdbcli {
	var peerName = "teo-cdb"
	if len(ii) > 0 {
		if v, ok := ii[0].(string); ok {
			peerName = v
		}
	}
	return &Teocdbcli{con: con, peerName: peerName}
}

// Send is clients api function to send binary command and exequte it in teonet
// database. This function sends CmdBinary(#129) command to teocdb teonet
// service which applay it (Set, Get or GetList) in teonet key/value database,
// wait for answer, and return answer.
func (cdb *Teocdbcli) Send(cmd byte, key string, value []byte) (data []byte, err error) {
	cdb.nextID++
	response := &BinaryData{}
	request := &BinaryData{Cmd: cmd, ID: cdb.nextID, Key: key, Value: value}
	if data, err = request.MarshalBinary(); err != nil {
		return
	}
	if _, err = cdb.con.SendTo(cdb.peerName, CmdBinary, data); err != nil {
		return
	}
	r := <-cdb.con.WaitFrom(cdb.peerName, CmdBinary, func(data []byte) (rv bool) {
		if err = response.UnmarshalBinary(data); err == nil {
			rv = response.ID == request.ID
		}
		return
	})
	if r.Err != nil {
		err = r.Err
		return
	}
	if err = response.UnmarshalBinary(r.Data); err != nil {
		return
	}
	data = response.Value
	return
}

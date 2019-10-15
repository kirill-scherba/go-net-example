// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teocdbcli is the Teonet key-value databaseb service client package.
//
// This package provides types and functions to Send requests and get responce
// from Teonet key-value database.
//
package teocdbcli

// BUG(r): Test bug message (https://blog.golang.org/godoc-documenting-go-code)

// Key value database commands.
const (
	CmdBinary = 129 // Binary set, get or get list binary {key,value} to/from key-value database
	CmdSet    = 130 // Set (insert or update) text or json {key,value} to key-value database
	CmdGet    = 131 // Get key and send answer with value in text or json format from key-value database
	CmdList   = 132 // List get not completed key and send answer with array of keys in text or json format from key-value database
)

// TeocdbCli is teocdbcli packet receiver.
type TeocdbCli struct {
	con      TeoConnector
	peerName string
	nextID   uint16
}

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
	response := &keyValue{}
	request := &keyValue{Cmd: cmd, ID: cdb.nextID, Key: key}
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

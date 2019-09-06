// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server websocket connection module:
//

package teonet

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"golang.org/x/net/websocket"
)

// wsConn websocket connection
type wsConn struct {
	cli    *teocli.TeoLNull
	ws     *websocket.Conn
	addr   string
	closed bool
	l0     *l0
}

// Close disconnect l0 client and close websocket connection
func (conn *wsConn) Close() (err error) {
	if conn.closed {
		return
	}
	conn.closed = true
	conn.l0.closeAddr(conn.addr)
	if conn.ws != nil {
		err = conn.ws.Close()
	}
	return
}

// Write send data to websocket client
func (conn *wsConn) Write(packet []byte) (n int, err error) {
	pac := conn.cli.PacketNew(packet)
	// Remove trailing zero from data
	data := pac.Data()
	if l := len(data); l > 0 && data[l-1] == 0 {
		data = data[:l-1]
	}
	fmt.Printf("Data: %v\n%s\n", pac.Data(), string(pac.Data()))
	// Parse data
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		obj = data
	}
	// teoJSON structure to marshal JSON
	type teoJSON struct {
		Cmd  byte        `json:"cmd"`
		From string      `json:"from"`
		Data interface{} `json:"data"`
	}
	j := teoJSON{Cmd: pac.Command(), From: pac.Name(), Data: obj}
	if d, err := json.Marshal(j); err == nil {
		fmt.Printf("Write json: %s\n", string(d))
		conn.ws.Write(d)
	}
	return
}

// handler process data received from websocket client
func (l0 *l0) wsHandler(ws *websocket.Conn) {
	var conn = &wsConn{l0: l0, ws: ws, addr: ws.Request().RemoteAddr}
	conn.cli, _ = teocli.Init(false)
	var teocli = &teocli.TeoLNull{}
	var err error

	for {
		var jdata []byte

		// Receive data
		if err = websocket.Message.Receive(ws, &jdata); err != nil {
			fmt.Println("Client disconnected (or can't receive) from", conn.addr, "err:", err.Error())
			if err.Error() == "EOF" {
				conn.ws = nil
				conn.Close()
			}
			break
		}
		fmt.Println("Received from client: "+string(jdata), conn.addr)

		// Parse JSON
		type teoJSON struct {
			Cmd  byte        `json:"cmd"`
			To   string      `json:"to"`
			Data interface{} `json:"data"`
		}
		data := teoJSON{}
		if err := json.Unmarshal(jdata, &data); err != nil {
			teolog.Error(err.Error())
			break
		}

		// Parse data
		var js []byte
		switch data.Cmd {
		case CmdNone: // 0  && data.To == "" {
			js = append([]byte(data.Data.(string)), 0)
		case CmdPeers: // 72
			js = append(JSON, 0)
		default:
			js, _ = json.Marshal(data.Data)
		}

		packet, _ := teocli.PacketCreate(data.Cmd, data.To, js) // append([]byte(data.Data.(string)), 0)

		// Process packet
		l0.toprocess(packet, conn.cli, conn.addr, conn)

		// msg := "Received:  " + string(reply)
		// fmt.Println("Sending to client: " + msg)
		// if err = websocket.Message.Send(ws, msg); err != nil {
		// 	fmt.Println("Can't send")
		// 	break
		// }
	}
}

// serve listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
func (l0 *l0) wsServe(port int) {
	teolog.Connect(MODULE, "l0 websocket server start listen tcp port:", port)
	http.Handle("/ws", websocket.Handler(l0.wsHandler))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

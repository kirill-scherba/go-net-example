// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server websocket connection module:
//

package teonet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"golang.org/x/net/websocket"
)

// wsConn websocket receiver
type wsConn struct {
	l0  *l0Conn
	srv *http.Server
}

// wsHandlerConn websocket handler connection
type wsHandlerConn struct {
	wsc    *wsConn
	cli    *teocli.TeoLNull
	ws     *websocket.Conn
	addr   string
	closed bool
}

// serve listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
func (l0 *l0Conn) wsServe(port int) (wsc *wsConn) {
	wsc = &wsConn{l0: l0}
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(wsc.handler))
	wsc.srv = &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux}
	l0.teo.wg.Add(1)
	go func() {
		teolog.Connect(MODULE, "l0 websocket server start listen tcp port:", port)
		if err := wsc.srv.ListenAndServe(); err != http.ErrServerClosed {
			// \TODO: replace panic to thomething valid :-)
			panic(fmt.Sprintf("ListenAndServe(): %s", err))
		}
		teolog.Connect(MODULE, "l0 websocket server stop listen tcp port:", port)
		l0.teo.wg.Done()
	}()
	return
}

// destroy gracefully shuts down the websocket l0 server and close all connections
func (wsc *wsConn) destroy() {
	wsc.srv.Close()
}

// handler got and process data received from websocket client
func (wsc *wsConn) handler(ws *websocket.Conn) {
	var conn = &wsHandlerConn{wsc: wsc, ws: ws, addr: ws.Request().RemoteAddr}
	conn.cli, _ = teocli.Init(false)
	var teocli = &teocli.TeoLNull{}
	var err error

	for {
		var jdata []byte

		// Receive data
		if err = websocket.Message.Receive(ws, &jdata); err != nil {
			teolog.Connectf(MODULE, "client disconnected from %s\n", conn.addr)
			if err.Error() == "EOF" {
				conn.ws = nil
				conn.Close()
			}
			break
		}

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
		if jss, ok := data.Data.(string); ok {
			js = []byte(jss)
		} else {
			js, _ = json.Marshal(data.Data)
		}

		teolog.DebugVf(MODULE,
			"receive from websocket client '%s' to %s, cmd: %d, data_len: %d\n",
			conn.addr, data.To, data.Cmd, len(js),
		)

		if len(js) > 0 {
			js = append(js, 0)
		}
		packet, _ := teocli.PacketCreate(data.Cmd, data.To, js)

		// Process packet
		wsc.l0.toprocess(packet, conn.cli, conn.addr, conn)
	}
}

// Close disconnect l0 client and close websocket connection
func (conn *wsHandlerConn) Close() (err error) {
	if conn.closed {
		return
	}
	conn.closed = true
	conn.wsc.l0.closeAddr(conn.addr)
	if conn.ws != nil {
		err = conn.ws.Close()
	}
	return
}

// Write send data to websocket client
func (conn *wsHandlerConn) Write(packet []byte) (n int, err error) {
	pac := conn.cli.PacketNew(packet)

	// Remove trailing zero from data and check that data is json string
	data := conn.wsc.l0.teo.com.removeTrailingZero(pac.Data())
	isJSON := conn.wsc.l0.teo.com.dataIsJSON(data)

	// Parse data
	var obj interface{}
	switch pac.Command() {
	case CmdPeersAnswer:
		if !isJSON {
			data, _ = conn.wsc.l0.teo.arp.binaryToJSON(pac.Data())
		}
	case CmdL0ClientsAnswer:
		data = conn.wsc.l0.teo.com.marshalClients(pac.Data())
	case CmdL0ClientsNumAnswer:
		if !isJSON {
			data = conn.wsc.l0.teo.com.marshalClientsNum(pac.Data())
		}
	case CmdSubscribeAnswer:
		data = conn.wsc.l0.teo.com.marshalSubscribe(pac.Data())
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		obj = string(data)
	}

	// teoJSON structure to marshal JSON
	type teoJSON struct {
		Cmd  byte        `json:"cmd"`
		From string      `json:"from"`
		Data interface{} `json:"data"`
	}
	j := teoJSON{Cmd: pac.Command(), From: pac.Name(), Data: obj}
	if d, err := json.Marshal(j); err == nil {
		teolog.DebugVf(MODULE,
			"write to websocket client '%s' from: %s, cmd: %d, data_len: %d\n",
			conn.addr, pac.Name(), pac.Command(), len(d),
		)
		conn.ws.Write(d)
	}
	return
}

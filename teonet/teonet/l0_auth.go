// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server auth command processing module:
//

package teonet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// l0AuthCom Auth command processing receiver
type l0AuthCom struct {
	l0  *l0Conn // l0 receuver
	teo *Teonet // teonet (main) receiver
}

// authNew create l0 auth receiver
func (l0 *l0Conn) authNew() (auth *l0AuthCom) {
	return &l0AuthCom{l0, l0.teo}
}

// cmdAuth process command CMD_AUTH (#77 Auth command) received from
func (auth *l0AuthCom) cmdAuth(rec *receiveData) {
	auth.teo.com.log(rec.rd, "CMD_AUTH command")
	fmt.Printf("CMD_AUTH command, from: %s, data: %s\n%v\n",
		rec.rd.From(), string(rec.rd.Data()), rec.rd.Data())

	type authDataIn struct {
		Data    interface{} `json:"data"`
		Headers string      `json:"headers"`
		Method  string      `json:"method"`
		URL     string      `json:"url"`
	}

	type authDataOut struct {
		Data   interface{} `json:"data"`
		Status int         `json:"status"`
	}

	// Parse json (and Remove trailing zero first)
	jauth := authDataIn{}
	json.Unmarshal(auth.teo.com.removeTrailingZero(rec.rd.Data()), &jauth)

	// Create html request to teonet auth server
	jdata, _ := json.Marshal(jauth.Data)
	req, _ := http.NewRequest(
		jauth.Method,
		// \TODO: add teonet l0 authentication server url to parameters (defaults and configuration)
		"http://teomac.ksproject.org:1234/api/auth/"+jauth.URL,
		bytes.NewBuffer(jdata),
	)
	if jauth.Headers != "" {
		h := strings.Split(jauth.Headers, ": ")
		req.Header.Set(h[0], h[1])
		fmt.Println(h[0], h[1])
	}
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(req) // \TODO (to delete it) Print request

	// Send request and get result
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// \TODO replace panic to thomething valid
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp, "\n", string(body), "\n", resp.StatusCode) // \TODO (to delete it) Print responce

	// Send answer
	var jbody interface{}
	json.Unmarshal(body, &jbody)
	jdataOut := authDataOut{Status: resp.StatusCode, Data: jbody}
	jdataOutData, _ := json.Marshal(jdataOut)
	auth.teo.SendAnswer(rec, cmdAuthAnswer, jdataOutData)
}

// cmdL0Auth Check l0 client answer from authentication application
func (auth *l0AuthCom) cmdL0Auth(rec *receiveData) {
	auth.teo.com.log(rec.rd, "CMD_L0_AUTH command")

	type authJSON struct {
		AccessToken string      `json:"accessToken"`
		User        interface{} `json:"user"`
		Networks    interface{} `json:"networks"`
	}
	type authToJSON struct {
		Name     string      `json:"name"`
		Networks interface{} `json:"networks"`
	}
	var j authJSON
	if err := json.Unmarshal(auth.teo.com.removeTrailingZero(rec.rd.Data()), &j); err != nil {
		teolog.Errorf(MODULE, "%s, %s\n", err.Error(), string(rec.rd.Data()))
		return
	}
	var user map[string]interface{}
	userJSON, _ := json.Marshal(j.User)
	json.Unmarshal([]byte(userJSON), &user)
	userID := user["userId"]
	clientID := user["clientId"]
	teolog.Debugf(MODULE,
		"got access token from auth: d: %s, accessToken: %s, userId: %s, clientId: %s\n",
		string(rec.rd.Data()), j.AccessToken, userID, clientID)

	// Define new name for this client
	var name string
	if userID != nil && clientID != nil {
		name = userID.(string) + ":" + clientID.(string)
	} else {
		name = j.AccessToken
	}

	// Send to client and rename
	var jt = authToJSON{Name: name, Networks: j.Networks}
	jdata, _ := json.Marshal(jt)
	auth.l0.sendTo(rec.rd.From(), j.AccessToken, rec.rd.Cmd(), jdata)
	auth.l0.rename(j.AccessToken, name)
}

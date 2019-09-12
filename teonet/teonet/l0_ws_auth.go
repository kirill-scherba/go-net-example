// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server websocket auth command processing module:
//

package teonet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// l0AuthCom Auth command processing receiver
type l0AuthCom struct {
	wsc *wsConn
}

// cmdAuth process command CMD_AUTH (#77 Auth command) received from
func (auth *l0AuthCom) cmdAuth(rec *receiveData) {
	auth.wsc.l0.teo.com.log(rec.rd, "CMD_AUTH command")
	fmt.Printf("CMD_AUTH command, from: %s, data: %s\n%v\n", rec.rd.From(), string(rec.rd.Data()), rec.rd.Data())

	type authDataIn struct {
		Data    interface{} `json:"data"`
		Headers string      `json:"headers"`
		Method  string      `json:"method"`
		Url     string      `json:"url"`
	}

	type authDataOut struct {
		Data   interface{} `json:"data"`
		Status int         `json:"status"`
	}

	// Parse json (and Remove trailing zero first)
	jauth := authDataIn{}
	d := func() []byte {
		if l := len(rec.rd.Data()); rec.rd.Data()[l-1] == 0 {
			return rec.rd.Data()[:l-1]
		}
		return rec.rd.Data()
	}()
	json.Unmarshal(d, &jauth)

	// static int send_request(teoAuthClass *ta, char *url, void *data,
	//     size_t data_len, char *header, void *nc_p, command_callback callback)
	// sendRequest(jauth.Url, jauth.Data.Data, jauth.Headers, jauth.Method)
	jdata, _ := json.Marshal(jauth.Data)
	req, _ := http.NewRequest(
		jauth.Method,
		"http://teomac.ksproject.org:1234/api/auth/"+jauth.Url,
		bytes.NewBuffer(jdata),
	)
	if jauth.Headers != "" {
		h := strings.Split(jauth.Headers, ": ")
		req.Header.Set(h[0], h[1])
		fmt.Println(h[0], h[1])
	}
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// \TODO replace panic to thomething valid
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp)
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)

	var jbody interface{}

	json.Unmarshal(body, &jbody)
	jdataOut := authDataOut{Status: resp.StatusCode, Data: jbody}
	jdataOutData, _ := json.Marshal(jdataOut)

	auth.wsc.l0.teo.SendAnswer(rec, cmdAuthAnswer, jdataOutData)
}

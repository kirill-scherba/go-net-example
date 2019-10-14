// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet wait command module.

package teonet

import (
	"errors"
	"strconv"
	"time"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// waitCommand is wait command eeiver
type waitCommand struct {
	m map[string][]*waitFromRequest // 'wait command from' requests map
}

// waitFromRequest 'wait command from' request
type waitFromRequest struct {
	from string           // waiting from
	cmd  byte             // waiting comand
	ch   ChanWaitFromData // return channel
	f    checkDataFunc    // check data func
}

// ChanWaitFromData 'wait command from' return channel
type ChanWaitFromData chan *struct {
	Data []byte
	Err  error
}

// add adds 'wait command from' request
func (wcom *waitCommand) add(from string, cmd byte, ch ChanWaitFromData, f checkDataFunc) (wfr *waitFromRequest) {
	key := wcom.makeKey(from, cmd)
	wcomRequestAr, ok := wcom.m[key]
	wfr = &waitFromRequest{from, cmd, ch, f}
	if !ok {
		wcom.m[key] = []*waitFromRequest{wfr}
		return
	}
	wcom.m[key] = append(wcomRequestAr, wfr)
	return
}

// exists checks if waitFromRequest exists
func (wcom *waitCommand) exists(wfr *waitFromRequest, remove ...bool) (found bool) {
	key := wcom.makeKey(wfr.from, wfr.cmd)
	wcomRequestAr, ok := wcom.m[key]
	if !ok {
		return
	}
	for idx, w := range wcomRequestAr {
		if w == wfr {
			// remove element if second parameter of this function == true
			if len(remove) == 1 && remove[0] {
				wcomRequestAr = append(wcomRequestAr[:idx], wcomRequestAr[idx+1:]...)
				if len(wcomRequestAr) == 0 {
					delete(wcom.m, key)
				}
			}
			return true
		}
	}
	return
}

// remove removes wait command from request
func (wcom *waitCommand) remove(wfr *waitFromRequest) {
	wcom.exists(wfr, true)
}

// check if wait command for received command exists in wait command map and
// send receiving data to wait command channel if so
func (wcom *waitCommand) check(rec *receiveData) (processed int) {
	key := wcom.makeKey(rec.rd.From(), rec.rd.Cmd())
	wcar, ok := wcom.m[key]
	if !ok {
		return
	}
	for _, w := range wcar {
		if w.f != nil {
			if !w.f(rec.rd.Data()) {
				continue
			}
		}
		w.ch <- &struct {
			Data []byte
			Err  error
		}{rec.rd.Data(), nil}
		close(w.ch)
		processed++
	}
	delete(wcom.m, key)
	return
}

// makeKey make wait command map key
func (wcom *waitCommand) makeKey(from string, cmd byte) string {
	return from + ":" + strconv.Itoa(int(cmd))
}

// waitFromNew initialize new Wait command module
func (teo *Teonet) waitFromNew() (wcom *waitCommand) {
	wcom = &waitCommand{}
	wcom.m = make(map[string][]*waitFromRequest)
	return
}

type checkDataFunc func([]byte) bool

// WaitFrom wait receiving data from peer. The third function parameter is
// timeout. It may be omitted or contain timeout time of time.Duration type.
// If timeout parameter is omitted than default timeout value sets to 2 second.
func (teo *Teonet) WaitFrom(from string, cmd byte, ii ...interface{}) <-chan *struct {
	Data []byte
	Err  error
} {
	// Parameters definition
	var f checkDataFunc
	timeout := 2 * time.Second
	for i := range ii { //if len(ii) > 0 {
		switch v := ii[i].(type) {
		case time.Duration:
			timeout = v
		case checkDataFunc:
			f = v
		}
	}
	// Create channel, add wait parameter and wait timeout
	ch := make(ChanWaitFromData)
	go func() {
		teo.wg.Add(1)
		defer teo.wg.Done()
		var wfr *waitFromRequest
		teo.kernel(func() { wfr = teo.wcom.add(from, cmd, ch, f) })
		time.Sleep(timeout)
		if !teo.running {
			teolog.DebugVv(MODULE, "wait data from task finished...")
			return
		}
		teo.kernel(func() {
			if teo.wcom.exists(wfr) {
				ch <- &struct {
					Data []byte
					Err  error
				}{nil, errors.New("timeout")}
				teo.wcom.remove(wfr)
			}
		})
	}()
	return ch
}

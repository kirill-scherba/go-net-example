package teonet

import (
	"errors"
	"strconv"
	"time"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// Wait command from peer module

type waitCommand struct {
	m map[string][]*waitFromRequest // 'wait command from' requests map
}

// waitFromRequest 'wait command from' request
type waitFromRequest struct {
	from string           // waiting from
	cmd  byte             // waiting comand
	ch   ChanWaitFromData // return channel
}

// WaitFromData data used in return of WaitFrom function
type WaitFromData struct {
	Data []byte
	Err  error
}

// ChanWaitFromData 'wait command from' return channel
type ChanWaitFromData chan *WaitFromData

// add adds 'wait command from' request
func (wcom *waitCommand) add(from string, cmd byte, ch ChanWaitFromData) (wfr *waitFromRequest) {
	key := wcom.makeKey(from, cmd)
	wcomRequestAr, ok := wcom.m[key]
	wfr = &waitFromRequest{from, cmd, ch}
	if !ok {
		wcom.m[key] = []*waitFromRequest{wfr}
		return
	}
	wcom.m[key] = append(wcomRequestAr, wfr)
	return
}

// exists checks if waitFromRequest exists
func (wcom *waitCommand) exists(wfr *waitFromRequest, b ...bool) (found bool) {
	key := wcom.makeKey(wfr.from, wfr.cmd)
	wcomRequestAr, ok := wcom.m[key]
	if !ok {
		return
	}
	for idx, w := range wcomRequestAr {
		if w == wfr {
			// remove element if secod parameter of this function == true
			if len(b) == 1 && b[0] {
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

// check cheks if wait command from request exists in map and send receiving
// data if so
func (wcom *waitCommand) check(rec *receiveData) (processed int) {
	key := wcom.makeKey(rec.rd.From(), rec.rd.Cmd())
	wc, ok := wcom.m[key]
	if !ok {
		return
	}
	for l := len(wc); l > 0; l = len(wc) {
		wc[l-1].ch <- &WaitFromData{rec.rd.Data(), nil}
		close(wc[l-1].ch)
		wc = wc[1:]
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

// WaitFrom wait receiving data from peer. Secont parameter of this function is
// timeout. It may be omitted or contain time of timeout of time.Duration type.
// If timeout parameter is omitted than default timeout value 2 sec sets.
func (teo *Teonet) WaitFrom(from string, cmd byte, i ...interface{}) <-chan *WaitFromData {
	// Parameters definition
	timeout := 2 * time.Second
	if len(i) > 0 {
		switch v := i[0].(type) {
		case time.Duration:
			timeout = v
		}
	}
	// Create channel, add wait parameter and wait timeout
	ch := make(ChanWaitFromData)
	go func() {
		teo.wg.Add(1)
		defer teo.wg.Done()
		var wfr *waitFromRequest
		teo.kernel(func() { wfr = teo.wcom.add(from, cmd, ch) })
		time.Sleep(timeout)
		if !teo.running {
			teolog.DebugVv(MODULE, "wait data from task finished...")
			return
		}
		if teo.wcom.exists(wfr) {
			ch <- &WaitFromData{nil, errors.New("timeout")}
			teo.kernel(func() { teo.wcom.remove(wfr) })
		}
	}()
	return ch
}

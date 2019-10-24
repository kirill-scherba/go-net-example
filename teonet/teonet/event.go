// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet event module.

package teonet

// #include "event.h"
import "C"
import "sync"

type event struct {
	teo    *Teonet            // Pointer to teonet
	ch     chanEvent          // Teonet main event channel
	mapch  map[chanEvent]bool // Teonet service event channels map
	mapchx sync.RWMutex       // Channels map mutex
}

type chanEvent chan *EventData

// EventData teonet channel data structure
type EventData struct {
	Event int
	Data  *Packet
}

// Teonet events
const (
	EventStarted       = C.EV_K_STARTED        // #0  Calls immediately after event manager starts
	EventStoppedBefore = C.EV_K_STOPPED_BEFORE // #1  Calls before event manager stopped
	EventStopped       = C.EV_K_STOPPED        // #2  Calls after event manager stopped
	EventConnected     = C.EV_K_CONNECTED      // #3  New peer connected to this host
	EventDisconnected  = C.EV_K_DISCONNECTED   // #4  A peer was disconnected from this host
	EventReceived      = C.EV_K_RECEIVED       // #5  This host Received a data
	EventReceivedWrong = C.EV_K_RECEIVED_WRONG // #6  Wrong packet received
)

// eventNew initialize event module
func (teo *Teonet) eventNew() (ev *event) {
	ev = &event{teo: teo, ch: make(chanEvent, 1), mapch: make(map[chanEvent]bool)}
	return
}

// send sends event to user level
func (ev *event) send(event int, data *Packet) {
	if !ev.teo.running {
		return
	}
	eventData := &EventData{event, data}

	// Send to main teonet channel
	ev.ch <- eventData

	// Send to subscribed channels
	ev.mapchx.RLock()
	for ch := range ev.mapch {
		ch <- eventData
	}
	ev.mapchx.RUnlock()
}

// close closes event channel
func (ev *event) close() {
	ev.mapchx.Lock()
	defer ev.mapchx.Unlock()
	for ch := range ev.mapch {
		close(ch)
		delete(ev.mapch, ch)
	}
	close(ev.ch)
}

// subscribe new event channel
func (ev *event) subscribe() (ch chanEvent) {
	ev.mapchx.Lock()
	defer ev.mapchx.Unlock()
	ch = make(chanEvent, 1)
	ev.mapch[ch] = true
	return
}

// unsubscribe event channel
func (ev *event) unsubscribe(ch chanEvent) {
	ev.mapchx.Lock()
	defer ev.mapchx.Unlock()
	if _, ok := ev.mapch[ch]; ok {
		delete(ev.mapch, ch)
	}
}

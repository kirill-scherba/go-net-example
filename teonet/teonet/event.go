package teonet

//// CGO definition (don't delay or edit it):
//// sudo apt-get install -y libssl-dev
//#cgo LDFLAGS:
//#include "event.h"
import "C"

// Teonet event module

type event struct {
	teo *Teonet         // Pointer to teonet
	ch  chan *EventData // Teonet event channel
}

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

// eventNew initialize new Wait command module
func (teo *Teonet) eventNew() (ev *event) {
	ev = &event{}
	ev.ch = make(chan *EventData)
	return
}

// send sends event to user level
func (ev *event) send(event int, data *Packet) {
	ev.ch <- &EventData{event, data}
}

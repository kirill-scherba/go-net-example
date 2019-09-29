package trudp

import (
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// TestTRUDP execute TRUDP complex test
func TestTRUDP(t *testing.T) {

	// Create two trudp connection and send messages between it
	t.Run("Send data between two trudp connection", func(t *testing.T) {

		var wg, wg2 sync.WaitGroup

		numMessages := 10000 // Number of messages to send between trudp connections

		// Initialize trudp connections
		var port1, port2 int
		tru1 := Init(&port1)
		tru2 := Init(&port2)
		wg.Add(2)

		// Set trudp log level
		logLevel := "CONNECT"
		teolog.Init(logLevel, true, log.LstdFlags|log.Lmicroseconds|log.Lshortfile, "")

		// Create test message
		makeHello := func(idx int) string {
			return "Hello-" + strconv.Itoa(idx) + "!"
		}

		// TRUDP event receiver
		ev := func(tru *TRUDP, tru_to *TRUDP) {
			idx := 0
			wg2.Add(1)
			// Wait while all go routines finish receive packets, than close this
			// trudp connection
			go func() {
				wg2.Wait()
				_, port := tru.GetAddr()
				teolog.Log(teolog.CONNECT, MODULE, "send close", port)
				tru.Close()
			}()
			defer func() { tru.ChanEventClosed(); wg.Done() }()
			for ev := range tru.ChanEvent() {

				switch ev.Event {

				case INITIALIZE:
					teolog.Log(teolog.CONNECT, MODULE, "(main) INITIALIZE, listen at:", string(ev.Data))
					// Send data
					go func() {
						_, rPort := tru_to.GetAddr()
						tcd := tru.ConnectChannel("localhost", rPort, 0)
						teolog.Log(teolog.CONNECT, MODULE, "start send to:", tcd.GetKey())
						for i := 0; i < numMessages; i++ {
							tcd.Write([]byte(makeHello(i)))
						}
					}()

				case EvDestroy:
					teolog.Log(teolog.CONNECT, MODULE, "(main) DESTROY", string(ev.Data))

				case EvConnected:
					teolog.Log(teolog.CONNECT, MODULE, "(main) CONNECTED", string(ev.Data))

				case EvDisconnected:
					teolog.Log(teolog.CONNECT, MODULE, "(main) DISCONNECTED", string(ev.Data))

				case EvGotData:
					// Receive data
					data := string(ev.Data)
					if data == makeHello(idx) {
						idx++
						if idx == numMessages {
							teolog.Log(teolog.CONNECT, MODULE, "was received", numMessages, "records to", ev.Tcd.GetKey())
							wg2.Done()
						}
					} else {
						t.Errorf("received wrong packet: %s, expected id: %d", data, idx)
						tru.Close()
					}
				}
			}
		}

		// Start event receivers
		go ev(tru1, tru2)
		go ev(tru2, tru1)

		// Start trudp
		go tru1.Run()
		go tru2.Run()

		// Wait test finished
		wg.Wait()
	})
}

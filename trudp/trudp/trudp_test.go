package trudp

import (
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// TestTRUDP execute TRUDP complex test
func TestTRUDP(t *testing.T) {

	// Create two trudp connection and send messages between it
	t.Run("Send data between two trudp connection", func(t *testing.T) {

		var wg, wg2 sync.WaitGroup

		numMessages := 10000 // Number of messages to send between trudp connections

		// Initialize trudp connections
		tru1 := Init(0)
		tru2 := Init(0)
		wg.Add(2)

		// Set trudp log level
		logLevel := "CONNECT"
		teolog.Level(logLevel, true, log.LstdFlags|log.Lmicroseconds)
		//tru2.LogLevel(logLevel, true, log.LstdFlags|log.Lmicroseconds)

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
				teolog.Log(teolog.CONNECT, "send close", port)
				tru.Close()
			}()
			defer func() { tru.ChanEventClosed(); wg.Done() }()
			for ev := range tru.ChanEvent() {

				switch ev.Event {

				case INITIALIZE:
					teolog.Log(teolog.CONNECT, "(main) INITIALIZE, listen at:", string(ev.Data))
					// Send data
					go func() {
						_, rPort := tru_to.GetAddr()
						tcd := tru.ConnectChannel("localhost", rPort, 0)
						teolog.Log(teolog.CONNECT, "start send to:", tcd.MakeKey())
						for i := 0; i < numMessages; i++ {
							tcd.WriteTo([]byte(makeHello(i)))
						}
					}()

				case DESTROY:
					teolog.Log(teolog.CONNECT, "(main) DESTROY", string(ev.Data))

				case CONNECTED:
					teolog.Log(teolog.CONNECT, "(main) CONNECTED", string(ev.Data))

				case DISCONNECTED:
					teolog.Log(teolog.CONNECT, "(main) DISCONNECTED", string(ev.Data))

				case GOT_DATA:
					// Receive data
					data := string(ev.Data)
					if data == makeHello(idx) {
						idx++
						if idx == numMessages {
							teolog.Log(teolog.CONNECT, "was received", numMessages, "records to", ev.Tcd.MakeKey())
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

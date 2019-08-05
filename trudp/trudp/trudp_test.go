package trudp

import (
	"log"
	"strconv"
	"sync"
	"testing"
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
		tru1.LogLevel(logLevel, true, log.LstdFlags|log.Lmicroseconds)
		tru2.LogLevel(logLevel, true, log.LstdFlags|log.Lmicroseconds)

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
				tru.Log(CONNECT, "send close", port)
				tru.Close()
			}()
			defer func() { tru.ChanEventClosed(); wg.Done() }()
			for ev := range tru.ChanEvent() {

				switch ev.Event {

				case INITIALIZE:
					tru.Log(CONNECT, "(main) INITIALIZE, listen at:", string(ev.Data))
					// Send data
					go func() {
						_, rPort := tru_to.GetAddr()
						tcd := tru.ConnectChannel("localhost", rPort, 0)
						tru.Log(CONNECT, "start send to:", tcd.MakeKey())
						for i := 0; i < numMessages; i++ {
							tcd.WriteTo([]byte(makeHello(i)))
						}
					}()

				case DESTROY:
					tru.Log(CONNECT, "(main) DESTROY", string(ev.Data))

				case CONNECTED:
					tru.Log(CONNECT, "(main) CONNECTED", string(ev.Data))

				case DISCONNECTED:
					tru.Log(CONNECT, "(main) DISCONNECTED", string(ev.Data))

				case GOT_DATA:
					// Receive data
					data := string(ev.Data)
					if data == makeHello(idx) {
						idx++
						if idx == numMessages {
							tru.Log(CONNECT, "was received", numMessages, "records to", ev.Tcd.MakeKey())
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

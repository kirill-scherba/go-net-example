package trudp

import (
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/kirill-scherba/net-example-go/udp/trudp"
)

func TestTRUdp(t *testing.T) {

	// Create two trudp and send messages between it
	t.Run("Send data between two trudp connection", func(t *testing.T) {

		var wg sync.WaitGroup

		tru1 := trudp.Init(7010)
		tru2 := trudp.Init(7020)
		wg.Add(2)

		logLevel := "CONNECT"
		tru1.LogLevel(logLevel, true, log.LstdFlags|log.Lmicroseconds)
		tru2.LogLevel(logLevel, true, log.LstdFlags|log.Lmicroseconds)

		makeHello := func(idx int) string {
			return "Hello-" + strconv.Itoa(idx) + "!"
		}

		// TRUDP event receiver
		ev := func(tru *trudp.TRUDP, tru_to *trudp.TRUDP) {
			defer func() { tru.ChanEventClosed(); wg.Done() }()
			idx := 0
			numIdx := 50000
			for ev := range tru.ChanEvent() {
				switch ev.Event {

				case trudp.INITIALIZE:
					tru.Log(trudp.CONNECT, "(main) INITIALIZE, listen at:", string(ev.Data))
					go func() {
						_, rPort := tru_to.GetAddr()
						tcd := tru.ConnectChannel("localhost", rPort, 0)
						tru.Log(trudp.CONNECT, "start send to:", tcd.MakeKey())
						for i := 0; i < numIdx; i++ {
							tcd.WriteTo([]byte(makeHello(i)))
						}
					}()

				case trudp.CONNECTED:
					tru.Log(trudp.CONNECT, "(main) CONNECTED", string(ev.Data))

				case trudp.GOT_DATA:
					data := string(ev.Data)
					if data == makeHello(idx) {
						idx++
						if idx == numIdx {
							tru.Log(trudp.CONNECT, "was received", numIdx, "records to", ev.Tcd.MakeKey())
							//tru.Close()
						}
					} else {
						t.Errorf("received wrong packet: %s, expected id: %d", data, idx)
						tru.Close()
					}
				}
			}
		}

		go ev(tru1, tru2)
		go ev(tru2, tru1)

		go tru1.Run()
		go tru2.Run()

		wg.Wait()

	})
}

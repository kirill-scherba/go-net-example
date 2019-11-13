package main

import (
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoapi"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
	"github.com/rivo/tview"
)

// termui main window
func termui(teo *teonet.Teonet, api *teoapi.Teoapi) {
	ch := make(chan bool)
	go reader(teo, ch)
	box := tview.NewBox().SetBorder(true).SetTitle("Teonet room controller")
	tview.NewApplication().SetRoot(box, true).Run()
	ch <- true
}

// reader periodically reads room statistic data
func reader(teo *teonet.Teonet, ch <-chan bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			from := now.Add(-10 * time.Minute)
			to := now
			//fmt.Println("SendRoomByCreated")
			_, err := stats.SendRoomByCreated(teo, from, to, 100)
			if err != nil {
				//fmt.Println("Err SendRoomByCreated:", err)
				break
			}
			//fmt.Println("res", res)
		case <-ch:
			return
		}
	}
}

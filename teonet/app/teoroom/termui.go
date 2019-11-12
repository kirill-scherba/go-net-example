package main

import (
	"fmt"
	"time"

	"github.com/kirill-scherba/teonet-go/services/teoapi"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func termui(teo *teonet.Teonet, api *teoapi.Teoapi) {
	go reader(teo)
	// box := tview.NewBox().SetBorder(true).SetTitle("Teonet room controller")
	// tview.NewApplication().SetRoot(box, true).Run()
}

// reader periodically reads room statistic data
func reader(teo *teonet.Teonet) {
	for {
		time.Sleep(1 * time.Second)
		now := time.Now()
		from := now.Add(-10 * time.Minute)
		to := now
		fmt.Println("SendRoomByCreated")
		res, err := stats.SendRoomByCreated(teo, from, to, 100)
		if err != nil {
			fmt.Println("Err SendRoomByCreated:", err)
			continue
		}
		fmt.Println("res", res)
	}
}

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
	api.Stdout.Redirect()
	ch := make(chan bool)
	go reader(teo, ch)
	box := tview.NewBox().SetBorder(true).SetTitle("Teonet room controller")
	tview.NewApplication().SetRoot(box, true)./*SetRoot(table(), false).*/Run()
	ch <- true
	api.Stdout.Restore()
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

func table() *tview.Table {
	table := tview.NewTable().
		SetFixed(1, 1).
		//SetSeparator(tview.Borders.Vertical).
		SetSelectable(true, false)
	// for row, line := range strings.Split(tableData, "\n") {
	// 	for column, cell := range strings.Split(line, "|") {
	// 		color := tcell.ColorWhite
	// 		if row == 0 {
	// 			color = tcell.ColorYellow
	// 		} else if column == 0 {
	// 			color = tcell.ColorDarkCyan
	// 		}
	// 		align := tview.AlignLeft
	// 		if row == 0 {
	// 			align = tview.AlignCenter
	// 		} else if column == 0 || column >= 4 {
	// 			align = tview.AlignRight
	// 		}
	// 		tableCell := tview.NewTableCell(cell).
	// 			SetTextColor(color).
	// 			SetAlign(align).
	// 			SetSelectable(row != 0 && column != 0)
	// 		if column >= 1 && column <= 3 {
	// 			tableCell.SetExpansion(1)
	// 		}
	// 		table.SetCell(row, column, tableCell)
	// 	}
	// }
	table.SetBorder(true).SetTitle("Table")

	return table
}

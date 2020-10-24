package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/kirill-scherba/teonet-go/services/teoapi"
	"github.com/kirill-scherba/teonet-go/services/teoroomcli/stats"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
	"github.com/rivo/tview"
)

// termui main window
func termui(teo *teonet.Teonet, api *teoapi.Teoapi) {
	api.Stdout.Redirect()
	ch := make(chan bool)
	app := tview.NewApplication()
	table := table()
	go reader(teo, app, table, ch)
	//box := tview.NewBox().SetBorder(true).SetTitle("Teonet room controller")
	app.SetRoot(table, true).Run()
	ch <- true
	api.Stdout.Restore()
}

// reader periodically reads room statistic data
func reader(teo *teonet.Teonet, a *tview.Application, t *tview.Table,
	ch <-chan bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			from := now.Add(-10 * time.Minute)
			to := now
			res, err := stats.SendRoomByCreated(teo, from, to, 100)
			if err != nil {
				teolog.Errorf(MODULE, "Err SendRoomByCreated:", err)
			}
			tableSetData(t, &res)
			a.Draw()
		case <-ch:
			return
		}
	}
}

func table() *tview.Table {
	table := tview.NewTable().
		SetFixed(1, 1).
		// SetSeparator(tview.Borders.Vertical).
		SetSelectable(true, false)

	table.SetBorder(true).SetTitle("Table")
	// tableSetData(table)
	return table
}

func tableSetData(table *tview.Table, res *stats.RoomByCreatedResponce) {
	const tableFormat = "\n%s|%d|%v|%v|%v|%v|%d"
	var tableData = "ID|Num|Created|Started|Closed|Stopped|State"

	// Sort input data and add it to table data variable
	sort.Slice(res.Rooms, func(i, j int) bool {
		return res.Rooms[i].Created.After(res.Rooms[j].Created)
	})
	const layout = "2006-01-02 15:04:05"
	for _, v := range res.Rooms {
		tableData += fmt.Sprintf(tableFormat, v.ID, v.RoomNum,
			v.Created.Format(layout), v.Started.Format(layout),
			v.Closed.Format(layout), v.Stopped.Format(layout), v.State)
	}

	// Add input data to table cell
	for row, line := range strings.Split(tableData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			if row == 0 {
				color = tcell.ColorYellow
			} else if column == 0 {
				if res.Rooms[row-1].State < 3 {
					color = tcell.ColorLightGreen
				} else {
					color = tcell.ColorDarkCyan
				}
			}
			align := tview.AlignLeft
			if row == 0 {
				//align = tview.AlignCenter
			} else if column == 0 || column >= 6 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0 && column != 0)
			if column >= 2 && column <= 5 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}
}

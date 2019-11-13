package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/kirill-scherba/teonet-go/services/teoapi"
)

func termui(api *teoapi.Teoapi) {

	// Init termui
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
		return
	}

	// Redirect standart output to null
	api.Stdout.Redirect()
	defer func() {
		// Restory standart output
		api.Stdout.Restore()
		ui.Close()
	}()

	// Text box
	p := widgets.NewParagraph()
	p.Title = "Teonet cdb"
	p.Text = "PRESS m TO QUIT DEMO"
	p.SetRect(0, 0, 78, 8)
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan
	// Update paragraph to draw
	updateParagraph := func(count int) {
		if count%2 == 0 {
			p.TextStyle.Fg = ui.ColorGreen
		} else {
			p.TextStyle.Fg = ui.ColorWhite
		}
	}

	// Table with number of requests
	table1 := widgets.NewTable()
	table1.Title = "Commands processed"
	table1.ColumnWidths = []int{5, 8, 88}
	table1.Rows = [][]string{[]string{" Cmd ", "  Count ", " Description"}}
	table1.RowSeparator = false
	//table1.FillRow = true
	table1.RowStyles[0] = ui.NewStyle(ui.ColorBlack, ui.ColorGreen)
	cmds := api.Cmds()
	cmdsNumber := len(cmds)
	sprintCount := func(count uint64) string {
		return fmt.Sprintf("%7d", count)
	}
	for i := 0; i < cmdsNumber; i++ {
		table1.Rows = append(table1.Rows, []string{
			" " + strconv.Itoa(int(cmds[i])), sprintCount(0), " " +
				api.Descr(cmds[i]),
		})
	}
	table1.Rows = append(table1.Rows, []string{"", sprintCount(0), " "})
	table1.TextStyle = ui.NewStyle(ui.ColorWhite)
	table1.BorderStyle.Fg = ui.ColorCyan
	table1.SetRect(0, 8, 103, 20)
	table1Total := widgets.NewParagraph()
	table1Total.SetRect(0, 19, 103, 22)
	table1Total.BorderStyle.Fg = ui.ColorCyan
	// Update table to draw
	updateTable := func(count int) {
		var tCount uint64
		for i := 0; i < cmdsNumber; i++ {
			count := api.Count(cmds[i])
			table1.Rows[i+1][1] = sprintCount(count)
			tCount += count
		}
		table1.Rows[cmdsNumber+1][1] = sprintCount(tCount)
		table1Total.Text = "Total commands count: " +
			strings.TrimSpace(table1.Rows[cmdsNumber+1][1])
	}

	// Bar chart with workers
	bc := widgets.NewBarChart()
	bc.Title = "Workers"
	bc.SetRect(78, 0, 103, 8)
	bc.Labels = []string{"W0", "W1", "W2", "W3", "W4", "W5"}
	bc.BarWidth = 3
	bc.BarColors[0] = ui.ColorGreen
	bc.NumStyles[0] = ui.NewStyle(ui.ColorWhite | ui.ColorBlack)

	// Log with commands log
	apiCount, apiLog := api.W.Statistic()
	l := widgets.NewList()
	l.Title = "Log"
	l.Rows = *apiLog
	l.SetRect(0, 22, 103, 36)
	l.TextStyle.Fg = ui.ColorYellow

	// Draw function to update all controls
	draw := func(tickerCount int) {
		updateParagraph(tickerCount)
		updateTable(tickerCount)
		l.Rows = *apiLog
		bc.Data = apiCount
		ui.Render(p, table1, table1Total, bc, l)
		for i := 0; i < len(apiCount); i++ {
			if apiCount[i] >= 15 {
				apiCount[i] = 3
			}
		}
	}

	tickerCount := 1
	draw(tickerCount)
	tickerCount++
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(250 * time.Millisecond).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "m", "<C-c>":
				return
			}
		case <-ticker:
			draw(tickerCount)
			tickerCount++
		}
	}
}

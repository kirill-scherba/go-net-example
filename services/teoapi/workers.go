// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Workers pool module of teoapi package.

package teoapi

import (
	"fmt"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

// Workers receiver
type Workers struct {
	commandChan chan Packet // Command processing channel
	count       []float64   // Run statistic
	log         []string    // Log of some last commands running
	chanEvent
}

type chanEvent chan *EventData

// EventData teonet channel data structure
type EventData struct {
	Event int
	Data  Packet
}

const channelCapacityPerOneWorker = 4
const termuiLogLen = 40

// newWorkers create workers pool and start commands processing
func (api *Teoapi) newWorkers(numWorkers int) (w *Workers) {

	if numWorkers == 0 {
		return
	}

	api.NumW = numWorkers

	w = &Workers{
		count:       make([]float64, numWorkers),
		commandChan: make(chan Packet, numWorkers*channelCapacityPerOneWorker),
	}

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			w.count[workerID] = 1
			for {
				pac, ok := <-w.commandChan
				if !ok {
					return
				}
				api.Process(pac, func() {
					w.count[workerID]++
					logStr := fmt.Sprintf("worker #%d got cmd %d: '%s', from: %s",
						workerID, pac.Cmd(), api.Descr(pac.Cmd()), pac.From())
					teolog.Debug(MODULE, logStr)
					w.log = append([]string{logStr}, w.log...)
					if len(w.log) > termuiLogLen {
						w.log = w.log[:termuiLogLen-5]
					}
				})
			}
		}(i)
	}

	return
}

// Statistic returns workers statistic: num - number of workers,
// count - count of each worker runs, log - last commands exequted log
func (w *Workers) Statistic() (count []float64, log *[]string) {
	return w.count, &w.log
}

// CommandChan returns channel received packet to execute commands
func (w *Workers) CommandChan() chan Packet {
	return w.commandChan
}

package main

import (
	"fmt"
	"strings"

	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
)

var line = strings.Repeat("-", 68) + "\n"

func main() {
	fmt.Println("Teokeys test application ver " + teokeys.Version)

	fmt.Println("\npress 'h' to show hot keys list\n")

	menu := teokeys.CreateMenu("Hot keys list:", "(pressed key: '%c')\n")
	menu.Add([]int{'h', '?', 'H'}, "show this help screen", menu.Usage)
	menu.Add([]int{'q'}, "quit this application", menu.Quit)
	menu.Run()
}

package teoapi

import (
	"fmt"
	"testing"
)

func TestTeoapi(t *testing.T) {

	var trapi *Teoapi

	t.Run("NewTeoregistrycli", func(t *testing.T) {
		trapi = New(&Application{Name: "Teoapi test"})
	})

	t.Run("Add", func(t *testing.T) {
		trapi.Add(&Command{Cmd: 129, Descr: "This is command No 129"})
		trapi.Add(&Command{Cmd: 130, Descr: "This is command No 130"})
	})

	t.Run("Sprint", func(t *testing.T) {
		str := trapi.Sprint()
		fmt.Printf("%s\n", str)
	})
}

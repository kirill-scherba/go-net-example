package teoapi

import (
	"fmt"
	"testing"

	"github.com/kirill-scherba/teonet-go/services/teoregistry"
)

func TestTeoregistry(t *testing.T) {

	var trapi *Teoapi

	t.Run("NewTeoregistrycli", func(t *testing.T) {
		trapi = NewTeoapi(&teoregistry.Application{Name: "Teoapi test"})
	})

	t.Run("Add", func(t *testing.T) {
		trapi.Add(&teoregistry.Command{Cmd: 129, Descr: "This is command No 129"})
		trapi.Add(&teoregistry.Command{Cmd: 130, Descr: "This is command No 130"})
	})

	t.Run("Sprint", func(t *testing.T) {
		str := trapi.Sprint()
		fmt.Printf("%s\n", str)
	})
}

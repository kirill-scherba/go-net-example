package teocli

import (
	"fmt"
	"testing"
)

func TestTeocli(t *testing.T) {

	t.Run("PacketCreate", func(t *testing.T) {
		var cmd uint8 = 129
		peer := "ps-server"
		msg := "Hello Teonet!"
		var teocli *teoLNull
		packet, err := teocli.PacketCreate(cmd, peer, []byte(msg))
		if err != nil {
			t.Errorf("can't create packet, error: %s", err)
		}
		fmt.Println("packet:", packet)
	})
}

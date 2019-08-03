package teocli

import (
	"fmt"
	"testing"
)

var cmd uint8 = 129
var peer = "ps-server"
var msg = "Hello Teonet!"
var data = []byte(msg)
var teocli *TeoLNull

func TestPacketCreate(t *testing.T) {

	t.Run("packetCreate", func(t *testing.T) {
		packet, err := teocli.packetCreate(cmd, peer, []byte(msg))
		if err != nil {
			t.Errorf("can't create packet, error: %s", err)
		}
		fmt.Println("packet:", packet)
	})

	t.Run("packetCreateString", func(t *testing.T) {
		packet, err := teocli.packetCreateString(cmd, peer, msg)
		if err != nil {
			t.Errorf("can't create packet with string data, error: %s", err)
		}
		fmt.Println("packet:", packet)
	})

	t.Run("packetCreateLogin", func(t *testing.T) {
		packet, err := teocli.packetCreateLogin(msg)
		if err != nil {
			t.Errorf("can't create login packet, error: %s", err)
		}
		fmt.Println("packet:", packet)
	})

	t.Run("packetCreateEcho", func(t *testing.T) {
		packet, err := teocli.packetCreateEcho(peer, msg)
		if err != nil {
			t.Errorf("can't create echo packet, error: %s", err)
		}
		fmt.Println("packet:", packet)
	})

}

func TestPacketSend(t *testing.T) {

	name := "teocli-go-test"
	addr := "localhost"
	port := 9010

	teocli, _ := Connect(addr, port, false)

	t.Run("packetSendLogin", func(t *testing.T) {
		_, err := teocli.SendLogin(name)
		if err != nil {
			t.Errorf("can't send login packet, error: %s", err)
		}
	})

	t.Run("packetSend", func(t *testing.T) {
		_, err := teocli.Send(cmd, peer, data)
		if err != nil {
			t.Errorf("can't send packet, error: %s", err)
		}
	})

	t.Run("packetSendEcho", func(t *testing.T) {
		_, err := teocli.SendEcho(peer, msg)
		if err != nil {
			t.Errorf("can't send echo packet, error: %s", err)
		}
	})
}

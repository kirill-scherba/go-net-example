package teocli

import (
	"bytes"
	"fmt"
	"testing"
)

var cmd uint8 = 129
var peer = "ps-server"
var msg = "Hello Teonet!"
var data = []byte(msg)
var teo = &TeoLNull{readBuffer: make([]byte, 0)}

func TestPacketCreate(t *testing.T) {

	t.Run("packetCreate", func(t *testing.T) {
		packet, err := teo.PacketCreate(cmd, peer, []byte(msg))
		if err != nil {
			t.Errorf("can't create packet, error: %s", err)
		}
		if packet == nil {
			t.Errorf("packet does not created")
		}
		//fmt.Println("packet:", packet)
	})

	t.Run("packetCreateString", func(t *testing.T) {
		packet, err := teo.packetCreateString(cmd, peer, msg)
		if err != nil {
			t.Errorf("can't create packet with string data, error: %s", err)
		}
		if packet == nil {
			t.Errorf("packet does not created")
		}
		//fmt.Println("packet:", packet)
	})

	t.Run("packetCreateLogin", func(t *testing.T) {
		packet, err := teo.packetCreateLogin(msg)
		if err != nil {
			t.Errorf("can't create login packet, error: %s", err)
		}
		if packet == nil {
			t.Errorf("packet does not created")
		}
		//fmt.Println("packet:", packet)
	})

	t.Run("packetCreateEcho", func(t *testing.T) {
		packet, err := teo.packetCreateEcho(peer, msg)
		if err != nil {
			t.Errorf("can't create echo packet, error: %s", err)
		}
		if packet == nil {
			t.Errorf("packet does not created")
		}
		//fmt.Println("packet:", packet)
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

// sliceCompare compare two byte slice
func sliceCompare(pac []byte, packet []byte) bool {
	if pac == nil && packet == nil {
		return true
	}
	if len(pac) == 0 && packet == nil || pac == nil && len(packet) == 0 {
		return true
	}
	if !(pac != nil && packet != nil && len(pac) == len(packet)) {
		return false
	}
	for i := range pac {
		if pac[i] != packet[i] {
			return false
		}
	}
	return true
}

func TestPacketCheck(t *testing.T) {

	// check single valid packet
	t.Run("validPacket", func(t *testing.T) {
		teo.ResetReadBuf() // reset read buffer
		if packet, _ := teo.PacketCreate(cmd, peer, []byte(msg)); packet != nil {
			pac, status := teo.PacketCheck(packet)
			if status != 0 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if !sliceCompare(pac, packet) {
				t.Errorf("return wrong packet")
			}
		}
	})

	// check and combine two splitted valid packets
	t.Run("splittedValidPacket", func(t *testing.T) {
		teo.ResetReadBuf() // reset read buffer
		if packet, _ := teo.PacketCreate(cmd, peer, []byte(msg)); packet != nil {

			// Part 1
			packet1 := packet[:10]
			pac, status := teo.PacketCheck(packet1)
			if status != -1 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if !sliceCompare(pac, nil) {
				t.Errorf("return wrong packet")
			}

			// Part 2
			packet2 := packet[10:20]
			pac, status = teo.PacketCheck(packet2)
			if status != -1 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if !sliceCompare(pac, nil) {
				t.Errorf("return wrong packet")
			}

			// Part 3
			packet3 := packet[20:]
			pac, status = teo.PacketCheck(packet3)
			if status != 0 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if !sliceCompare(pac, packet) {
				t.Errorf("return wrong packet")
			}
		}
	})

	// check and combine two valid packets splitted in one buffer
	t.Run("splittedValidPacketToBuf", func(t *testing.T) {
		teo.ResetReadBuf() // reset read buffer
		packet1, _ := teo.PacketCreate(cmd, peer, append([]byte(msg), 1))
		packet2, _ := teo.PacketCreate(cmd, peer, append([]byte(msg), 2))
		if packet1 != nil && packet2 != nil {
			buf := append(packet1, packet2...)
			//fmt.Println("combined packet:", buf)

			// Get first valid packet from buffer (now the teocli read buffer is
			// empty and input function buffer contains two valid packet).
			// The PacketCheck return one valid teocli packet and save next one to
			// read buffer
			pac, status := teo.PacketCheck(buf)
			if status != 0 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if pac == nil {
				t.Errorf("return wrong packet")
			}
			if !sliceCompare(pac, packet1) {
				t.Errorf("\n"+
					"a packet: %v\n"+
					"wait for: %v", pac, packet1)
			}
			//fmt.Println("first packet:   ", pac)

			// Get next valid packet saved in teocli read buffer (now the teocli read
			// buffer contains one valid packet and input function buffer is nil).
			// The PacketCheck return one valid teocli packet from read buffer and
			// free it
			pac, status = teo.PacketCheck(nil)
			if status != 0 {
				t.Errorf("return wrong status %d for valid packet", status)
			}
			if pac == nil {
				t.Errorf("return wrong packet")
			}
			if !sliceCompare(pac, packet2) {
				t.Errorf("\n"+
					"a packet: %v\n"+
					"wait for: %v", pac, packet1)
			}
			//fmt.Println("second packet:  ", pac)

			// Try to get next packet - result should be nil, read buffer - nil
			pac, status = teo.PacketCheck(nil)
			if status != -1 {
				t.Errorf("return wrong status %d for empty packet", status)
			}
			if pac != nil {
				t.Errorf("return wrong packet")
			}
			if !sliceCompare(teo.readBuffer, nil) {
				t.Errorf("wrong content of teocli read buffer: %v", teo.readBuffer)
			}

		}
	})

	// check and combine two splitted invalid packets
	t.Run("splittedInvalidPacket", func(t *testing.T) {
		teo.ResetReadBuf() // reset read buffer
		if packet, _ := teo.PacketCreate(cmd, peer, []byte(msg)); packet != nil {
			packet1 := packet[:20]
			pac, status := teo.PacketCheck(packet1)
			if status != -1 {
				t.Errorf("return wrong status %d for invalid packet", status)
			}
			if pac != nil {
				t.Errorf("return wrong %d bytes packet: %v", len(pac), pac)
			}

			packet2 := packet[10:]
			pac, status = teo.PacketCheck(packet2)
			if status != 1 {
				t.Errorf("return wrong status %d for invalid packet", status)
			}
			if sliceCompare(pac, packet) {
				t.Errorf("return valid packet")
			}
		}
	})

	// check and combine splitted 9 invalid packets with length = 1
	t.Run("splittedInvalidPacketSmallFirst", func(t *testing.T) {
		teo.ResetReadBuf() // reset read buffer
		if packet, _ := teo.PacketCreate(cmd, peer, []byte(msg)); packet != nil {
			packet1 := []byte("z")
			waitStatus := -1
			for i := 0; i < 9; i++ {
				pac, status := teo.PacketCheck(packet1)
				if i == 8 {
					waitStatus = 1
				}
				if status != waitStatus {
					t.Errorf("return wrong status %d for valid packet", status)
				}
				if !sliceCompare(pac, nil) {
					t.Errorf("return wrong packet")
				}
			}
		}
	})
}

// Test bytes packet
func TestBytes(t *testing.T) {

	var b bytes.Buffer // A Buffer needs no initialization.
	b.Write([]byte("Hello "))
	fmt.Fprintf(&b, "world!")
	//b.WriteTo(os.Stdout)
	sliceCompare(b.Bytes(), []byte("Hello world!"))
}

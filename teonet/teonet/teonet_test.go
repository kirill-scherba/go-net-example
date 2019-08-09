package teonet

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPacket(t *testing.T) {

	t.Run("packetCreate", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			cmd := 129 + i
			var from string
			var data []byte
			if i > 0 {
				from = "pp-mun-2"
				data = []byte("Hello!")
			}
			// Check packet create
			pac := packetCreateNew(cmd, from, data)
			if pac.Len() != pac.FromLen()+pac.DataLen()+2 {
				t.Errorf("wrong packet length: %d", pac.Len())
			}
			if cmd != pac.Cmd() {
				t.Errorf("wrong Cmd in created packet: %d", pac.Cmd())
			}
			if from != pac.From() {
				t.Errorf("wrong From in created packet: %s", pac.From())
			}
			if len(from)+1 != pac.FromLen() {
				t.Errorf("wrong From length in created packet: %d", pac.FromLen())
			}
			if !bytes.Equal(data, pac.Data()) {
				t.Errorf("wrong Data in created packet: %v", pac.Data())
			}
			if pac.DataLen() != len(pac.Data()) {
				t.Errorf("wrong Data length in created packet: %d", pac.DataLen())
			}
			// Check Parse
			rd, err := pac.Parse()
			if err != nil {
				t.Error(err)
			}
			if int(rd.raw_data_len) != pac.Len() {
				t.Errorf("wrong Packet length in rd: %d", rd.raw_data_len)
			}
			if int(rd.cmd) != pac.Cmd() {
				t.Errorf("wrong Cmd in rd: %d", rd.data_len)
			}
			if rd.From() != pac.From() {
				t.Errorf("wrong From in rd: %s", rd.From())
			}
			if int(rd.from_len) != pac.FromLen() {
				t.Errorf("wrong From length in rd: %d", rd.from_len)
			}
			if !bytes.Equal(rd.Data(), pac.Data()) {
				t.Errorf("wrong Data in rd: %v", rd.Data())
			}
			if int(rd.data_len) != pac.DataLen() {
				t.Errorf("wrong Data length in rd: %d", rd.data_len)
			}
			fmt.Printf(""+
				"packet: %v\n"+
				"packet length: %d\n"+
				"cmd: %d\n"+
				"from: %s\n"+
				"fromLen: %d\n"+
				"data: %v\n"+
				"dataLen: %d\n"+
				"%v\n",
				pac.packet, pac.Len(),
				pac.Cmd(), pac.From(), pac.FromLen(), pac.Data(), pac.DataLen(),
				rd,
			)
		}
	})
}

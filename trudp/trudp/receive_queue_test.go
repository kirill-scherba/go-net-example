package trudp

import (
	"strconv"
	"testing"
)

func TestReceiveQueue(t *testing.T) {

	const numElements = 10

	trudp := &TRUDP{}
	pac := &packetType{trudp: trudp}
	tcd := &ChannelData{trudp: trudp, receiveQueue: receiveQueueInit()}

	// create 10 elements 0..9
	t.Run("adding elements", func(t *testing.T) {
		for i := 0; i < numElements; i++ {
			packet := pac.newData(uint32(i), 0, []byte("hello"+strconv.Itoa(i))).copy()
			tcd.receiveQueueAdd(packet)
		}
	})

	// find existing elements and check packet
	t.Run("find all existing", func(t *testing.T) {
		for id := 0; id < numElements; id++ {
			rqd, ok := tcd.receiveQueueFind(uint32(id))
			switch {

			case !ok:
				t.Errorf("can't find existing element with id: %d", id)

			case uint32(id) != rqd.packet.ID():
				t.Errorf("wrong id: %d, should be: %d", id, rqd.packet.ID())

			case string(rqd.packet.Data()) != "hello"+strconv.Itoa(id):
				t.Errorf("wrong data in packet with id: %d", id)

			}
		}
	})

	// find and remove element then find it again
	t.Run("find and remove", func(t *testing.T) {
		id := uint32(5)
		_, ok := tcd.receiveQueueFind(id)
		if !ok {
			t.Errorf("does not existing tests element with id: %d", id)
		}
		tcd.receiveQueueRemove(id)

		_, ok = tcd.receiveQueueFind(id)
		if !ok {
			t.Errorf("found does not existing element with id: %d", id)
		}
	})

	// reset queue (remove all elements)
	t.Run("reset queue", func(t *testing.T) {
		tcd.receiveQueueReset()
		for id := 0; id < numElements; id++ {
			_, ok := tcd.receiveQueueFind(uint32(id))
			if ok {
				t.Errorf("after reset was found does not existing element with id: %d", id)
			}
		}
	})

}

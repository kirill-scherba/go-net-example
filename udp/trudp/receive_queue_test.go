package trudp

import (
	"strconv"
	"testing"
)

func TestReceiveQueue(t *testing.T) {

	const numElements = 10

	trudp := &TRUDP{}
	pac := &packetType{trudp: trudp}
	tcd := &channelData{trudp: trudp}
	trudp.log(NONE, "TestReceiveQueue initialized")

	// create 10 elements 0..9
	t.Run("adding elements", func(t *testing.T) {
		for i := 0; i < numElements; i++ {
			packet := pac.dataCreateNew(uint(i), 0, []byte("hello"+strconv.Itoa(i))).copy()
			tcd.receiveQueueAdd(packet)
		}
	})

	// find existing elements and check packet
	t.Run("find all existing", func(t *testing.T) {
		for id := 0; id < numElements; id++ {
			idx, rqd, err := tcd.receiveQueueFind(uint(id))
			switch {

			case err != nil:
				t.Errorf("can't find existing element with id: %d", id)

			case uint(idx) != rqd.packet.getID():
				t.Errorf("wrong index: %d, should be: %d", idx, id)

			case string(rqd.packet.getData()) != "hello"+strconv.Itoa(id):
				t.Errorf("wrong data in packet with id: %d", id)

			}
		}
	})

	// find and remove element then find it again
	t.Run("find and remove", func(t *testing.T) {
		id := uint(5)
		idx, _, err := tcd.receiveQueueFind(id)
		if err != nil {
			t.Errorf("does not existing tests element with id: %d", id)
		}
		tcd.receiveQueueRemove(idx)

		_, _, err = tcd.receiveQueueFind(id)
		if err == nil {
			t.Errorf("found does not existing element with id: %d", id)
		}
	})

	// reset queue (remove all elements)
	t.Run("reset queue", func(t *testing.T) {
		tcd.receiveQueueReset()
		for id := 0; id < numElements; id++ {
			_, _, err := tcd.receiveQueueFind(uint(id))
			if err == nil {
				t.Errorf("after reset was found does not existing element with id: %d", id)
			}
		}
	})

}

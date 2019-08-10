package trudp

import (
	"container/list"
	"strconv"
	"testing"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

func TestReceiveQueue(t *testing.T) {

	const numElements = 10

	trudp := &TRUDP{}
	pac := &packetType{trudp: trudp}
	tcd := &ChannelData{trudp: trudp, receiveQueue: list.New()}
	teolog.Log(teolog.NONE, "TestReceiveQueue initialized")

	// create 10 elements 0..9
	t.Run("adding elements", func(t *testing.T) {
		for i := 0; i < numElements; i++ {
			packet := pac.dataCreateNew(uint32(i), 0, []byte("hello"+strconv.Itoa(i))).copy()
			tcd.receiveQueueAdd(packet)
		}
	})

	// find existing elements and check packet
	t.Run("find all existing", func(t *testing.T) {
		for id := 0; id < numElements; id++ {
			_, rqd, err := tcd.receiveQueueFind(uint32(id))
			switch {

			case err != nil:
				t.Errorf("can't find existing element with id: %d", id)

			case uint32(id) != rqd.packet.getID():
				t.Errorf("wrong id: %d, should be: %d", id, rqd.packet.getID())

			case string(rqd.packet.getData()) != "hello"+strconv.Itoa(id):
				t.Errorf("wrong data in packet with id: %d", id)

			}
		}
	})

	// find and remove element then find it again
	t.Run("find and remove", func(t *testing.T) {
		id := uint32(5)
		e, _, err := tcd.receiveQueueFind(id)
		if err != nil {
			t.Errorf("does not existing tests element with id: %d", id)
		}
		tcd.receiveQueueRemove(e)

		_, _, err = tcd.receiveQueueFind(id)
		if err == nil {
			t.Errorf("found does not existing element with id: %d", id)
		}
	})

	// reset queue (remove all elements)
	t.Run("reset queue", func(t *testing.T) {
		tcd.receiveQueueReset()
		for id := 0; id < numElements; id++ {
			_, _, err := tcd.receiveQueueFind(uint32(id))
			if err == nil {
				t.Errorf("after reset was found does not existing element with id: %d", id)
			}
		}
	})

}

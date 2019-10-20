package cdb

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/services/teoroom"
)

func TestCdb(t *testing.T) {

	roomID := gocql.TimeUUID()
	roomNum := 1
	var err error
	var db *cdb

	t.Run("Connect", func(t *testing.T) {
		db, err = newDb()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer db.close()

	t.Run("Set-Creating", func(t *testing.T) {
		// room := &Room{
		// 	ID:      roomID,
		// 	RoomNum: roomNum,
		// 	Started: time.Now(),
		// 	State:   teoroom.RoomCreating,
		// }
		// err = db.set(room)
		roomID, err = db.setCreating(roomNum)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Running", func(t *testing.T) {
		// room := &Room{
		// 	ID:    roomID,
		// 	State: teoroom.RoomRunning,
		// }
		// err = db.set(room, db.roomsMetadata.Columns[4])
		err = db.setRunning(roomID)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Stopped", func(t *testing.T) {
		room := &Room{
			ID:      roomID,
			State:   teoroom.RoomStopped,
			Stopped: time.Now(),
		}
		err = db.set(room, db.roomsMetadata.Columns[3], db.roomsMetadata.Columns[4])
		if err != nil {
			t.Error(err)
			return
		}
	})
}

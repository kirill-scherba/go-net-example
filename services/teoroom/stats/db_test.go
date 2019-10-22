package stats

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
)

func TestCdb(t *testing.T) {

	roomID := gocql.TimeUUID()
	roomNum := uint32(1)
	var err error
	var db *db

	t.Run("Connect", func(t *testing.T) {
		db, err = newDb("teoroom_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer db.close()

	t.Run("Set-Creating", func(t *testing.T) {
		err = db.setCreating(roomID, roomNum)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Running", func(t *testing.T) {
		err = db.setRunning(roomID)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Closed", func(t *testing.T) {
		err = db.setClosed(roomID)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Stopped", func(t *testing.T) {
		err = db.setStopped(roomID)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

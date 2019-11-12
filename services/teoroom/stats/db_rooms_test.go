package stats

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocql/gocql"
)

func TestDbRooms(t *testing.T) {

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

	t.Run("getByCreated", func(t *testing.T) {

		// const fromStr = "2019-10-01T00:00:00Z"
		// const toStr = "2019-10-02T00:00:00Z"
		// var from, to time.Time
		// var err error
		// from, err = time.Parse(time.RFC3339, fromStr)
		// to, err = time.Parse(time.RFC3339, toStr)

		now := time.Now()
		from := now.Add(-10 * time.Minute)
		to := now
		res, err := db.getByCreated(from, to, 1)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println("getByCreated:", res)
	})
}

package stats

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
)

func TestDbClients(t *testing.T) {

	roomID := gocql.TimeUUID()
	clientID := gocql.TimeUUID()
	gameStat := []byte("Hello")
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

	t.Run("Set-Added", func(t *testing.T) {
		err = db.setAdded(roomID, clientID)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-Leave", func(t *testing.T) {
		err = db.setLeave(roomID, clientID)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Set-GameStat", func(t *testing.T) {
		err = db.setGameStat(roomID, clientID, gameStat)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

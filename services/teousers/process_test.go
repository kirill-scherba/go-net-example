package teousers

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func TestProcess(t *testing.T) {

	const AppName = "teotest-7755-2"
	userID := gocql.TimeUUID()
	teo := &teonet.Teonet{}
	var err error
	var u *Users

	t.Run("Connect", func(t *testing.T) {
		u, err = Connect("teousers_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer u.Close()

	t.Run("ComCreateUser", func(t *testing.T) {
	})

	t.Run("ComCheckUser", func(t *testing.T) {
		data := userID.Bytes()
		pac := teo.PacketCreateNew("teo-from", 129, data)
		err = u.ComCheckUser(pac)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

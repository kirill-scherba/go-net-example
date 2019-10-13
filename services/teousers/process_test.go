package teousers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

func TestUserNew(t *testing.T) {
	var data []byte
	var in *UserNew
	var err error

	t.Run("Marshal", func(t *testing.T) {
		in = &UserNew{
			UserID:      gocql.TimeUUID(),
			AccessToken: gocql.TimeUUID(),
		}
		fmt.Printf("Input UserNew: %s, %s\n", in.UserID.String(),
			in.AccessToken.String())
		if data, err = in.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Marshal %d bytes UserNew: %v\n", len(data), data)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		out := &UserNew{}
		if err = out.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Unarshal UserNew: %s, %s\n", out.UserID.String(),
			out.AccessToken.String())
		if !(in.UserID == out.UserID && in.AccessToken == out.AccessToken) {
			t.Error(errors.New("output UserNew not equal to input UserNew"))
			return
		}
	})
}

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

	t.Run("ComCheckUser", func(t *testing.T) {
		data := userID.Bytes()
		pac := teo.PacketCreateNew("teo-from", 129, data)
		exists, err := u.ComCheckUser(pac)
		if err != nil {
			t.Error(err)
			return
		}
		if exists {
			t.Error(errors.New("return true when user does not exists"))
			return
		}
	})

	// t.Run("ComCreateUser", func(t *testing.T) {
	// 	pac := teo.PacketCreateNew("teo-from", 129, nil)
	// 	u.ComCreateUser(pac)
	// })

	// t.Run("ComCheckUser", func(t *testing.T) {
	// 	data := userID.Bytes()
	// 	pac := teo.PacketCreateNew("teo-from", 129, data)
	// 	exists, err := u.ComCheckUser(pac)
	// 	if err != nil {
	// 		t.Error(err)
	// 		return
	// 	}
	// 	if !exists {
	// 		t.Error(errors.New("return false when user exists"))
	// 		return
	// 	}
	// })
}

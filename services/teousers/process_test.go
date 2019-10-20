package teousers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	cli "github.com/kirill-scherba/teonet-go/services/teouserscli"
	"github.com/kirill-scherba/teonet-go/teonet/teonet"
)

type Teoemu struct{}

func (t *Teoemu) SendTo(peer string, cmd byte, data []byte) (int, error)         { return 0, nil }
func (t *Teoemu) SendAnswer(pac interface{}, cmd byte, data []byte) (int, error) { return 0, nil }

func TestUserNew(t *testing.T) {
	var data []byte
	var in *cli.UserResponce
	var err error

	t.Run("Marshal", func(t *testing.T) {
		in = &cli.UserResponce{
			ID:          gocql.TimeUUID(),
			AccessToken: gocql.TimeUUID(),
			Prefix:      "game001",
		}
		fmt.Printf("Input UserNew: %s, %s, %s\n", in.ID.String(),
			in.AccessToken.String(), in.Prefix)
		if data, err = in.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Marshal %d bytes UserNew: %v\n", len(data), data)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		out := &cli.UserResponce{}
		if err = out.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Unmarshal UserNew: %s, %s, %s\n", out.ID.String(),
			out.AccessToken.String(), out.Prefix)
		if !(in.ID == out.ID && in.AccessToken == out.AccessToken) {
			t.Error(errors.New("output UserNew not equal to input UserNew"))
			return
		}
	})
}

func TestProcess(t *testing.T) {

	const AppName = "teotest-7755-2"
	userID := gocql.TimeUUID()
	teoemu := &Teoemu{}
	teo := &teonet.Teonet{}
	var userNew *cli.UserResponce
	var err error
	var u *Users

	// Connect
	t.Run("Connect", func(t *testing.T) {
		u, err = Connect(teoemu, "teousers_test")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("Connected to database\n")
	})
	defer u.Close()

	// Check non existing user
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

	// Create user using wrong rquest data (wron prefix)
	t.Run("ComCreateUserWrong", func(t *testing.T) {
		pac := teo.PacketCreateNew("teo-from", 129, []byte("tg001-new"))
		userNew, err = u.ComCreateUser(pac)
		if err == nil {
			t.Error(errors.New("ComCreateUser request = \"tg001-new\" should return error"))
			return
		}
	})
	if err == nil {
		return
	}

	// Create user
	t.Run("ComCreateUser", func(t *testing.T) {
		pac := teo.PacketCreateNew("teo-from", 129, []byte("tg001"))
		userNew, err = u.ComCreateUser(pac)
		if err != nil {
			t.Fatal(err)
			return
		}
	})
	if err != nil {
		return
	}

	// Check existing user by binary user id
	t.Run("ComCheckUser", func(t *testing.T) {
		data := userNew.ID.Bytes()
		pac := teo.PacketCreateNew("teo-from", 129, data)
		exists, err := u.ComCheckUser(pac)
		if err != nil {
			t.Error(err)
			return
		}
		if !exists {
			t.Error(errors.New("return false when user exists"))
			return
		}
	})

	// Check existing user by text with prefix and user id
	t.Run("ComCheckUser", func(t *testing.T) {
		data := []byte("tg001-" + userNew.ID.String())
		pac := teo.PacketCreateNew("teo-from", 129, data)
		exists, err := u.ComCheckUser(pac)
		if err != nil {
			t.Error(err)
			return
		}
		if !exists {
			t.Error(errors.New("return false when user exists"))
			return
		}
	})

	// Remove test user to clear db
	u.delete(userNew)
}

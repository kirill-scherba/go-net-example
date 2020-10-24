package stats

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"unsafe"

	"github.com/gocql/gocql"
)

var ErrWrongUnmarshalData = errors.New("wrong unmarshal data")

func TestRoomCreateRequest(t *testing.T) {
	var data []byte
	var err error
	t.Run("MarshalUnmarshal", func(t *testing.T) {

		var roomNum uint32 = 123
		var roomID = gocql.TimeUUID()

		req := &RoomCreateRequest{roomID, roomNum}
		if data, err = req.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}

		if err = req.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		if req.RoomNum != roomNum {
			err = ErrWrongUnmarshalData
			t.Error(err)
			return
		}
	})
}

func ExampleRoomCreateRequest_MarshalBinary() {
	var roomID = gocql.TimeUUID()
	req := RoomCreateRequest{roomID, 123}
	data, _ := req.MarshalBinary()
	l := int(unsafe.Sizeof(roomID))
	fmt.Println(data[l:])
	// Output: [123 0 0 0]
}

func ExampleRoomCreateRequest_UnmarshalBinary() {
	var roomID = gocql.TimeUUID()
	data := append(roomID.Bytes(), []byte{123, 0, 0, 0}...)
	req := &RoomCreateRequest{}
	req.UnmarshalBinary(data)
	fmt.Println(req.RoomNum)
	// Output: 123
}

func BenchmarkRoomCreateRequestMarshalBinary(b *testing.B) {
	var roomID = gocql.TimeUUID()
	req := &RoomCreateRequest{roomID, 123}
	req.MarshalBinary()
}

func BenchmarkRoomCreateRequestUnmarshalBinary(b *testing.B) {
	data := []byte{123, 0, 0, 0}
	req := &RoomCreateRequest{}
	req.UnmarshalBinary(data)
}

func TestRoomCreateResponce(t *testing.T) {
	var data []byte
	var err error
	t.Run("MarshalUnmarshal", func(t *testing.T) {

		var RoomID = gocql.TimeUUID()

		res := &RoomCreateResponce{RoomID}
		if data, err = res.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}

		if err = res.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		if res.RoomID != RoomID {
			err = ErrWrongUnmarshalData
			t.Error(err)
			return
		}
	})
}

func ExampleRoomCreateResponce_MarshalBinary() {
	roomID, _ := gocql.ParseUUID("a5f8a6b5-f39e-11e9-adbc-40a3cc55de62")
	res := &RoomCreateResponce{roomID}
	data, _ := res.MarshalBinary()
	fmt.Println(data)
	// Output: [165 248 166 181 243 158 17 233 173 188 64 163 204 85 222 98]
}

func ExampleRoomCreateResponce_UnmarshalBinary() {
	data := []byte{
		165, 248, 166, 181, 243, 158, 17, 233, 173, 188, 64, 163, 204, 85, 222,
		98,
	}
	res := &RoomCreateResponce{}
	res.UnmarshalBinary(data)
	fmt.Println(res.RoomID)
	// Output: a5f8a6b5-f39e-11e9-adbc-40a3cc55de62
}

func BenchmarkRoomCreateResponceMarshalBinary(b *testing.B) {
	roomID, _ := gocql.ParseUUID("a5f8a6b5-f39e-11e9-adbc-40a3cc55de62")
	res := &RoomCreateResponce{roomID}
	res.MarshalBinary()
}

func BenchmarkRoomCreateResponceUnmarshalBinary(b *testing.B) {
	data := []byte{
		165, 248, 166, 181, 243, 158, 17, 233, 173, 188, 64, 163, 204, 85, 222,
		98,
	}
	res := &RoomCreateResponce{}
	res.UnmarshalBinary(data)
}

func TestClientStatusRequest(t *testing.T) {
	var data []byte
	var err error
	t.Run("MarshalUnmarshal", func(t *testing.T) {

		roomID, _ := gocql.ParseUUID("a5f8a6b5-f39e-11e9-adbc-40a3cc55de62")
		id, _ := gocql.ParseUUID("66d68dc9-f4ed-11e9-bd5f-107b4445878a")
		var gameStat = []byte("Hello!")

		// i = 0 - full request; i = 1 - request with GameStat = nil
		for i := 0; i < 2; i++ {

			req := &ClientStateRequest{
				RoomID:   roomID,
				ID:       id,
				GameStat: gameStat,
			}

			if data, err = req.MarshalBinary(); err != nil {
				t.Error(err)
				return
			}

			req = &ClientStateRequest{}
			if err = req.UnmarshalBinary(data); err != nil {
				t.Error(err)
				return
			}

			if !(req.RoomID.String() == roomID.String() &&
				req.ID.String() == id.String() &&
				string(req.GameStat) == string(gameStat)) {
				t.Error(errors.New("ivalid unmarshalled values"))
				fmt.Printf(" i = %d, %v\n", i, req)
				return
			}

			gameStat = nil
		}

	})
}

func TestRoomByCreatedResponce(t *testing.T) {

	var data []byte
	var err error
	t.Run("MarshalUnmarshal", func(t *testing.T) {

		now := time.Now()
		id := gocql.TimeUUID()
		resq := &RoomByCreatedResponce{
			ReqID: 717,
			Rooms: []Room{Room{
				ID:      id,
				RoomNum: 919,
				Created: now,
				Started: now.Add(time.Second * 10),
				Closed:  now.Add(time.Second * 20),
				Stopped: now.Add(time.Second * 20),
				State:   232,
			}},
		}

		data, err = resq.MarshalBinary()
		if err != nil {
			t.Error(err)
			return
		}

		fmt.Println("resq.MarshalBinary():", resq, data, len(data))

		res := &RoomByCreatedResponce{}
		err = res.UnmarshalBinary(data)
		fmt.Println("res.UnmarshalBinary(data):", res, data)
		if err != nil {
			t.Error(err)
			return
		}
		if res.ReqID != resq.ReqID {
			err = errors.New("UnmarshalBinary err, wrong ReqID")
			t.Error(err)
			return
		}
		if res.Rooms[0].RoomNum != resq.Rooms[0].RoomNum {
			err = errors.New("UnmarshalBinary err, wrong RoomNum")
			t.Error(err)
			return
		}
		// if res.Rooms[0].Created != resq.Rooms[0].Created {
		// 	err = errors.New("UnmarshalBinary err, wrong Created")
		// 	t.Error(err)
		// 	return
		// }
		// if res.Rooms[0].Started != resq.Rooms[0].Started {
		// 	err = errors.New("UnmarshalBinary err, wrong Started")
		// 	t.Error(err)
		// 	return
		// }
		// if res.Rooms[0].Closed != resq.Rooms[0].Closed {
		// 	err = errors.New("UnmarshalBinary err, wrong Closed")
		// 	t.Error(err)
		// 	return
		// }
		// if res.Rooms[0].Stopped != resq.Rooms[0].Stopped {
		// 	err = errors.New("UnmarshalBinary err, wrong Stopped")
		// 	t.Error(err)
		// 	return
		// }
		if res.Rooms[0].State != resq.Rooms[0].State {
			err = errors.New("UnmarshalBinary err, wrong State")
			t.Error(err)
			return
		}
	})
}

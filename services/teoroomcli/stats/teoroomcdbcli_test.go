package stats

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gocql/gocql"
)

var ErrWrongUnmarshalData = errors.New("wrong unmarshal data")

func TestRoomCreateRequest(t *testing.T) {
	var data []byte
	var err error
	t.Run("MarshalUnmarshal", func(t *testing.T) {

		var roomNum uint32 = 123
		var roomID = gocql.TimeUUID()

		rec := &RoomCreateRequest{roomID, roomNum}
		if data, err = rec.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}

		if err = rec.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		if rec.RoomNum != roomNum {
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
	fmt.Println(data)
	// Output: [123 0 0 0]
}

func ExampleRoomCreateRequest_UnmarshalBinary() {
	data := []byte{123, 0, 0, 0}
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

		rec := &RoomCreateResponce{RoomID}
		if data, err = rec.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}

		if err = rec.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		if rec.RoomID != RoomID {
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

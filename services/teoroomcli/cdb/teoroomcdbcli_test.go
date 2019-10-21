package cdb

import (
	"errors"
	"fmt"
	"testing"
)

func TestCdb(t *testing.T) {

	var data []byte
	var err error

	t.Run("RoomCreateRequest", func(t *testing.T) {

		var roomNum uint32 = 123

		rec := &RoomCreateRequest{roomNum}
		if data, err = rec.MarshalBinary(); err != nil {
			t.Error(err)
			return
		}

		if err = rec.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}

		if rec.RoomNum != roomNum {
			err = errors.New("wrong unmarshal data")
			t.Error(err)
			return
		}
	})
}

func ExampleRoomCreateRequest_MarshalBinary() {
	rec := RoomCreateRequest{123}
	data, _ := rec.MarshalBinary()
	fmt.Println(data)
	// Output: [123 0 0 0]
}

func ExampleRoomCreateRequest_UnmarshalBinary() {
	data := []byte{123, 0, 0, 0}
	rec := &RoomCreateRequest{}
	rec.UnmarshalBinary(data)
	fmt.Println(rec.RoomNum)
	// Output: 123
}

func BenchmarkMarshalBinary(b *testing.B) {
	rec := RoomCreateRequest{123}
	rec.MarshalBinary()
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	data := []byte{123, 0, 0, 0}
	rec := &RoomCreateRequest{}
	rec.UnmarshalBinary(data)
}

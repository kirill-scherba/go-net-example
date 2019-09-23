package teoroom

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
)

// Clients commands (commands executet in client) -----------------------------

// TeoConnector is teonet connector interface. It may be servers (*Teonet) or
// clients (*TeoLNull) connector
type TeoConnector interface {
	SendTo(peer string, cmd byte, data []byte) (int, error)
}

// RoomRequest send room request from client
func RoomRequest(con TeoConnector, peer string, i interface{}) {
	switch data := i.(type) {
	case nil:
		con.SendTo(peer, ComRoomRequest, nil)
	case []byte:
		con.SendTo(peer, ComRoomRequest, data)
	default:
		err := fmt.Errorf("Invalid type %T in SendTo function", data)
		panic(err)
	}
}

// Disconnect send disconnect command
func Disconnect(con TeoConnector, peer string, i interface{}) {
	switch data := i.(type) {
	case nil:
		con.SendTo(peer, ComDisconnect, nil)
	case []byte:
		con.SendTo(peer, ComDisconnect, data)
	case byte:
		con.SendTo(peer, ComDisconnect, append([]byte{}, data))
	default:
		err := fmt.Errorf("Invalid type %T in SendTo function", data)
		panic(err)
	}
}

// SendData send data from client
func SendData(con TeoConnector, peer string, ii ...interface{}) (num int, err error) {
	buf := new(bytes.Buffer)
	for _, i := range ii {
		switch d := i.(type) {
		case nil:
			err = binary.Write(buf, binary.LittleEndian, "nil")
		case encoding.BinaryMarshaler:
			var dd []byte
			if dd, err = d.MarshalBinary(); err == nil {
				err = binary.Write(buf, binary.LittleEndian, dd)
			}
		case int:
			err = binary.Write(buf, binary.LittleEndian, uint64(d))
		default:
			err = binary.Write(buf, binary.LittleEndian, d)
		}
		if err != nil {
			return
		}
	}
	return con.SendTo(peer, ComRoomData, buf.Bytes())
}

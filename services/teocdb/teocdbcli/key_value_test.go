// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package teocdbcli

import (
	"fmt"
	"testing"
)

func TestKeyValueBinary(t *testing.T) {

	t.Run("MarshalUnmarshalBinary", func(t *testing.T) {

		cmd := byte(1)
		id := uint16(2)
		key := "test.key.123"
		value := []byte("Hello world!")

		bdInput := &keyValue{cmd, id, key, value, false}
		bdOutput := &keyValue{}
		data, err := bdInput.MarshalBinary()
		//fmt.Println(data)
		if err != nil {
			t.Error(err)
			return
		}
		if err = bdOutput.UnmarshalBinary(data); err != nil {
			t.Error(err)
			return
		}
		if bdOutput.Cmd != cmd || bdOutput.ID != id || bdOutput.Key != key ||
			string(value) != string(bdOutput.Value) {
			t.Errorf("unmarshalled structure fields values"+
				" not equal to input values:\n%d, %d, '%s', '%s'\n"+
				"data: %v",
				bdOutput.Cmd, bdOutput.ID, bdOutput.Key, string(bdOutput.Value),
				data,
			)
			return
		}
		fmt.Printf(
			"data: %v\n"+"cmd: %d\n"+"id: %d\n"+"key: %s\n"+"value: %s\n",
			data, bdOutput.Cmd, bdOutput.ID, bdOutput.Key, string(bdOutput.Value))
	})
}
func TestKeyValueText(t *testing.T) {

	t.Run("UnmarshalText", func(t *testing.T) {

		id := uint16(77)
		key := "test.key.5678"
		value := []byte("Hello!")

		type fields struct {
			Cmd           byte
			ID            uint16
			Key           string
			Value         []byte
			requestInJSON bool
		}
		type args struct {
			text []byte
		}
		tests := []struct {
			name    string
			fields  fields
			args    args
			wantErr bool
		}{
			// Test cases:
			{"text-key", fields{Key: key},
				args{[]byte(fmt.Sprintf("%s", key) /* "test.key.5678" */)}, false},

			{"text-key-id-", fields{Key: key, ID: id},
				args{[]byte(fmt.Sprintf("%s,%d,", key, id) /* "test.key.5678,77," */)}, false},

			{"text-key-id", fields{Key: key, ID: id, Cmd: CmdGet},
				args{[]byte(fmt.Sprintf("%s,%d", key, id) /* "test.key.5678,77," */)}, false},

			{"text-key-value", fields{Key: key, Value: value},
				args{[]byte(fmt.Sprintf("%s,%s", key, string(value)) /* "test.key.5678,Hello!" */)}, false},

			{"text-key-id-value", fields{Key: key, ID: id, Value: value},
				args{[]byte(fmt.Sprintf("%s,%d,%s", key, id, string(value)) /* "test.key.5678,77,Hello!" */)}, false},

			{"text-key-id-value", fields{Key: key, ID: id, Value: []byte("55")},
				args{[]byte(fmt.Sprintf("%s,%d,%d", key, id, 55) /* "test.key.5678,77,55" */)}, false},

			{"json-empty", fields{},
				args{[]byte(fmt.Sprintf(`{}`))}, true},

			{"json-key", fields{Key: key},
				args{[]byte(fmt.Sprintf(`{"key":"%s"}`, key) /* "test.key.5678" */)}, false},

			{"json-key-id", fields{Key: key, ID: id},
				args{[]byte(fmt.Sprintf(`{"key":"%s","id":%d}`, key, id))}, false},

			{"json-key-id-value", fields{Key: key, ID: id, Value: value},
				args{[]byte(fmt.Sprintf(`{"key":"%s","id":%d,"value":"%s"}`, key, id, string(value)))}, false},

			{"json-key-id-value", fields{Key: key, ID: id, Value: []byte("55")},
				args{[]byte(fmt.Sprintf(`{"key":"%s","id":%d,"value":%d}`, key, id, 55))}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				// Unmarshal
				kv := &keyValue{Cmd: tt.fields.Cmd}
				err := kv.UnmarshalText(tt.args.text)
				if (err != nil) != tt.wantErr {
					t.Errorf("KeyValue.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Display result
				fmt.Printf("%s: %s -> KeyValue%v\n", func() string {
					if kv.RequestInJSON {
						return "json"
					}
					return "text"
				}(), string(tt.args.text), kv)

				// If want error and has unmarshal error
				if err != nil && tt.wantErr {
					return
				}

				// Check values
				if isErr := (kv.Key != tt.fields.Key); isErr != tt.wantErr {
					t.Errorf("Key not valid: %s", kv.Key)
					return
				}
				if isErr := (kv.ID != tt.fields.ID); isErr != tt.wantErr {
					t.Errorf("Id not valid: %d, need: %d", kv.ID, tt.fields.ID)
					return
				}
				if isErr := (string(kv.Value) != string(tt.fields.Value)); isErr != tt.wantErr {
					t.Errorf("Value not valid: %s (%v)", string(kv.Value), kv)
					return
				}

				// Marshal
				d, err := kv.MarshalText()
				if err != nil {
					t.Errorf("KeyValue.MarshalText() error = %v, wantErr %v",
						err, tt.wantErr)
					return
				}
				fmt.Printf("      %s\n", string(d))
			})
		}
	})
}

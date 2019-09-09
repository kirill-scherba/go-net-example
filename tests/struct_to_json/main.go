package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func StructToJSON(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {

	var data = struct {
		IVal int    `json:"i_val"`
		SVal string `json:"s_val"`
	}{
		25,
		"this is 25",
	}

	jsonBytes, err := StructToJSON(data)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Printf("%s", string(jsonBytes))
}

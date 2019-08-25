package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
)

// Test creating raw byte buffer from go struct

func main() {

	fmt.Println("Buf test application ver 0.0.1")

	type data struct {
		q  byte
		b  [3]byte
		q2 byte
		m  uint32
	}

	myData := data{1, [3]byte{2, 3, 4}, 5, 0x01287353}

	stdoutDumper := hex.Dumper(os.Stdout)

	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, myData)

	fmt.Printf("buf: %v\n", buf.Bytes())
	stdoutDumper.Write(buf.Bytes())
	fmt.Printf("\n")
}

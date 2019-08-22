package main

import "C"

// GoTst a go test function return cstring, it should be free after use
//export GoTst
func GoTst(name string) *C.char {
	str := "Hello " + name + "!"
	return C.CString(str)
}

func main() {}

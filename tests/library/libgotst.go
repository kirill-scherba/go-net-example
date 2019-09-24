// This example shows how to create dynamic library using go and use it in C
// application.
package main

import "C"

// GoTst a go test function return cstring, it should be free after use
//export GoTst
func GoTst(name string) *C.char {
	str := "Hello " + name + "!"
	return C.CString(str)
}

func main() {}

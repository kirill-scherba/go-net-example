package main

import (
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("Hello, WebAssembly!")
	js.Global().Get("document").Call("getElementsByTagName", "body").
		Index(0).Set("innerHTML", "Hello, World!")
}

package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

// Run much goroutines to calculate memory used

func main() {
	fmt.Printf("Goroutines sample app ver. 1.0.0\n")

	const (
		numDefault     = 1000
		timeoutDefault = 30
	)

	var num int
	var wg sync.WaitGroup
	var timeout int

	// Read application flag
	flag.IntVar(&num, "n", numDefault, "number of goroutines")
	flag.IntVar(&timeout, "t", timeoutDefault, "time of one goroutine running in sec")
	flag.Parse()

	// Run gooroutines which sleep during timeout end exit
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			wg.Done()
		}()
	}
	wg.Wait()
}

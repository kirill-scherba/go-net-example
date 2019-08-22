package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

func main() {

	running := false //true
	const count = 500000
	quit := make(chan bool)
	t := time.Now()

	var tostop int64 = 0
	for i := 0; i < count; i++ {
		go func(num int) {
			for {
				if !running {
					if atomic.AddInt64(&tostop, 1) == count {
						quit <- true
					}
					break
				}
				time.Sleep(1 * time.Microsecond)
				//log.Println("go #", num)
			}
		}(i)
	}

	// fmt.Println("press Enter to stop")
	// fmt.Scanf("\n")

	running = false
	//t := time.Now()
	<-quit

	fmt.Println(time.Since(t))
}

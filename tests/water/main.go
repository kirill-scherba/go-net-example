// Tap interface create sample
//
// Build it firsf:
//
// 	go get
//  go build
//
// Then run:
//
// 	./water
//
// And In a new terminal:
//
//	sudo ip addr add 10.1.0.10/24 dev O_O
//	sudo ip link set dev O_O up

package main

import (
	"log"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
)

func main() {
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = "O_0"

	ifce, err := water.New(config)
	if err != nil {
		log.Fatal(err)
	}
	var frame ethernet.Frame

	for {
		frame.Resize(1500)
		n, err := ifce.Read([]byte(frame))
		if err != nil {
			log.Fatal(err)
		}
		frame = frame[:n]
		log.Printf("Dst: %s\n", frame.Destination())
		log.Printf("Src: %s\n", frame.Source())
		log.Printf("Ethertype: % x\n", frame.Ethertype())
		log.Printf("Payload: % x\n", frame.Payload())
	}
}

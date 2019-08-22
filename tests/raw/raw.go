package main

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {
	protocol := "icmp" // "gre" "tcp"
	netaddr, _ := net.ResolveIPAddr("ip4", "0.0.0.0")
	conn, _ := net.ListenIP("ip4:"+protocol, netaddr)

	buf := make([]byte, 1024)

	for {
		numRead, addr, _ := conn.ReadFrom(buf)

		// Decode a packet
		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)

		for _, layer := range packet.Layers() {
			fmt.Println("PACKET LAYER:", layer.LayerType())
		}

		fmt.Printf("from: %s, network: %s\n% X\n\n",
			addr.String(), addr.Network(), buf[:numRead])
	}
}

package trudp

import (
	"fmt"
	"syscall"
	"time"
)

// BenchmarkSyscallUDP test function to connect to UDP with syscall packet
func BenchmarkSyscallUDP() { //t *testing.B) {

	fmt.Println("BenchmarkSyscallUDP initialized")

	//  Create UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		panic(err)
	}

	//  Bind UDP socket to local port so we can receive pings
	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: 9095, Addr: [4]byte{0, 0, 0, 0}}); err != nil {
		panic(err)
	}

	// Address to send
	addr := &syscall.SockaddrInet4{
		Port: 8000,
		Addr: [4]byte{127, 0, 0, 1},
	}

	flags := 0 //syscall.MSG_DONTWAIT
	buf := make([]byte, 2048)
	pac := &packetType{}

	// Read from socket
	go func() {
		for {
			n, _, err := syscall.Recvfrom(fd, buf, 0) // syscall.MSG_DONTWAIT)
			if err != nil {
				fmt.Println("err =", err)
			}
			// if n < 0 {
			// 	break
			// }

			fmt.Println("n =", n)
		}
	}()

	t := time.Now()
	for i := 0; i < 135000; i++ {
		if i%200 == 0 { // 200
			time.Sleep(850 * time.Microsecond) // 800
		}
		packet := pac.newData(uint32(i), 1, []byte("hello"))
		if err := syscall.Sendto(fd, packet.data, flags, addr); err != nil {
			panic(err)
		}
		pac.freeCreated(packet.data)
	}
	res := time.Since(t)

	time.Sleep(5 * time.Second)
	fmt.Println(res)
	syscall.Close(fd)
}

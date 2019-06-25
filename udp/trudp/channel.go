package trudp

import (
	"net"
	"strconv"
	"time"
)

type ChannelData struct {
	id uint
}

// ConnectChannel to remote host by UDP
func (trudp *TRUDP) ConnectChannel(rhost string, rport int, ch int) (tcd *ChannelData) {

	service := rhost + ":" + strconv.Itoa(rport)
	rUDPAddr, err := net.ResolveUDPAddr(network, service)
	if err != nil {
		panic(err)
	}
	trudp.log(CONNECT, "connecting to host", rUDPAddr, "at channel", ch)

	tcd = &ChannelData{
		id: 0,
	}

	// Send hello to remote host
	packet := trudp.packet.dataCreateNew(tcd.getId(), ch, []byte(helloMsg))
	defer trudp.packet.freeCreated(packet)
	trudp.conn.WriteToUDP(packet, rUDPAddr)

	// Keep alive: send Ping
	go func(conn *net.UDPConn) {
		for {
			time.Sleep(pingInterval * time.Millisecond)
			packet := trudp.packet.dataCreateNew(tcd.getId(), ch, []byte(echoMsg))
			defer trudp.packet.freeCreated(packet)
			trudp.conn.WriteToUDP(packet, rUDPAddr)
		}
	}(trudp.conn)

	return
}

// getId return new packe id
func (tcd *ChannelData) getId() uint {
	tcd.id++
	return tcd.id
}

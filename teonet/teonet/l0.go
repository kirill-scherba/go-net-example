// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet L0 server module:
//
// L0 server is intended to connect teonet clients to the teonet network. Teonet
// clients should use uses teocli packet which connect it to teonet L0 server.
// The L0 server allow tcp or trudp connection which incuded to teocli packet.
// This module contain data and methods to realise L0 server functions.

package teonet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/kirill-scherba/teonet-go/teocli/teocli"
	"github.com/kirill-scherba/teonet-go/teolog/teolog"
	"github.com/kirill-scherba/teonet-go/trudp/trudp"
)

// Teonet L0 server module

// l0Conn is Module data structure
type l0Conn struct {
	teo     *Teonet            // Pointer to Teonet
	stat    *l0Stat            // Statistic
	allow   bool               // Allow L0 Server
	wsAllow bool               // Allow L0 websocket server
	tcpPort int                // TCP port (if 0 - not allowed TCP)
	wsPort  int                // Websocket TCP port (if 0 - not allowed websocket)
	conn    net.Listener       // TCP listener connection
	ch      chan *packet       // Packet processing channel
	ma      map[string]*client // Clients address map
	mn      map[string]*client // Clients name map
	mux     sync.Mutex         // Maps mutex
	closed  bool               // Closet flag
}

// packet is Packet processing channels data structure
type packet struct {
	packet []byte
	client *client
}

// conn is an interface to make one parameter for tcp 'conn net.Conn' and
// trudp '*trudp.ChannelData' connection
type conn interface {
	Write([]byte) (int, error)
	Close() error
}

// client is clients data structure
type client struct {
	name string           // name
	addr string           // address (ip:port:ch)
	conn conn             // Connection tcp (net.Conn) or trudp (*trudp.ChannelData)
	stat clientStat       // Statistic
	cli  *teocli.TeoLNull // teocli connection to use readBuffer
}

// clientStat client statistic
type clientStat struct {
	send    int // send packes to client counter
	receive int // receive packes from client counter
	// sendRT    trudp.RealTimeSpeed // send packes to client real time counter
	// receiveRT trudp.RealTimeSpeed // receive packes to client real time counter
}

// l0New initialize l0 module
func (teo *Teonet) l0New() *l0Conn {
	l0 := &l0Conn{
		teo:     teo,                 // Pointer to Teonet
		allow:   teo.param.L0allow,   // Allow udp and tcp server(if tcp port > 0)
		tcpPort: teo.param.L0tcpPort, // Allow tcp server(if allow is true)
		wsAllow: teo.param.L0wsAllow, // Allow websocket server(if websocket port > 0)
		wsPort:  teo.param.L0wsPort,  // Allow websocket server(if wsAllow is true)
	}
	l0.stat = l0.l0StatNew()
	if l0.allow || l0.wsAllow {
		// Start L0 pocessing
		l0.ma = make(map[string]*client)
		l0.mn = make(map[string]*client)
		l0.process()
		// Start udp l0 server
		if l0.allow {
			teolog.Connect(MODULE, "l0 server start listen udp port:", l0.teo.param.Port)
		}
		// Start tcp l0 server
		if l0.tcpPort > 0 {
			l0.tcpServer(&l0.tcpPort)
			teo.param.L0tcpPort = l0.tcpPort
		}
		// Start websocket l0 server
		if l0.wsAllow && l0.wsPort > 0 {
			go l0.wsServe(l0.wsPort)
		}
	}
	return l0
}

// destroy destroys l0 module
func (l0 *l0Conn) destroy() {
	if l0.allow {
		l0.closeAll()
		teolog.Connect(MODULE, "l0 server stop listen udp port:", l0.teo.param.Port)
		if l0.conn != nil {
			l0.conn.Close()
			l0.conn = nil
		}
		if !l0.closed {
			close(l0.ch)
			l0.closed = true
		}
	}
}

// add adds new client
func (l0 *l0Conn) add(client *client) {
	teolog.Debugf(MODULE, "new client %s (%s) connected\n", client.name, client.addr)
	l0.closeName(client.name)
	l0.mux.Lock()
	l0.ma[client.addr] = client
	l0.mn[client.name] = client
	l0.mux.Unlock()
	l0.stat.updated()
}

// close closes(disconnect) connected client
func (l0 *l0Conn) close(client *client) (err error) {
	if client == nil {
		err = errors.New("client is nil")
		teolog.Error(MODULE, err.Error())
		return
	}
	teolog.Debugf(MODULE, "client %s (%s) disconnected\n", client.name, client.addr)
	if client.conn != nil {
		client.conn.Close()
		client.conn = nil
	}
	l0.mux.Lock()
	delete(l0.ma, client.addr)
	delete(l0.mn, client.name)
	l0.mux.Unlock()
	l0.stat.updated()
	return
}

// closeAddr closes(disconnect) connected client by address
func (l0 *l0Conn) closeAddr(addr string) (done bool) {
	if client, ok := l0.findAddr(addr); ok {
		l0.close(client)
		done = true
	}
	return
}

// closeName closes(disconnect) connected client by name
func (l0 *l0Conn) closeName(name string) (done bool) {
	if client, ok := l0.findName(name); ok {
		l0.close(client)
		done = true
	}
	return
}

// closeAll close(disconnect) all connected clients
func (l0 *l0Conn) closeAll() {
	for _, client := range l0.mn {
		l0.close(client)
	}
}

// tcpServer TCP L0 server
func (l0 *l0Conn) tcpServer(port *int) {

	const (
		network  = "tcp"
		hostName = ""
	)

	var err error

	// Start tcp server
	for {
		l0.conn, err = net.Listen(network, hostName+":"+strconv.Itoa(*port))
		if err == nil {
			break
		}
		*port++
		fmt.Println("the", *port-1, "is busy, try next port:", *port)
	}
	// If input function parameter port was 0 than get it from connection for
	// future use
	if *port == 0 {
		*port = l0.conn.Addr().(*net.TCPAddr).Port
	}
	teolog.Connect(MODULE, "l0 server start listen tcp port:", *port)

	// Listen for an incoming connection
	go func(port int) {
		for {
			conn, err := l0.conn.Accept()
			if err != nil {
				teolog.Debug(MODULE, "stop accepting: ", err.Error())
				//os.Exit(1)
				break
			}
			// Handle connections in a new goroutine.
			go l0.handleConnection(conn)
		}
		teolog.Connect(MODULE, "l0 server stop listen tcp port:", port)
	}(*port)
}

// findAddr finds client in clients map by address
func (l0 *l0Conn) findAddr(addr string) (client *client, ok bool) {
	l0.mux.Lock()
	client, ok = l0.ma[addr]
	l0.mux.Unlock()
	return
}

// findName finds client in clients map by name
func (l0 *l0Conn) findName(name string) (client *client, ok bool) {
	l0.mux.Lock()
	client, ok = l0.mn[name]
	l0.mux.Unlock()
	return
}

// exists return true if client exists in clients map
func (l0 *l0Conn) exists(client *client) (ok bool) {
	for _, cli := range l0.mn {
		if cli == client {
			ok = true
			return
		}
	}
	return
}

// toprocess send packet to packet processing
func (l0 *l0Conn) toprocess(p []byte, cli *teocli.TeoLNull, addr string, conn conn) {
	l0.ch <- &packet{packet: p, client: &client{cli: cli, addr: addr, conn: conn}}
	return
}

// check checks that received trudp packet is l0 packet and process it so.
// Return satatus 1 if not processed(if it is not teocli packet), or 0 if
// processed and send, or -1 if part of packet received and we wait next
// subpacket
func (l0 *l0Conn) check(tcd *trudp.ChannelData, packet []byte) (p []byte, status int) {
	if !l0.allow {
		return nil, 1
	}
	// Find this trudp key (connection) in cliens table and get cli, or create
	// new if not fount
	key := tcd.GetKey()
	var cli *teocli.TeoLNull
	client, ok := l0.findAddr(key)
	if !ok {
		cli, _ = teocli.Init(false) // create new client
	} else {
		cli = client.cli
	}
	return l0.packetCheck(cli, key, tcd, packet)
}

// packetCheck check that received tcp packet is l0 packet and process it so.
// Return satatus 1 if not processed(if it is not teocli packet), or 0 if
// processed and send, or -1 if part of packet received and we wait next subpacket
func (l0 *l0Conn) packetCheck(cli *teocli.TeoLNull, addr string, conn conn, data []byte) (p []byte, status int) {
check:
	p, status = cli.PacketCheck(data)
	switch status {
	case 0:
		l0.toprocess(p, cli, addr, conn)
		data = nil
		goto check // check next packet in read buffer
	case -1:
		if data != nil {
			teolog.Debugf(MODULE, "packet not received yet (got part of packet)\n")
		}
	case 1:
		teolog.Debugf(MODULE, "wrong packet received (drop it): %d, data: %v\n", len(p), p)
	}
	return
}

// Handle TCP connection
func (l0 *l0Conn) handleConnection(conn net.Conn) {
	teolog.Connectf(MODULE, "l0 server tcp client %v connected...", conn.RemoteAddr())
	cli, _ := teocli.Init(true)
	b := make([]byte, 2048)
	for {
		n, err := conn.Read(b)
		if err != nil {
			break
		}
		teolog.Debugf(MODULE, "got %d bytes data from tcp clien: %v\n",
			n, conn.RemoteAddr().String())
		l0.packetCheck(cli, conn.RemoteAddr().String(), conn, b[:n])
	}
	teolog.Connectf(MODULE, "l0 server tcp client %v disconnected...", conn.RemoteAddr())
	if !l0.closeAddr(conn.RemoteAddr().String()) {
		conn.Close()
	}
}

// Process received packets
func (l0 *l0Conn) process() {
	teolog.Debugf(MODULE, "l0 packet process started\n")
	l0.ch = make(chan *packet)
	l0.teo.wg.Add(1)
	go func() {
		for pac := range l0.ch {
			teolog.Debugf(MODULE,
				"valid packet received from client %s, length: %d\n",
				pac.client.addr, len(pac.packet),
			)
			p := pac.client.cli.PacketNew(pac.packet)

			// Find address in clients map and add if it absent, or send packet to
			// peer if client already exsists
			client, ok := l0.findAddr(pac.client.addr)
			if !ok {
				// First message from client should contain login command. Loging
				// command parameters: cmd = 0 and name = "" and data = 'client name'.
				// When we got valid data in l0 login we add clien to cliens map and all
				// next commands form this clients resend to peers with this name (name
				// from login command data)
				if d := p.Data(); p.Command() == 0 && p.Name() == "" {
					pac.client.name = string(d[:len(d)-1])
					l0.stat.receive(pac.client, d) //p.Data())
					l0.add(pac.client)

					// \TODO: Send to auth
					TEO_AUTH := "teo-auth"
					fmt.Printf("Send to auth: %s, data: %v\n", TEO_AUTH, d)
					l0.teo.SendTo(TEO_AUTH, CmdUser, d)

					// (kev->kc, TEO_AUTH, CMD_USER,
					// 				kld->name, kld->name_length);

				} else {
					teolog.Debugf(MODULE,
						"incorrect login packet received from client %s, disconnect...\n",
						pac.client.addr)
					fmt.Printf("cmd: %d, to: %s, data: %v\n", p.Command(), p.Name(), p.Data())
					// Send http answer to tcp request
					switch pac.client.conn.(type) {
					case net.Conn:
						pac.client.conn.Write([]byte("HTTP/1.1 200 OK\n" +
							"Content-Type: text/html\n\n" +
							"<html><body>Hello!</body></html>\n"))
					}
					pac.client.conn.Close()
				}
				continue
			}
			// Send command to peer for exising client
			l0.stat.receive(client, p.Data())
			l0.sendToPeer(p.Name(), client.name, p.Command(), p.Data())
		}
		l0.closeAll()
		teolog.Debugf(MODULE, "l0 packet process stopped\n")
		l0.teo.wg.Done()
	}()
}

// packetCreate creates packet data for sendToPeer and sendToL0
func (l0 *l0Conn) packetCreate(client string, cmd byte, data []byte) []byte {
	buf := new(bytes.Buffer)
	le := binary.LittleEndian
	binary.Write(buf, le, cmd)                       // Command
	binary.Write(buf, le, byte(len(client)+1))       // Client name length (include trailing zero)
	binary.Write(buf, le, uint16(len(data)))         // Packet data length
	binary.Write(buf, le, append([]byte(client), 0)) // Client name (include trailing zero)
	binary.Write(buf, le, []byte(data))              // Packet data
	return buf.Bytes()
}

// packetParse parse command data
func (l0 *l0Conn) packetParse(d []byte) (name string, cmd byte, data []byte) {
	buf := bytes.NewReader(d)
	le := binary.LittleEndian
	var fromLen byte
	var dataLen uint16
	binary.Read(buf, le, &cmd)     // Command
	binary.Read(buf, le, &fromLen) // From client name length (include trailing zero)
	binary.Read(buf, le, &dataLen) // Packet data length
	name = func() string {
		nameBuf := make([]byte, fromLen)
		binary.Read(buf, le, nameBuf)
		return string(nameBuf[:len(nameBuf)-1])
	}()
	data = func() []byte {
		d := make([]byte, dataLen)
		binary.Read(buf, le, d)
		return d
	}()
	return
}

// sendToPeer (send from L0 server to peer) send packet received from client to peer
func (l0 *l0Conn) sendToPeer(peer string, client string, cmd byte, data []byte) {
	teolog.Debugf(MODULE,
		"send cmd: %d, %d bytes data packet to peer %s, from client: %s",
		cmd, len(data), peer, client,
	)
	l0.teo.SendTo(peer, CmdL0, l0.packetCreate(client, cmd, data)) // Send to peer
}

// sendToL0 (send from peer to L0 server) send packet from peer to client
func (l0 *l0Conn) sendToL0(peer string, client string, cmd byte, data []byte) {
	teolog.Debugf(MODULE,
		"send cmd: %d, %d bytes data packet to l0 %s, from client: %s",
		cmd, len(data), peer, client,
	)
	l0.teo.SendTo(peer, CmdL0To, l0.packetCreate(client, cmd, data)) // Send to L0
}

// sendTo send command from this host L0 server to L0 client connected to this server
func (l0 *l0Conn) sendTo(from string, name string, cmd byte, data []byte) {

	// Get client data from name map
	l0.mux.Lock()
	client, ok := l0.mn[name]
	l0.mux.Unlock()
	if !ok {
		teolog.Debugf(MODULE, "can't find client '%s' in clients map\n", name)
		return
	}

	teolog.Debugf(MODULE, "got cmd: %d, %d bytes data packet from peer %s, to client: %s\n",
		cmd, len(data), from, name)

	packet, err := client.cli.PacketCreate(uint8(cmd), from, data)
	if err != nil {
		teolog.Error(MODULE, err.Error())
		return
	}

	teolog.Debugf(MODULE, "send cmd: %d, %d bytes data packet, to %s l0 client: %s\n",
		cmd, len(data), l0.network(client), client.name)

	l0.stat.send(client, packet)
	client.conn.Write(packet)
}

// network return network type of conn: 'tcp' or' trudp' (in string)
func (l0 *l0Conn) network(client *client) (network string) {
	switch client.conn.(type) {
	case net.Conn:
		network = "tcp"
	case *trudp.ChannelData:
		network = "trudp"
	case *wsConn:
		network = "ws"
	}
	return
}

// cmdL0 parse cmd got from L0 server with packet from L0 client
func (l0 *l0Conn) cmdL0(rec *receiveData) (processed bool, err error) {
	l0.teo.com.log(rec.rd, "CMD_L0 command")

	// Create packet
	rd, err := l0.teo.packetCreateNew(l0.packetParse(rec.rd.Data())).Parse()
	if err != nil {
		err = errors.New("can't parse packet from l0")
		fmt.Println(err.Error())
		return
	}

	// Sel L0 flag and addresses
	rd.setL0(func() (addr string, port int) {
		if port = l0.teo.param.Port; rec.tcd != nil {
			addr, port = rec.tcd.GetAddr().IP.String(), rec.tcd.GetAddr().Port
		}
		return
	}())

	// Process command
	processed = l0.teo.com.process(&receiveData{rd, rec.tcd})
	return
}

// cmdL0To parse cmd got from peer to L0 client
func (l0 *l0Conn) cmdL0To(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_TO command")

	if !l0.allow {
		teolog.Debugf(MODULE, "can't process this command because I'm not L0 server\n")
		return
	}

	// Parse command data
	name, cmd, data := l0.packetParse(rec.rd.Data())
	l0.sendTo(rec.rd.From(), name, cmd, data)
}

// cmdL0Auth Check l0 client answer from authentication application
func (l0 *l0Conn) cmdL0Auth(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_AUTH command")

	type authJSON struct {
		UserId      string   `json:"userId"`
		ClientId    string   `json:"clientId"`
		Username    string   `json:"username"`
		AccessToken string   `json:"accessToken"`
		User        string   `json:"user"`
		Networks    []string `json:"networks"`
	}
	type authToJSON struct {
		Name     string   `json:"name"`
		Networks []string `json:"networks"`
	}
	var j authJSON
	if err := json.Unmarshal(rec.rd.Data()[:rec.rd.DataLen()-1], &j); err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("d: %s, AccessToken: %s\n", string(rec.rd.Data()), j.AccessToken)
	// { \"name\": \"%s\", \"networks\": %s }
	// \TODO set correct name
	var jt = authToJSON{Name: j.AccessToken, Networks: j.Networks}
	jdata, _ := json.Marshal(jt)
	l0.sendTo(rec.rd.From(), j.AccessToken, rec.rd.Cmd(), jdata)
}

// cmdL0ClientsNumber parse cmd 'got clients number' and send answer with number of clients
func (l0 *l0Conn) cmdL0ClientsNumber(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_CLIENTS_N command")
	if !l0.allow {
		teolog.Error(MODULE, "can't process this command because I'm not L0 server\n")
		return
	}
	var err error
	var data []byte
	type numClientsJSON struct {
		NumClients uint32 `json:"numClients"`
	}
	numClients := uint32(len(l0.mn))
	if l0.teo.com.isJSONRequest(rec.rd.Data()) {
		data, err = l0.teo.com.structToJSON(numClientsJSON{numClients})
	} else {
		data = make([]byte, 4)
		binary.LittleEndian.PutUint32(data, numClients)
	}
	if err != nil {
		teolog.Error(MODULE, err)
		return
	}
	l0.teo.SendAnswer(rec, CmdL0ClientsNumAnswer, data)
}

// cmdL0Clients parse cmd 'got clients list' and send answer with list of clients
func (l0 *l0Conn) cmdL0Clients(rec *receiveData) {
	l0.teo.com.log(rec.rd, "CMD_L0_CLIENTS command")
	if !l0.allow {
		teolog.Debugf(MODULE, "can't process this command because I'm not L0 server\n")
		return
	}
	// \TODO: write code ...
}

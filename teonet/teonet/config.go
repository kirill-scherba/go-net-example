// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet config module.
//
// Parse and save, restore teonet parameters from configuration file.

package teonet

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kirill-scherba/teonet-go/services/teoapi"
)

// Parameters Teonet parameters
type Parameters struct {
	Name             string `json:"name"`             // this host client name
	Port             int    `json:"port"`             // local port
	RAddr            string `json:"r-addr"`           // remote host address
	RPort            int    `json:"r-port"`           // remote host port
	RChan            int    `json:"r-ch"`             // remote host channel(for TRUdp only)
	Network          string `json:"network"`          // teonet network name
	Loglevel         string `json:"log-level"`        // show log messages level
	LogFilter        string `json:"log-filter"`       // log messages filter
	LogToSyslogF     bool   `json:"log-to-syslog"`    // forbid hotkeys menu
	ForbidHotkeysF   bool   `json:"forbid-hotkeys"`   // forbid hotkeys menu
	ShowTrudpStatF   bool   `json:"show-trudp"`       // show trudp statistic
	ShowPeersStatF   bool   `json:"show-peers"`       // show peers table
	ShowClientsStatF bool   `json:"show-clients"`     // show clients table
	ShowParametersF  bool   `json:"show-params"`      // show parameters
	SaveConfigF      bool   `json:"save-config"`      // save current parameters to config
	ShowHelpF        bool   `json:"show-help"`        // show usage
	IPv6Allow        bool   `json:"ipv6-allow"`       // allow IPv6 support (not supported in Teonet-C)
	DisallowEncrypt  bool   `json:"disallow-encrypt"` // disable teonet packets encryption
	CtrlcF           bool   `json:"ctrlc"`            // use Ctrl+C to gracefully exit from application
	L0allow          bool   `json:"l0-allow"`         // allow l0 server
	L0tcpPort        int    `json:"l0-tcp-port"`      // l0 Server tcp port number (default 9000)
	L0wsAllow        bool   `json:"l0-ws-allow"`      // allow l0 WebSocket server
	L0wsPort         int    `json:"l0-ws-port"`       // l0 Server websocket tcp port number (default 9080)
}

// Params read Teonet parameters from configuration file and parse application
// flars and arguments
func Params(ii ...interface{}) (param *Parameters) {
	// Teonet parameters and config
	param = CreateParameters()
	param.ReadConfig()

	// This application Usage message
	flag.Usage = func() {
		fmt.Printf("usage: %s [OPTIONS] host_name\n\noptions:\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Teonet flags
	flag.IntVar(&param.Port, "p", param.Port, "local host port")
	flag.StringVar(&param.Network, "n", param.Network, "teonet network name")
	flag.StringVar(&param.RAddr, "a", param.RAddr, "remote host address to connect to remote host")
	flag.IntVar(&param.RChan, "c", param.RChan, "remote host channel (to connect to remote host TRUDP channel)")
	flag.IntVar(&param.RPort, "r", param.RPort, "remote host port (to connect to remote host)")
	flag.StringVar(&param.Loglevel, "log-level", param.Loglevel, "show log messages level")
	flag.StringVar(&param.LogFilter, "log-filter", param.LogFilter, "set log messages filter")
	flag.BoolVar(&param.LogToSyslogF, "log-to-syslog", param.LogToSyslogF, "save log messages to syslog")
	flag.BoolVar(&param.ForbidHotkeysF, "forbid-hotkeys", param.ForbidHotkeysF, "forbid hotkeys menu")
	flag.BoolVar(&param.ShowTrudpStatF, "show-trudp", param.ShowTrudpStatF, "show trudp statistic")
	flag.BoolVar(&param.ShowPeersStatF, "show-peers", param.ShowPeersStatF, "show peers table")
	flag.BoolVar(&param.ShowClientsStatF, "show-clients", param.ShowClientsStatF, "show clients table")
	flag.BoolVar(&param.ShowParametersF, "show-params", param.ShowParametersF, "show application parameters")
	flag.BoolVar(&param.SaveConfigF, "save-config", param.SaveConfigF, "save parameters to configuration file")
	flag.BoolVar(&param.ShowHelpF, "h", false, "show this help message")
	flag.BoolVar(&param.IPv6Allow, "ipv6", param.IPv6Allow, "allow ipv6 connection")
	flag.BoolVar(&param.L0allow, "l0-allow", param.L0allow, "allow l0 server")
	flag.IntVar(&param.L0tcpPort, "l0-tcp-port", param.L0tcpPort, "l0 server tcp port number")
	flag.BoolVar(&param.L0wsAllow, "l0-ws-allow", param.L0wsAllow, "allow l0 websocket server")
	flag.IntVar(&param.L0wsPort, "l0-ws-port", param.L0wsPort, "l0 websocket server tcp port number")
	flag.BoolVar(&param.DisallowEncrypt, "disable-encrypt", param.DisallowEncrypt, "disable teonet packets encryption")

	// Teonet api flags
	var showAPI bool
	var trapi *teoapi.Teoapi
	if len(ii) > 0 {
		switch v := ii[0].(type) {
		case *teoapi.Teoapi:
			flag.BoolVar(&showAPI, "api", param.DisallowEncrypt, "show teonet application api")
			trapi = v
		}
	}

	// Parse flags
	flag.Parse()

	// Show teonet api if api flag is set
	if showAPI {
		fmt.Println(trapi.Sprint())
		os.Exit(0)
	}

	// Teonet Arguments
	args := flag.Args()
	if len(args) > 0 && len(flag.Arg(0)) > 0 {
		param.Name = flag.Arg(0)
	}

	// Check requered
	if param.ShowHelpF || len(param.Name) == 0 {
		if len(args) == 0 {
			fmt.Printf("argument host_name not defined\n")
		}
		flag.Usage()
		os.Exit(0)
	} else {
		param.WriteConfig()
	}

	return
}

// CreateParameters create new Teonet parameters with default values
func CreateParameters() (param *Parameters) {
	param = new(Parameters)
	param.setDefault()
	return
}

// ReadConfig read teonet configuration file
func (param *Parameters) ReadConfig() {
	param.read("teonet.conf")
}

// WriteConfig write teonet configuration file
func (param *Parameters) WriteConfig() {
	param.write("teonet.conf.out")
	if param.SaveConfigF {
		param.SaveConfigF = false
		param.write("teonet.conf")
	}
}

// configDir return configuration files folder
func (param *Parameters) configDir() string {
	home := os.Getenv("HOME")
	return home + "/.config/teonet/"
}

// setDefault set default teonet parameters
func (param *Parameters) setDefault() {
	param.Network = "local"
	param.RAddr = "localhost"
	param.Loglevel = "DEBUG"
	param.CtrlcF = true
	param.ShowParametersF = true
}

// read read teonet parameters from selected configuration file
func (param *Parameters) read(fileName string) {
	confDir := param.configDir()
	f, err := os.Open(confDir + fileName)
	if err != nil {
		return
	}
	data := make([]byte, 1024)
	n, err := f.Read(data)
	if err != nil {
		return
	}
	json.Unmarshal(data[:n], param)
}

// println print teonet parameters
func (param *Parameters) println() {
	if param.ShowParametersF {
		j, _ := json.MarshalIndent(param, "", " ")
		fmt.Println("Teonet parameters:\n" + string(j))
	}
}

// write write teonet parameters to selected configuration file
func (param *Parameters) write(fileName string) {
	confDir := param.configDir()
	if err := os.MkdirAll(confDir, os.ModePerm); err != nil {
		panic(err)
	}
	f, _ := os.Create(confDir + fileName)
	j, _ := json.MarshalIndent(param, "", " ")
	param.println()
	f.Write(j)
}

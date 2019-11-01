// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teolog is the Teonet loger package
//
package teolog

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"

	"github.com/kirill-scherba/teonet-go/teokeys/teokeys"
)

// Type of log messages
const (
	NONE = iota
	ERROR
	CONNECT
	MESSAGE
	DEBUG
	DEBUGv
	DEBUGvv
)

const (
	strNONE    = "NONE"
	strERROR   = "ERROR"
	strCONNECT = "CONNECT"
	strMESSAGE = "MESSAGE"
	strDEBUG   = "DEBUG"
	strDEBUGv  = "DEBUGv"
	strDEBUGvv = "DEBUGvv"
	strUNKNOWN = "UNKNOWN"
)

type logParam struct {
	log      *log.Logger
	level    int
	filter   string
	toSyslog bool
}

var param logParam

// None show NONE log string
func None(p ...interface{}) {
	logOutput(2, NONE, p...)
}

// Nonef show NONE log formatted string
func Nonef(module string, format string, p ...interface{}) {
	logOutputf(2, NONE, module, format, p...)
}

// Connect show CONNECT log string
func Connect(p ...interface{}) {
	logOutput(2, CONNECT, p...)
}

// Connectf show CONNECT log formatted string
func Connectf(module string, format string, p ...interface{}) {
	logOutputf(2, CONNECT, module, format, p...)
}

// Error show ERROR log string
func Error(p ...interface{}) {
	logOutput(2, ERROR, p...)
}

// Errorf show ERROR log formatted string
func Errorf(module string, format string, p ...interface{}) {
	logOutputf(2, ERROR, module, format, p...)
}

// Errorfd show ERROR formatted string (with calldepth)
func Errorfd(calldepth int, module string, format string, p ...interface{}) {
	logOutputf(calldepth+2, DEBUGv, module, format, p...)
}

// Message show MESSAGE log string
func Message(p ...interface{}) {
	logOutput(2, MESSAGE, p...)
}

// Messagef show MESSAGE log formatted string
func Messagef(module string, format string, p ...interface{}) {
	logOutputf(2, MESSAGE, module, format, p...)
}

// Debug show DEBUG log string
func Debug(p ...interface{}) {
	logOutput(2, DEBUG, p...)
}

// Debugf show DEBUG log formatted string
func Debugf(module string, format string, p ...interface{}) {
	logOutputf(2, DEBUG, module, format, p...)
}

// DebugV show DEBUGv log string
func DebugV(p ...interface{}) {
	logOutput(2, DEBUGv, p...)
}

// DebugVf show DEBUGv formatted string
func DebugVf(module string, format string, p ...interface{}) {
	logOutputf(2, DEBUGv, module, format, p...)
}

// DebugVfd show DEBUGv formatted string (with calldepth)
func DebugVfd(calldepth int, module string, format string, p ...interface{}) {
	logOutputf(calldepth+2, DEBUGv, module, format, p...)
}

// DebugVv show DEBUGvv log string
func DebugVv(p ...interface{}) {
	logOutput(2, DEBUGvv, p...)
}

// DebugVvf show DEBUGvv log formatted string
func DebugVvf(module string, format string, p ...interface{}) {
	logOutputf(2, DEBUGvv, module, format, p...)
}

// Log show log string
func Log(level int, p ...interface{}) {
	logOutput(2, level, p...)
}

// Logf show log formatted string
func Logf(level int, module string, format string, p ...interface{}) {
	logOutputf(2, level, module, format, p...)
}

// Log show log string
func logOutput(calldepth int, level int, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 1+len(p))
		pp = append(append(pp, LoglevelStringColor(level)), p...)
		msg := fmt.Sprintln(pp...)
		if checkFilter(msg) {
			param.log.Output(calldepth+1, removeTEsc(msg, param.toSyslog))
		}
	}
}

// Log show log formatted string
func logOutputf(calldepth int, level int, module string, format string, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 2+len(p))
		pp = append(append(pp, LoglevelStringColor(level), module), p...)
		msg := fmt.Sprintf("%s %s "+format, pp...)
		if checkFilter(msg) {
			param.log.Output(calldepth+1, removeTEsc(msg, param.toSyslog))
		}
	}
}

// Init initial module and sets log level
// Avalable level values: NONE, CONNECT, ERROR, MESSAGE, DEBUG, DEBUGv, DEBUGvv
func Init(level interface{}, flags int, filter string, toSyslogF bool, syslogPrefix string) {

	if toSyslogF {
		param.log, _ = syslog.NewLogger(syslog.LOG_DEBUG, flags)
		param.log.SetPrefix(syslogPrefix + ": ")
		param.toSyslog = toSyslogF
	} else {
		param.log = log.New(os.Stdout, teokeys.ANSIDarkGrey, flags)
	}

	// Set log flags
	if flags == 0 {
		flags = log.LstdFlags
	}
	log.SetFlags(flags)

	// Set log level
	SetLoglevel(level)

	// Set log filter
	SetFilter(filter)

	// Show log level and log filter
	fmt.Println("log level:", LoglevelString(param.level))
	if param.filter != "" {
		fmt.Println("log filter:", param.filter)
	}
	fmt.Println()
}

// SetLoglevel sets log level in int or string format
func SetLoglevel(level interface{}) {
	switch l := level.(type) {
	case int:
		param.level = l
	case string:
		param.level = Loglevel(l)
	default:
		param.level = DEBUG
	}
}

// Loglevel return log level in int format
func Loglevel(lstr string) (level int) {
	switch lstr {
	case strNONE:
		level = NONE
	case strERROR:
		level = ERROR
	case strCONNECT:
		level = CONNECT
	case strMESSAGE:
		level = MESSAGE
	case strDEBUG:
		level = DEBUG
	case strDEBUGv:
		level = DEBUGv
	case strDEBUGvv:
		level = DEBUGvv
	default:
		level = DEBUG
	}
	return
}

// LoglevelString return trudp log level in string format
func LoglevelString(level int) (strLoglevel string) {
	switch level {
	case NONE:
		strLoglevel = strNONE
	case CONNECT:
		strLoglevel = strCONNECT
	case ERROR:
		strLoglevel = strERROR
	case MESSAGE:
		strLoglevel = strMESSAGE
	case DEBUG:
		strLoglevel = strDEBUG
	case DEBUGv:
		strLoglevel = strDEBUGv
	case DEBUGvv:
		strLoglevel = strDEBUGvv
	default:
		strLoglevel = strUNKNOWN
	}
	return
}

// LoglevelStringColor return trudp log level in string format with ansi collor
func LoglevelStringColor(level int) (strLoglevel string) {
	switch level {
	case NONE:
		strLoglevel = teokeys.Color(teokeys.ANSIGrey, strNONE)
	case CONNECT:
		strLoglevel = teokeys.Color(teokeys.ANSIGreen, strCONNECT)
	case ERROR:
		strLoglevel = teokeys.Color(teokeys.ANSIRed, strERROR)
	case MESSAGE:
		strLoglevel = teokeys.Color(teokeys.ANSICyan, strMESSAGE)
	case DEBUG:
		strLoglevel = teokeys.Color(teokeys.ANSIBlue, strDEBUG)
	case DEBUGv:
		strLoglevel = teokeys.Color(teokeys.ANSIBrown, strDEBUGv)
	case DEBUGvv:
		strLoglevel = teokeys.Color(teokeys.ANSIMagenta, strDEBUGvv)
	default:
		strLoglevel = strUNKNOWN
	}
	return
}

// Filter return logger filter
func Filter() string {
	return param.filter
}

// SetFilter sets logger filter
func SetFilter(filter string) {
	param.filter = filter
}

// checkFilter parse log message strings and return true if filter allow (show message)
func checkFilter(message string) bool {
	if param.filter == "" {
		return true
	}
	return strings.Contains(message, param.filter)
}

// removeTEsc removes terminal escape text formating in input string and return
// new string without this format
func removeTEsc(str string, toSyslog bool) (retStr string) {
	if !toSyslog {
		retStr = str
		return
	}
	var skipEsc bool
	l := len(str)
	for i := 0; i < l; i++ {
		switch {
		case skipEsc && str[i] == 'm':
			skipEsc = false
		case !skipEsc && str[i] == '\033':
			skipEsc = true
		case !skipEsc:
			retStr += string(str[i])
		}
	}
	return
}

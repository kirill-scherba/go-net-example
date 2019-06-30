package trudp

import (
	"fmt"
	"log"
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

// log shows log message in terminal
func (trudp *TRUDP) log(level int, p ...interface{}) {
	if level <= trudp.logLevel {
		if trudp.logLog {
			log.Println(p...)
		} else {
			fmt.Println(p...)
		}
	}
}

// LogLevel sets TRUDP log level
// Avalable level values: NONE, CONNECT, MESSAGE, DEBUG, DEBUGv, DEBUGvv
func (trudp *TRUDP) LogLevel(level interface{}, logLog bool, flag int) {

	// Set log type
	trudp.logLog = logLog
	if logLog {
		if flag == 0 {
			flag = log.LstdFlags
		}
		log.SetFlags(flag)
	}

	// Set log level
	switch l := level.(type) {
	case int:
		trudp.logLevel = l
	case string:
		switch l {
		case strNONE:
			trudp.logLevel = NONE
		case strERROR:
			trudp.logLevel = ERROR
		case strCONNECT:
			trudp.logLevel = CONNECT
		case strMESSAGE:
			trudp.logLevel = MESSAGE
		case strDEBUG:
			trudp.logLevel = DEBUG
		case strDEBUGv:
			trudp.logLevel = DEBUGv
		case strDEBUGvv:
			trudp.logLevel = DEBUGvv
		default:
			trudp.logLevel = DEBUG
		}
	default:
		trudp.logLevel = DEBUG
	}

	// Show log level
	fmt.Println("show time in log:", trudp.logLog)
	fmt.Println("log level:", trudp.LogLevelString())
}

// LogLevelString return trudp log level in string format
func (trudp *TRUDP) LogLevelString() (strLogLevel string) {

	switch trudp.logLevel {
	case NONE:
		strLogLevel = strNONE
	case CONNECT:
		strLogLevel = strCONNECT
	case ERROR:
		strLogLevel = strERROR
	case MESSAGE:
		strLogLevel = strMESSAGE
	case DEBUG:
		strLogLevel = strDEBUG
	case DEBUGv:
		strLogLevel = strDEBUGv
	case DEBUGvv:
		strLogLevel = strDEBUGvv
	default:
		strLogLevel = strUNKNOWN
	}

	return
}

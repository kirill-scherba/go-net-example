package trudp

import (
	"fmt"
	"log"
)

const (
	NONE = iota
	CONNECT
	MESSAGE
	DEBUG
	DEBUG_V
	DEBUG_VV
)

const (
	strNONE     = "NONE"
	strCONNECT  = "CONNECT"
	strMESSAGE  = "MESSAGE"
	strDEBUG    = "DEBUG"
	strDEBUG_V  = "DEBUG_V"
	strDEBUG_VV = "DEBUG_VV"
	strUNKNOWN  = "UNKNOWN"
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
// Avalable level values: NONE, CONNECT, MSSAGE, DEBUG, DEBUG_V, DEBUG_VV
func (trudp *TRUDP) LogLevel(level interface{}, log bool) {

	// Set log type
	trudp.logLog = log

	// Set log level
	switch l := level.(type) {
	case int:
		trudp.logLevel = l
	case string:
		switch l {
		case strNONE:
			trudp.logLevel = NONE
		case strCONNECT:
			trudp.logLevel = CONNECT
		case strMESSAGE:
			trudp.logLevel = MESSAGE
		case strDEBUG:
			trudp.logLevel = DEBUG
		case strDEBUG_V:
			trudp.logLevel = DEBUG_V
		case strDEBUG_VV:
			trudp.logLevel = DEBUG_VV
		default:
			trudp.logLevel = DEBUG
		}
	default:
		trudp.logLevel = DEBUG
	}

	// Show log level
	fmt.Println("log level:", trudp.LogLevelString())
	fmt.Println("show time in log:", trudp.logLog)
}

// LogLevelString reurn trudp log level in string format
func (trudp *TRUDP) LogLevelString() (strLogLevel string) {

	switch trudp.logLevel {
	case NONE:
		strLogLevel = strNONE
	case CONNECT:
		strLogLevel = strCONNECT
	case MESSAGE:
		strLogLevel = strMESSAGE
	case DEBUG:
		strLogLevel = strDEBUG
	case DEBUG_V:
		strLogLevel = strDEBUG_V
	case DEBUG_VV:
		strLogLevel = strDEBUG_VV
	default:
		strLogLevel = strUNKNOWN
	}

	return
}

package teolog

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

type logParam struct {
	level   int
	useLogF bool
}

var param logParam

// Log shows log message in terminal
func Log(level int /*module string,*/, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		//if module == "" {
		pp = make([]interface{}, 0, 1+len(p))
		pp = append(append(pp, LevelString(level)), p...)
		// } else {
		// 	pp = make([]interface{}, 0, 2+len(p))
		// 	pp = append(append(pp, LevelString(level), module), p...)
		// }
		if param.useLogF {
			log.Println(pp...)
		} else {
			fmt.Println(pp...)
		}
	}
}

// Level sets log level
// Avalable level values: NONE, CONNECT, MESSAGE, DEBUG, DEBUGv, DEBUGvv
func Level(level interface{}, useLogF bool, flag int) {

	// Set log type
	param.useLogF = useLogF
	if useLogF {
		if flag == 0 {
			flag = log.LstdFlags
		}
		log.SetFlags(flag)
	}

	// Set log level
	switch l := level.(type) {
	case int:
		param.level = l
	case string:
		switch l {
		case strNONE:
			param.level = NONE
		case strERROR:
			param.level = ERROR
		case strCONNECT:
			param.level = CONNECT
		case strMESSAGE:
			param.level = MESSAGE
		case strDEBUG:
			param.level = DEBUG
		case strDEBUGv:
			param.level = DEBUGv
		case strDEBUGvv:
			param.level = DEBUGvv
		default:
			param.level = DEBUG
		}
	default:
		param.level = DEBUG
	}

	// Show log level
	fmt.Println("show time in log:", param.useLogF)
	fmt.Println("log level:", LevelString(param.level))
	fmt.Println()
}

// LogLevelString return trudp log level in string format
func LevelString(level int) (strLogLevel string) {

	switch /*param.*/ level {
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

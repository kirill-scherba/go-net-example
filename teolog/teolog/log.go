package teolog

import (
	"fmt"
	"log"
	"os"

	"github.com/kirill-scherba/net-example-go/teokeys/teokeys"
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
	level int
	log   *log.Logger
}

var param logParam

// None show NONE log string
func None(a ...interface{}) {
	Log(NONE, a)
}

// Connect show CONNECT log string
func Connect(a ...interface{}) {
	Log(CONNECT, a)
}

// Error show ERROR log string
func Error(a ...interface{}) {
	Log(ERROR, a)
}

// Debug show DEBUG log string
func Debug(a ...interface{}) {
	Log(DEBUG, a)
}

// DebugV show DEBUGv string
func DebugV(a ...interface{}) {
	Log(DEBUGv, a)
}

// DebugVv show DEBUGvv log string
func DebugVv(a ...interface{}) {
	Log(DEBUGvv, a)
}

// Log show log string
func Log(level int, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 1+len(p))
		pp = append(append(pp, LevelStringColor(level)), p...)
		param.log.Output(2, fmt.Sprintln(pp...))
	}
}

// Init initial module and sets log level
// Avalable level values: NONE, CONNECT, ERROR, MESSAGE, DEBUG, DEBUGv, DEBUGvv
func Init(level interface{}, useLogF bool, flags int) {

	param.log = log.New(os.Stdout, "", flags)

	// Set log flags
	if flags == 0 {
		flags = log.LstdFlags
	}
	log.SetFlags(flags)

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
	fmt.Println("log level:", LevelString(param.level))
	fmt.Println()
}

// LevelString return trudp log level in string format
func LevelString(level int) (strLogLevel string) {

	switch level {
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

// LevelStringColor return trudp log level in string format with ansi collor
func LevelStringColor(level int) (strLogLevel string) {
	switch level {
	case NONE:
		strLogLevel = teokeys.Color(teokeys.ANSIGrey, strNONE)
	case CONNECT:
		strLogLevel = teokeys.Color(teokeys.ANSIGreen, strCONNECT)
	case ERROR:
		strLogLevel = teokeys.Color(teokeys.ANSIRed, strERROR)
	case MESSAGE:
		strLogLevel = teokeys.Color(teokeys.ANSICyan, strMESSAGE)
	case DEBUG:
		strLogLevel = teokeys.Color(teokeys.ANSIBlue, strDEBUG)
	case DEBUGv:
		strLogLevel = teokeys.Color(teokeys.ANSIBrown, strDEBUGv)
	case DEBUGvv:
		strLogLevel = teokeys.Color(teokeys.ANSIMagenta, strDEBUGvv)
	default:
		strLogLevel = strUNKNOWN
	}
	return
}

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
func None(p ...interface{}) {
	logOutput(2, NONE, p...)
}

// None show NONE log formatted string
func Nonef(module string, format string, p ...interface{}) {
	logOutputf(2, NONE, module, format, p...)
}

// Connect show CONNECT log string
func Connect(p ...interface{}) {
	logOutput(2, CONNECT, p...)
}

// Connect show CONNECT log formatted string
func Connectf(module string, format string, p ...interface{}) {
	logOutputf(2, CONNECT, module, format, p...)
}

// Error show ERROR log string
func Error(p ...interface{}) {
	logOutput(2, ERROR, p...)
}

// Error show ERROR log formatted string
func Errorf(module string, format string, p ...interface{}) {
	logOutputf(2, ERROR, module, format, p...)
}

// Debug show DEBUG log string
func Debug(p ...interface{}) {
	logOutput(2, DEBUG, p...)
}

// Debug show DEBUG log formatted string
func Debugf(module string, format string, p ...interface{}) {
	logOutputf(2, DEBUG, module, format, p...)
}

// DebugV show DEBUGv log string
func DebugV(p ...interface{}) {
	logOutput(2, DEBUGv, p...)
}

// DebugV show DEBUGv formatted string
func DebugVf(module string, format string, p ...interface{}) {
	logOutputf(2, DEBUGv, module, format, p...)
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

// Log show log formatted string
func Logf(level int, module string, format string, p ...interface{}) {
	logOutputf(2, level, module, format, p...)
}

// Log show log string
func logOutput(calldepth int, level int, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 1+len(p))
		pp = append(append(pp, LevelStringColor(level)), p...)
		param.log.Output(calldepth+1, fmt.Sprintln(pp...))
	}
}

// Log show log formatted string
func logOutputf(calldepth int, level int, module string, format string, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 2+len(p))
		pp = append(append(pp, LevelStringColor(level), module), p...)
		param.log.Output(calldepth+1, fmt.Sprintf("%s %s "+format, pp...))
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

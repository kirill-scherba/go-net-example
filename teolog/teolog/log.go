package teolog

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	level  int
	log    *log.Logger
	filter string
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
		pp = append(append(pp, LevelStringColor(level)), p...)
		msg := fmt.Sprintln(pp...)
		if checkFilter(msg) {
			param.log.Output(calldepth+1, msg)
		}
	}
}

// Log show log formatted string
func logOutputf(calldepth int, level int, module string, format string, p ...interface{}) {
	if level <= param.level {
		var pp []interface{}
		pp = make([]interface{}, 0, 2+len(p))
		pp = append(append(pp, LevelStringColor(level), module), p...)
		msg := fmt.Sprintf("%s %s "+format, pp...)
		if checkFilter(msg) {
			param.log.Output(calldepth+1, msg)
		}
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
		param.level = LogLevel(l)
	default:
		param.level = DEBUG
	}

	// Show log level
	fmt.Println("log level:", LevelString(param.level))
	fmt.Println()
}

// LogLevel return trudp log level in int format
func LogLevel(lstr string) (level int) {
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

// Get logger filter
func GetFilter() string {
	return param.filter
}

// Set logger filter
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

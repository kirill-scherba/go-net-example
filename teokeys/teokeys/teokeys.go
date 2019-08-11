package teokeys

// CLI hotkeys

//// CGO definition (don't delay or edit it):
//#include "rutil.h"
import "C"

// Ansi collors
const (
	ANSINone         = C._ANSI_NONE
	ANSICls          = C._ANSI_CLS
	ANSIBlack        = C._ANSI_BLACK
	ANSIRed          = C._ANSI_RED
	ANSIGreen        = C._ANSI_GREEN
	ANSIBrown        = C._ANSI_BROWN
	ANSIBlue         = C._ANSI_BLUE
	ANSIMagenta      = C._ANSI_MAGENTA
	ANSICyan         = C._ANSI_CYAN
	ANSIGrey         = C._ANSI_GREY
	ANSIDarkGrey     = C._ANSI_DARKGREY
	ANSILightGred    = C._ANSI_LIGHTRED
	ANSILightGreen   = C._ANSI_LIGHTGREEN
	ANSIYellow       = C._ANSI_YELLOW
	ANSILightBlue    = C._ANSI_LIGHTBLUE
	ANSILightMagenta = C._ANSI_LIGHTMAGENTA
	ANSILightCyan    = C._ANSI_LIGHTCYAN
	ANSIWhite        = C._ANSI_WHITE
)

// Color add ansi color to string
func Color(color string, text string) string {
	return color + text + ANSINone
}

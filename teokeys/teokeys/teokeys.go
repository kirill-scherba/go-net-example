// Copyright 2019 Kirill Scherba <kirill@scherba.ru> as Teokeys Author.
// All rights reserved. Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teokeys is the Teonet terminal hotkeys menu, cursor position and colors
// processing package
//
package teokeys

// CLI cursor position, colors and hotkeys processing module
// Created 2019-08-10 by Kirill Scherba <kirill@scherba.ru>

//// CGO definition (don't delay or edit it):
//#include "rutil.h"
import "C"

// Version is Teokeys package version
const Version = "3.0.0"

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
	ANSILightRed     = C._ANSI_LIGHTRED
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

// GetchNb non block getch() return 0 if no keys pressed
func GetchNb() int {
	return int(C.nb_getch())
}

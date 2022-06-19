package key

import (
	"unicode/utf8"

	"github.com/ardnew/embedit/seq/ascii"
)

// const Error = utf8.RuneError
const Error = utf8.MaxRune + 1

// Control key codes (C0, C1).
//
// From: https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	Null      = ascii.NUL // 0x00
	Alert     = ascii.BEL // 0x07
	Enter     = ascii.CR  // 0x0D
	Escape    = ascii.ESC // 0x1B
	FileSep   = ascii.FS  // 0x1C
	GroupSep  = ascii.GS  // 0x1D
	RecordSep = ascii.RS  // 0x1E
	UnitSep   = ascii.US  // 0x1F
	Space     = ascii.SP  // 0x20
	Backspace = ascii.DEL // 0x7F

	CtrlA = ascii.SOH // 0x01
	CtrlB = ascii.STX // 0x02
	CtrlC = ascii.ETX // 0x03
	CtrlD = ascii.EOT // 0x04
	CtrlE = ascii.ENQ // 0x05
	CtrlF = ascii.ACK // 0x06
	CtrlG = ascii.BEL // 0x07
	CtrlH = ascii.BS  // 0x08
	CtrlI = ascii.TAB // 0x09
	CtrlJ = ascii.LF  // 0x0A
	CtrlK = ascii.VT  // 0x0B
	CtrlL = ascii.FF  // 0x0C
	CtrlM = ascii.CR  // 0x0D
	CtrlN = ascii.SO  // 0x0E
	CtrlO = ascii.SI  // 0x0F
	CtrlP = ascii.DLE // 0x10
	CtrlQ = ascii.DC1 // 0x11
	CtrlR = ascii.DC2 // 0x12
	CtrlS = ascii.DC3 // 0x13
	CtrlT = ascii.DC4 // 0x14
	CtrlU = ascii.NAK // 0x15
	CtrlV = ascii.SYN // 0x16
	CtrlW = ascii.ETB // 0x17
	CtrlX = ascii.CAN // 0x18
	CtrlY = ascii.EM  // 0x19
	CtrlZ = ascii.SUB // 0x1A
)

// Control key sequences.
const (
	Unknown = 0xD800 /* UTF-16 surrogate area */ + iota
	Up
	Down
	Left
	Right
	AltLeft
	AltRight
	Home
	End
	Insert
	Delete
	PageUp
	PageDown
	DeleteWord
	DeleteLeading
	DeleteTrailing
	ClearScreen
	PasteStart
	PasteEnd
	F0
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
	F13
	F14
	F15
	F16
	F17
	F18
	F19
	F20
	surrogateMask = Unknown | 0x03FF
)

// Escape sequences.
var (
	CSI = []byte{Escape, '['}                     // Ctrl seq intro
	SOP = []byte{Escape, '[', '2', '0', '0', '~'} // Start of paste
	EOP = []byte{Escape, '[', '2', '0', '1', '~'} // End of paste
	CLS = []byte{Escape, '[', '2', 'J'}           // Clear screen
	XY0 = []byte{Escape, '[', 'H'}                // Set cursor X=0 Y=0
	KIL = []byte{Escape, '[', 'K'}                // Clear line right
	DEL = []byte{' ', Escape, '[', 'D'}           // Delete next rune
)

// IsPrintable returns true iff key is a visible, non-whitespace key.
func IsPrintable(key rune) bool {
	return key >= Space && (key < Unknown || surrogateMask < key)
}

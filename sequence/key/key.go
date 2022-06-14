package key

import (
	"unicode/utf8"
)

// const Error = utf8.RuneError
const Error = utf8.MaxRune + 1

// ASCII key constants
const (
	Null = iota
	CtrlA
	CtrlB
	CtrlC
	CtrlD
	CtrlE
	CtrlF
	CtrlG
	CtrlH
	CtrlI
	CtrlJ
	CtrlK
	CtrlL
	CtrlM
	CtrlN
	CtrlO
	CtrlP
	CtrlQ
	CtrlR
	CtrlS
	CtrlT
	CtrlU
	CtrlV
	CtrlW
	CtrlX
	CtrlY
	CtrlZ
	Escape
	_
	_
	_
	_
	Space
	Backspace = 0x7F
)

// ANSI control key constants
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
	DeleteWord
	DeleteLine
	ClearScreen
	PasteStart
	PasteEnd

	surrogateMask = Unknown | 0x03FF
)

// ANSI control sequences
var (
	CSI = []byte{Escape, '['}
	SOP = []byte{Escape, '[', '2', '0', '0', '~'} // Start of paste
	EOP = []byte{Escape, '[', '2', '0', '1', '~'} // End of paste
	CLR = []byte{' ', Escape, '[', 'D'}
)

// IsPrintable returns true iff key is a visible, non-whitespace key.
func IsPrintable(key rune) bool {
	return key >= Space && (key < Unknown || surrogateMask < key)
}

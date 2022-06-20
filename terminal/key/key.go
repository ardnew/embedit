package key

import (
	"unicode/utf8"

	"github.com/ardnew/embedit/seq/ansi"
)

// Error defines an invalid Rune.
//
// The conventional value utf8.RuneError is not used because it is considered
// a valid rune by package unicode/utf8 and returns 3-bytes from RuneLen.
// Using a proper invalid UTF-8 value allows us to catch invalid runes and
// strings in a sane manner without potentially using up precious bytes.
const Error = utf8.MaxRune + 1

// Application-defined control key codes.
const (
	Unknown rune = 0xD800 /* UTF-16 surrogate area */ + iota
	Up
	Down
	Left
	Right
	AltLeft
	AltRight
	Enter
	Backspace
	Home
	End
	Insert
	Delete
	PageUp
	PageDown
	DeleteWord
	Kill
	KillPrevious
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
	Interrupt
	EndOfFile
	surrogateMask = Unknown | 0x03FF
)

// IsControl returns true iff key is a control key code.
func IsControl(key rune) bool {
	return Unknown < key && key < surrogateMask
}

// IsPrintable returns true iff key is a visible, non-whitespace key.
func IsPrintable(key rune) bool {
	return key >= ansi.Space && !IsControl(key)
}

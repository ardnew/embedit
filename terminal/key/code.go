package key

import (
	"bytes"
	"unicode/utf8"
)

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

// CRLF is the output line delimiter
const (
	CRLF  = "\r\n"
	Blank = " "
)

// ANSI control sequences
const (
	csi        = string(rune(Escape)) + `[` // Extra rune cast to suppress warning
	pasteStart = csi + `200~`
	pasteEnd   = csi + `201~`
	erase      = ` ` + csi + `D`
)

// Parse tries to parse a key sequence from b.
// If successful, it returns the key and the remainder of b.
// Otherwise, it returns utf8.RuneError.
func Parse(b []byte, pasting bool) (rune, []byte) {
	if len(b) == 0 {
		return utf8.RuneError, nil
	}

	if !pasting {
		switch b[0] {
		case CtrlA:
			return Home, b[1:]
		case CtrlB:
			return Left, b[1:]
		case CtrlE:
			return End, b[1:]
		case CtrlF:
			return Right, b[1:]
		case CtrlH:
			return Backspace, b[1:]
		case CtrlK: // <-- TBD
			return DeleteLine, b[1:]
		case CtrlL:
			return ClearScreen, b[1:]
		case CtrlN:
			return Down, b[1:]
		case CtrlP:
			return Up, b[1:]
		case CtrlU:
			return DeleteLine, b[1:]
		case CtrlW:
			return DeleteWord, b[1:]
		}
	}

	if b[0] != Escape {
		if !utf8.FullRune(b) {
			return utf8.RuneError, b
		}
		r, l := utf8.DecodeRune(b)
		return r, b[l:]
	}

	if bytes.HasPrefix(b, bytes.NewBufferString(csi).Bytes()) {
		if pasting {
			if bytes.HasPrefix(b, bytes.NewBufferString(pasteEnd).Bytes()) {
				return PasteEnd, b[len(pasteEnd):]
			}
		} else {
			if len(b) >= 3 {
				switch b[2] {
				case 'A':
					return Up, b[3:]
				case 'B':
					return Down, b[3:]
				case 'C':
					return Right, b[3:]
				case 'D':
					return Left, b[3:]
				case 'H':
					return Home, b[3:]
				case 'F':
					return End, b[3:]
				}
				if len(b) >= 6 && b[2] == '1' && b[3] == ';' && b[4] == '3' {
					switch b[5] {
					case 'C':
						return AltRight, b[6:]
					case 'D':
						return AltLeft, b[6:]
					}
				}
			}
			if bytes.HasPrefix(b, bytes.NewBufferString(pasteStart).Bytes()) {
				return PasteStart, b[len(pasteStart):]
			}
		}
	}

	// If we get here then we have a key that we don't recognise, or a partial
	// sequence.
	// It's not clear how one should find the end of a sequence without knowing
	// them all, but it seems [a-zA-Z~] only appears at the end of a sequence.
	for i, c := range b[0:] {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '~' {
			return Unknown, b[i+1:]
		}
	}

	return utf8.RuneError, b
}

// IsPrintable returns true iff key is a visible, non-whitespace key.
func IsPrintable(key rune) bool {
	return key >= Space && (key < Unknown || surrogateMask < key)
}

// VisibleLen returns the number of visible glyphs in s.
func VisibleLen(s []rune) int {
	escape := false
	length := 0
	for _, r := range s {
		switch {
		case escape:
			escape = (r < 'a' || 'z' < r) && (r < 'A' || 'Z' < r)
		case r == Escape:
			escape = true
		default:
			length++
		}
	}
	return length
}

package key

import (
	"bytes"
	"unicode/utf8"

	"github.com/ardnew/embedit/sequence/key"
)

type Parser interface {
	Parse()
}

// Parse tries to parse a key sequence from b.
// If successful, it returns the key and the remainder of b.
// Otherwise, it returns key.Error.
func Parse(b []byte, pasting bool) (rune, []byte) {
	if len(b) == 0 {
		return key.Error, nil
	}

	if !pasting {
		switch b[0] {
		case key.CtrlA:
			return key.Home, b[1:]
		case key.CtrlB:
			return key.Left, b[1:]
		case key.CtrlE:
			return key.End, b[1:]
		case key.CtrlF:
			return key.Right, b[1:]
		case key.CtrlH:
			return key.Backspace, b[1:]
		case key.CtrlK: // <-- TBD
			return key.DeleteLine, b[1:]
		case key.CtrlL:
			return key.ClearScreen, b[1:]
		case key.CtrlN:
			return key.Down, b[1:]
		case key.CtrlP:
			return key.Up, b[1:]
		case key.CtrlU:
			return key.DeleteLine, b[1:]
		case key.CtrlW:
			return key.DeleteWord, b[1:]
		}
	}

	if b[0] != key.Escape {
		if !utf8.FullRune(b) {
			return key.Error, b
		}
		r, l := utf8.DecodeRune(b)
		return r, b[l:]
	}

	if bytes.HasPrefix(b, key.CSI) {
		if pasting {
			if bytes.HasPrefix(b, key.EOP) {
				return key.PasteEnd, b[len(key.EOP):]
			}
		} else {
			if len(b) >= 3 {
				switch b[2] {
				case 'A':
					return key.Up, b[3:]
				case 'B':
					return key.Down, b[3:]
				case 'C':
					return key.Right, b[3:]
				case 'D':
					return key.Left, b[3:]
				case 'H':
					return key.Home, b[3:]
				case 'F':
					return key.End, b[3:]
				}
				if len(b) >= 6 && b[2] == '1' && b[3] == ';' && b[4] == '3' {
					switch b[5] {
					case 'C':
						return key.AltRight, b[6:]
					case 'D':
						return key.AltLeft, b[6:]
					}
				}
			}
			if bytes.HasPrefix(b, key.SOP) {
				return key.PasteStart, b[len(key.SOP):]
			}
		}
	}

	// If we get here then we have a key that we don't recognise, or a partial
	// sequence.
	// It's not clear how one should find the end of a sequence without knowing
	// them all, but it seems [a-zA-Z~] only appears at the end of a sequence.
	for i, c := range b[0:] {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '~' {
			return key.Unknown, b[i+1:]
		}
	}

	return key.Error, b
}

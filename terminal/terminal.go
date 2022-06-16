package terminal

import (
	"io"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/seq"
	"github.com/ardnew/embedit/seq/eol"
	"github.com/ardnew/embedit/seq/key"
	"github.com/ardnew/embedit/terminal/clipboard/paste"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/history"
	"github.com/ardnew/embedit/terminal/line"
	"github.com/ardnew/embedit/terminal/wire"
)

// Terminal contains the state and configuration of an input/output user
// interface device.
// The rw field abstracts how input/output is implemented by the host.
type Terminal struct {
	rw      io.ReadWriter
	control wire.Control

	cursor  cursor.Cursor
	display display.Display
	history history.History

	in  seq.Buffer
	out seq.Buffer

	paste paste.State

	valid bool
}

// Configure initializes the Terminal configuration.
func (t *Terminal) Configure(
	rw io.ReadWriter, prompt []rune, width, height int,
) *Terminal {
	t.valid = false
	t.rw = rw
	t.history.Configure(
		t.cursor.Configure(
			t.control.Configure(t, t.in.Configure(eol.LF), t.out.Configure(eol.CRLF)),
			t.display.Configure(width, height, prompt, true),
		))
	return t.init()
}

// init initializes the state of a configured Terminal.
func (t *Terminal) init() *Terminal {
	t.valid = true
	t.paste = paste.Inactive
	t.Line().Overwrite(nil)
	return t
}

// Swell copies bytes from an input device to the receiver's input buffer.
func (t *Terminal) Swell() (n int, err error) {
	i, err := io.Copy(&t.in, t.rw)
	return int(i), err
}

// Flush copies bytes from the receiver's output buffer to an output device.
func (t *Terminal) Flush() (n int, err error) {
	i, err := io.Copy(t.rw, &t.out)
	return int(i), err
}

func (t *Terminal) Cursor() *cursor.Cursor {
	return &t.cursor
}

func (t *Terminal) Line() *line.Line {
	return t.history.Line()
}

// PressKey processes a given keypress on the current line.
func (t *Terminal) PressKey(k rune) (ok bool) {
	l := t.Line()
	if t.paste.IsActive() && k != key.Enter {
		return l.InsertRune(k) == nil
	}

	pos := l.Position()
	siz := l.RuneCount()
	// viz := l.GlyphCount()

	switch k {
	case key.Backspace:
		if pos == 0 {
			return
		}
		ok = l.ErasePrevRune(1) == nil

	case key.AltLeft:
		// Move left by 1 word.
		ok = l.MoveTo(pos-l.RuneCountToPrevWord()) == nil

	case key.AltRight:
		// Move right by 1 word.
		ok = l.MoveTo(pos+l.RuneCountToNextWord()) == nil

	case key.Left:
		if pos == 0 {
			return
		}
		ok = l.MoveTo(pos-1) == nil

	case key.Right:
		if pos == siz {
			return
		}
		ok = l.MoveTo(pos+1) == nil

	case key.Home:
		if pos == 0 {
			return
		}
		ok = l.MoveTo(0) == nil

	case key.End:
		if pos == siz {
			return
		}
		ok = l.MoveTo(siz) == nil

		/*
			case key.Up:
				entry, ok := t.history.NthPreviousEntry(t.historyIndex + 1)
				if !ok {
					return "", false
				}
				if t.historyIndex == -1 {
					t.historyPending = string(t.line)
				}
				t.historyIndex++
				runes := []rune(entry)
				t.setLine(runes, len(runes))
			case key.Down:
				switch t.historyIndex {
				case -1:
					return
				case 0:
					runes := []rune(t.historyPending)
					t.setLine(runes, len(runes))
					t.historyIndex--
				default:
					entry, ok := t.history.NthPreviousEntry(t.historyIndex - 1)
					if ok {
						t.historyIndex--
						runes := []rune(entry)
						t.setLine(runes, len(runes))
					}
				}
		*/

	case key.Enter:
		l.MoveTo(siz)
		t.out.WriteEOL()
		l.LineFeed()

	case key.DeleteWord:
		// Delete zero or more spaces and then one or more characters.
		l.ErasePrevRune(l.RuneCountToPrevWord())

	case key.DeleteLine:
		// Delete everything from the current cursor position to the
		// end of line.
		l.MoveTo(siz)
		l.ErasePrevRune(siz - pos)

	case key.CtrlD:
		// Erase the character under the current position.
		// The EOF case when the line is empty is handled in
		// readLine().
		if pos < siz {
			l.MoveTo(pos + 1)
			l.ErasePrevRune(1)
		}

	case key.CtrlU:
		l.ErasePrevRune(pos)

	case key.ClearScreen:
		// Erase the screen and move the cursor to the home position.
		l.ClearScreen()

	default:
		// if t.AutoCompleteCallback != nil {
		// 	prefix := string(t.line[:pos])
		// 	suffix := string(t.line[pos:])

		// 	t.lock.Unlock()
		// 	newLine, newPos, completeOk := t.AutoCompleteCallback(prefix+suffix, len(prefix), k)
		// 	t.lock.Lock()

		// 	if completeOk {
		// 		t.setLine([]rune(newLine), utf8.RuneCount([]byte(newLine)[:newPos]))
		// 		return
		// 	}
		// }
		if !key.IsPrintable(k) {
			return
		}
		if l.RuneCount() == config.RunesPerLine {
			return
		}
		l.InsertRune(k)

	}
	return
}

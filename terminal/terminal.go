package terminal

import (
	"io"

	"github.com/ardnew/embedit/config/limits"
	"github.com/ardnew/embedit/errors"
	"github.com/ardnew/embedit/seq"
	"github.com/ardnew/embedit/seq/eol"
	"github.com/ardnew/embedit/terminal/clipboard/paste"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/history"
	"github.com/ardnew/embedit/terminal/key"
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
	rw io.ReadWriter, prompt []rune, width, height int, flush bool,
) *Terminal {
	t.valid = false
	t.rw = rw
	t.history.Configure(
		flush,
		t.cursor.Configure(
			flush,
			t.control.Configure(t,
				t.in.Configure(eol.LF),
				t.out.Configure(eol.CRLF)),
			t.display.Configure(width, height, prompt, true),
		))
	return t.init()
}

// init initializes the state of a configured Terminal.
func (t *Terminal) init() *Terminal {
	t.valid = true
	t.paste = paste.Inactive
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

func (t *Terminal) ReadLine() (err error) {
	wasEnabled := t.display.EnablePrompt(true)
	defer t.display.EnablePrompt(wasEnabled)
	l := t.Line()
	if err = l.ShowPrompt(); err != nil {
		return
	}
	eol := false
	for !eol {
		for t.in.Len() > 0 {
			k, sz := t.in.Parse(t.paste.IsActive())
			if k == key.Unknown {
				l.MoveCursorTo(l.RuneCount())
				t.out.WriteEOL()
				t.cursor.WriteBuf(t.in.Last())
				eol = true
			} else {
				if k == key.Error || sz == 0 {
					break
				}
				eol, err = t.HandleKey(k)
			}
		}
		if eol {
			if l.IsPasted() {
				err = &errors.ErrPasteIndicator
			}
			t.Flush()
		} else {
			t.Flush()
			t.Swell()
		}
	}
	return
}

func (t *Terminal) HandleKey(k rune) (eol bool, err error) {
	eol, err = t.handleKey(k)
	if eol && t.display.Echo() {
		t.history.Add()
		t.out.WriteEOL()
	}
	return
}

// HandleKey processes a given keypress on the current line.
func (t *Terminal) handleKey(k rune) (eol bool, err error) {
	l := t.Line()
	// If we are actively pasting, all keys other than Enter and the end-of-paste
	// sequence should be inserted literally into the line.
	if t.paste.IsActive() && k != key.Enter && k != key.PasteEnd {
		return false, l.InsertRune(k)
	}

	pos := l.Position()
	siz := l.RuneCount()

	switch k {

	case key.Enter:
		l.MoveCursorTo(siz)
		eol = true

	case key.Backspace:
		if pos > 0 {
			l.ErasePreviousRuneCount(1)
		}

	case key.Interrupt:
		eol = true
		err = io.ErrUnexpectedEOF

	case key.EndOfFile:
		if siz == 0 {
			eol = true
			err = io.EOF
		} else if pos < siz {
			// Erase the character under the current position — "rubout".
			l.MoveCursor(+1)
			l.ErasePreviousRuneCount(1)
		}

	case key.Up:
		t.history.Back()

	case key.Down:
		t.history.Forward()

	case key.Left:
		if pos > 0 {
			l.MoveCursor(-1)
		}

	case key.Right:
		if pos < siz {
			l.MoveCursor(+1)
		}

	case key.AltLeft:
		// Move left by 1 word.
		l.MoveCursor(-l.RuneCountToStartOfWord())

	case key.AltRight:
		// Move right by 1 word.
		l.MoveCursor(+l.RuneCountToStartOfNextWord())

	case key.Home:
		if pos > 0 {
			l.MoveCursorTo(0)
		}

	case key.End:
		if pos < siz {
			l.MoveCursorTo(siz)
		}

	case key.Delete:
		if pos < siz {
			// Erase the character under the current position — "rubout".
			l.MoveCursor(+1)
			l.ErasePreviousRuneCount(1)
		}

	case key.DeleteWord:
		// Move to the end of the current word iff cursor is not on white space.
		l.MoveCursor(+l.RuneCountToEndOfWord())
		// Delete zero or more spaces and then one or more characters.
		l.ErasePreviousRuneCount(l.RuneCountToStartOfWord())

	case key.KillPrevious:
		// Delete everything from the current cursor position to the start of line.
		l.ErasePreviousRuneCount(pos)

	case key.Kill:
		// Delete everything from the current cursor position to the end of line.
		l.MoveCursorTo(siz)
		l.ErasePreviousRuneCount(siz - pos)

	case key.ClearScreen:
		// Erase the screen and move the cursor to the home position.
		l.ClearScreen()
		l.ShowPrompt()

	case key.PasteStart:
		t.paste = paste.Active
		if siz == 0 {
			l.SetIsPasted(true)
		}

	case key.PasteEnd:
		t.paste = paste.Inactive

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
		if siz < limits.RunesPerLine {
			// If we've reached here, then we are inserting a key outside of a bracketed
			// paste operation.
			l.SetIsPasted(false)

			if key.IsPrintable(k) {
				err = l.InsertRune(k)
			}
		}

	}
	return
}

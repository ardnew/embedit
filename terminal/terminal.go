package terminal

import (
	"io"
	"unicode/utf8"

	"github.com/ardnew/embedit/sequence"
	"github.com/ardnew/embedit/sequence/eol"
	"github.com/ardnew/embedit/sequence/key"
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

	in  sequence.Sequence
	out sequence.Sequence

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
	t.Line().Set(nil)
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
	if t.pasteActive && k != key.Enter {
		t.addKeyToLine(k)
		return
	}

	switch k {
	case key.Backspace:
		if t.pos == 0 {
			return
		}
		t.eraseNPreviousChars(1)
	case key.AltLeft:
		// move left by a word.
		t.pos -= t.countToLeftWord()
		t.moveCursorToPos(t.pos)
	case key.AltRight:
		// move right by a word.
		t.pos += t.countToRightWord()
		t.moveCursorToPos(t.pos)
	case key.Left:
		if t.pos == 0 {
			return
		}
		t.pos--
		t.moveCursorToPos(t.pos)
	case key.Right:
		if t.pos == len(t.line) {
			return
		}
		t.pos++
		t.moveCursorToPos(t.pos)
	case key.Home:
		if t.pos == 0 {
			return
		}
		t.pos = 0
		t.moveCursorToPos(t.pos)
	case key.End:
		if t.pos == len(t.line) {
			return
		}
		t.pos = len(t.line)
		t.moveCursorToPos(t.pos)
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
	case key.Enter:
		t.moveCursorToPos(len(t.line))
		t.queue([]rune("\r\n"))
		line = string(t.line)
		ok = true
		t.line = t.line[:0]
		t.pos = 0
		t.cursorX = 0
		t.cursorY = 0
		t.maxLine = 0
	case key.DeleteWord:
		// Delete zero or more spaces and then one or more characters.
		t.eraseNPreviousChars(t.countToLeftWord())
	case key.DeleteLine:
		// Delete everything from the current cursor position to the
		// end of line.
		for i := t.pos; i < len(t.line); i++ {
			t.queue(space)
			t.advanceCursor(1)
		}
		t.line = t.line[:t.pos]
		t.moveCursorToPos(t.pos)
	case key.CtrlD:
		// Erase the character under the current position.
		// The EOF case when the line is empty is handled in
		// readLine().
		if t.pos < len(t.line) {
			t.pos++
			t.eraseNPreviousChars(1)
		}
	case key.CtrlU:
		t.eraseNPreviousChars(t.pos)
	case key.ClearScreen:
		// Erases the screen and moves the cursor to the home position.
		t.queue([]rune("\x1b[2J\x1b[H"))
		t.queue(t.prompt)
		t.cursorX, t.cursorY = 0, 0
		t.advanceCursor(visualLength(t.prompt))
		t.setLine(t.line, t.pos)
	default:
		if t.AutoCompleteCallback != nil {
			prefix := string(t.line[:t.pos])
			suffix := string(t.line[t.pos:])

			t.lock.Unlock()
			newLine, newPos, completeOk := t.AutoCompleteCallback(prefix+suffix, len(prefix), k)
			t.lock.Lock()

			if completeOk {
				t.setLine([]rune(newLine), utf8.RuneCount([]byte(newLine)[:newPos]))
				return
			}
		}
		if !isPrintable(k) {
			return
		}
		if len(t.line) == maxLineLength {
			return
		}
		t.addKeyToLine(k)
	}
	return
}

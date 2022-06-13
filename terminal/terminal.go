package terminal

import (
	"io"

	"github.com/ardnew/embedit/sequence"
	"github.com/ardnew/embedit/sequence/eol"
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
	in      sequence.Sequence
	out     sequence.Sequence
	valid   bool
}

// Configure initializes the Terminal configuration.
func (t *Terminal) Configure(
	rw io.ReadWriter, prompt []rune, width, height int,
) *Terminal {
	t.valid = false
	t.rw = rw
	t.history.Configure(
		t.cursor.Configure(
			t.control.Configure(t, t.in.Configure(eol.CRLF), t.out.Configure(eol.CRLF)),
			t.display.Configure(width, height, prompt, true),
		))
	return t.init()
}

// init initializes the state of a configured Terminal.
func (t *Terminal) init() *Terminal {
	t.valid = true
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

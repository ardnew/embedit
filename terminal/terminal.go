package terminal

import (
	"io"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/sequence"
	"github.com/ardnew/embedit/terminal/line"
	"github.com/ardnew/embedit/volatile"
)

// Terminal contains the state and configuration of an input/output user
// interface device.
// The rw field abstracts how input/output is implemented by the host.
type Terminal struct {
	rw     io.ReadWriter
	prompt []rune
	line   line.Line
	i      sequence.Sequence
	o      sequence.Sequence
	width  volatile.Register32
	height volatile.Register32
	valid  bool
}

// Configure initializes the Terminal configuration.
func (t *Terminal) Configure(rw io.ReadWriter, prompt []rune, width, height int) *Terminal {
	t.valid = false
	t.rw = rw
	_ = t.i.Configure()
	_ = t.o.Configure()
	_ = t.line.Configure(t, t)
	t.prompt = prompt
	t.width.Set(uint32(width))
	t.height.Set(uint32(height))
	return t.init()
}

// init initializes the state of a configured Terminal.
func (t *Terminal) init() *Terminal {
	t.valid = true
	if t.prompt == nil {
		t.prompt = []rune(config.DefaultPrompt)
	}
	return t
}

// Width returns the Terminal width.
func (t *Terminal) Width() int { return int(t.width.Get()) }

// Height returns the Terminal height.
func (t *Terminal) Height() int { return int(t.height.Get()) }

// Size returns the Terminal width and height.
func (t *Terminal) Size() (width, height int) {
	return int(t.width.Get()), int(t.height.Get())
}

// SetSize sets the Terminal width and height.
func (t *Terminal) SetSize(width, height int) {
	if width <= 0 {
		width = config.DefaultWidth
	}
	if height <= 0 {
		height = config.DefaultHeight
	}
	t.width.Set(uint32(width))
	t.height.Set(uint32(height))
}

// Echo returns true if and only if input keystrokes are echoed to output.
func (t *Terminal) Echo() bool {
	return true
}

// Line returns the active user input line.
func (t *Terminal) Line() *line.Line {
	if t == nil {
		return nil
	}
	return &t.line
}

// Prompt returns the user input prompt.
func (t *Terminal) Prompt() []rune {
	if t.prompt == nil {
		return []rune(config.DefaultPrompt)
	}
	return t.prompt
}

// Read copies up to len(p) bytes from the receiver's input buffer to p.
func (t *Terminal) Read(p []byte) (n int, err error) {
	return t.i.Read(p)
}

// ReadWire copies bytes from an input device to the receiver's input buffer.
func (t *Terminal) ReadWire() (n int, err error) {
	i, err := io.Copy(&t.i, t.rw)
	return int(i), err
}

// Write copies up to len(p) bytes from p to the receiver's output buffer.
func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.o.Append(p)
}

// WriteWire copies bytes from the receiver's output buffer to an output device.
func (t *Terminal) WriteWire() (n int, err error) {
	i, err := io.Copy(t.rw, &t.o)
	return int(i), err
}

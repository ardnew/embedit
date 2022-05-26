package terminal

import (
	"io"

	"github.com/ardnew/embedit/cursor"
	"github.com/ardnew/embedit/sequence"
	"github.com/ardnew/embedit/volatile"
)

// Terminal contains the state and configuration of an input/output user
// interface device.
// The rw field abstracts how input/output is implemented by the host.
type Terminal struct {
	rw     io.ReadWriter
	Pos    cursor.Cursor
	i      sequence.Sequence // Input byte buffer
	o      sequence.Sequence // Output byte buffer
	width  volatile.Register32
	height volatile.Register32
	valid  bool
}

// Configure initializes the Terminal configuration.
func (t *Terminal) Configure(rw io.ReadWriter, width, height int) *Terminal {
	t.valid = false
	t.rw = rw
	_ = t.Pos.Configure(t)
	_ = t.i.Configure()
	_ = t.o.Configure()
	t.width.Set(uint32(width))
	t.height.Set(uint32(height))
	return t.init()
}

// init initializes the state of a configured Terminal.
func (t *Terminal) init() *Terminal {
	t.valid = true
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

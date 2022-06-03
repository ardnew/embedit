package cursor

import (
	"bytes"
	"io"

	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/key"
	"github.com/ardnew/embedit/terminal/wire"
	"github.com/ardnew/embedit/text/ascii"
	"github.com/ardnew/embedit/text/utf8"
	"github.com/ardnew/embedit/volatile"
)

// Cursor defines the position of the cursor.
type Cursor struct {
	disp  display.Display
	wire  wire.ReadWriter
	x, y  volatile.Register32 // X=0: left edge; Y=0: first row, current Line
	maxY  volatile.Register32 // Greatest value of Y so far
	valid bool
}

// Configure initializes the Cursor configuration.
func (c *Cursor) Configure(disp display.Display, wire wire.ReadWriter) *Cursor {
	c.valid = false
	c.disp = disp
	c.wire = wire
	return c.init()
}

// init initializes the state of a configured Cursor.
func (c *Cursor) init() *Cursor {
	c.valid = true
	c.maxY.Set(0)
	_, _ = c.Set(0, 0)
	return c
}

// Reset resets the X, Y, and MaxY coordinates.
func (c *Cursor) Reset() *Cursor {
	c.maxY.Set(0)
	_, _ = c.Set(0, 0)
	return c
}

// Echo returns true if and only if input keystrokes are echoed to output.
func (c *Cursor) Echo() bool {
	return c != nil && c.disp != nil && c.disp.Echo()
}

// X returns the Cursor's X coordinate.
func (c *Cursor) X() int { return int(c.x.Get()) }

// Y returns the Cursor's Y coordinate.
func (c *Cursor) Y() int { return int(c.y.Get()) }

// Get returns the Cursor's X, Y coordinates.
func (c *Cursor) Get() (x, y int) { return int(c.x.Get()), int(c.y.Get()) }

// Set sets the Cursor's X, Y coordinates and updates MaxY if necessary.
// If the given coordinates are out-of-bounds, uses the nearest boundary values.
// Returns the boundary-limited X, Y coordinates.
func (c *Cursor) Set(x, y int) (X, Y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	c.x.Set(uint32(x))
	c.y.Set(uint32(y))
	if y > int(c.maxY.Get()) {
		c.maxY.Set(uint32(y))
	}
	return x, y
}

// WriteLine appends line to the output buffer and advances the cursor's current
// position to the end of the line.
func (c *Cursor) WriteLine(line []rune) (err error) {
	width := c.disp.Width()
	for len(line) != 0 {
		free := width - c.X()
		todo := len(line)
		if todo > free {
			todo = free
		}
		var n int
		for _, r := range line[:todo] {
			nc, errc := io.Copy(c.wire, utf8.Rune(r))
			if errc != nil {
				continue
			}
			n += int(nc)
		}
		aerr := c.Advance(key.VisibleLen(line[:todo]))
		if err == nil {
			err = aerr
		}
		if err != nil {
			return
		}
		line = line[len(line[:todo]):]
	}
	return
}

// Move appends key sequences to the output buffer which will move the cursor
// in the given directions by the given number of positions, relative to the
// cursor's current position.
// The key sequences are buffered but not written to output.
func (c *Cursor) Move(up, down, left, right int) (err error) {
	// 1 unit up can be expressed as ^[[A or ^[A
	// 5 units up can be expressed as ^[[5A
	// - - - - - - - - - - - - - - - - - - - -
	// Up
	if up == 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		c.wire.WriteByte('A')
	} else if up > 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		io.Copy(c.wire, ascii.Uint32(up))
		c.wire.WriteByte('A')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Down
	if down == 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		c.wire.WriteByte('B')
	} else if down > 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		io.Copy(c.wire, ascii.Uint32(down))
		c.wire.WriteByte('B')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Right
	if right == 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		c.wire.WriteByte('C')
	} else if right > 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		io.Copy(c.wire, ascii.Uint32(right))
		c.wire.WriteByte('C')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Left
	if left == 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		c.wire.WriteByte('D')
	} else if left > 1 {
		c.wire.WriteByte(key.Escape)
		c.wire.WriteByte('[')
		io.Copy(c.wire, ascii.Uint32(left))
		c.wire.WriteByte('D')
	}
	c.wire.WriteWire()
	return
}

// MoveTo appends key sequences to the output buffer which will move the cursor
// to the given logical position in the text.
// The key sequences are buffered but not written to output.
func (c *Cursor) MoveTo(pos int) error {
	if !c.disp.Echo() {
		return nil
	}
	x := pos + key.VisibleLen(c.disp.Prompt())
	y := x / c.disp.Width()
	x %= c.disp.Width()
	var (
		cX, cY     = c.Get()
		u, d, l, r int
	)
	if y < cY {
		u = cY - y
	}
	if y > cY {
		d = y - cY
	}
	if x < cX {
		l = cX - x
	}
	if x > cX {
		r = x - cX
	}
	_, _ = c.Set(x, y)
	return c.Move(u, d, l, r)
}

// Advance updates the X, Y coordinates — and possibly appends CR, LF to the
// output buffer, if necessary — based on the given number of places to move
// along an input line.
func (c *Cursor) Advance(places int) (err error) {
	x, y := c.Get()
	x += places
	y += x / c.disp.Width()
	x %= c.disp.Width()
	x, _ = c.Set(x, y)
	// Normally terminals will advance the current position when writing a
	// character. But that doesn't happen for the last character in a line.
	// However, when writing a character (except a new line) that causes a line
	// wrap, the position will be advanced two places.
	//
	// So, if we are stopping at the end of a line, we need to write a newline so
	// that our cursor can be advanced to the next line.
	if places > 0 && x == 0 {
		_, err = c.wire.Write(bytes.NewBufferString(key.CRLF).Bytes())
	}
	return
}

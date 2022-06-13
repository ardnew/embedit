package cursor

import (
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/key"
	"github.com/ardnew/embedit/terminal/wire"
	"github.com/ardnew/embedit/text/ascii"
	"github.com/ardnew/embedit/volatile"
)

// Cursor defines the position of the cursor.
type Cursor struct {
	ctrl  *wire.Control
	disp  *display.Display
	x, y  volatile.Register32 // X=0: left edge; Y=0: first row, current Line
	maxY  volatile.Register32 // Greatest value of Y so far
	ascii ascii.Uint32
	valid bool
}

// Configure initializes the Cursor configuration.
func (c *Cursor) Configure(ctrl *wire.Control, disp *display.Display) *Cursor {
	c.valid = false
	c.ctrl = ctrl
	c.disp = disp
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

func (c *Cursor) Control() *wire.Control    { return c.ctrl }
func (c *Cursor) Display() *display.Display { return c.disp }

// X returns the Cursor's X coordinate.
func (c *Cursor) X() int { return int(c.x.Get()) }

// Y returns the Cursor's Y coordinate.
func (c *Cursor) Y() int { return int(c.y.Get()) }

// Get returns the Cursor's X, Y coordinates.
func (c *Cursor) Get() (x, y int) { return int(c.x.Get()), int(c.y.Get()) }

// Set sets the X, Y coordinates and updates MaxY if necessary.
// If the given coordinates are out-of-bounds, uses the nearest boundary values
// without row/column-wrapping.
// Returns the boundary-limited X, Y coordinates.
func (c *Cursor) Set(x, y int) (X, Y int) {
	w, h := c.disp.Size()
	if x < 0 {
		x = 0
	} else if x >= w {
		x = w - 1
	}
	if y < 0 {
		y = 0
	} else if y >= h {
		y = h - 1
	}
	c.x.Set(uint32(x))
	c.y.Set(uint32(y))
	if y > int(c.maxY.Get()) {
		c.maxY.Set(uint32(y))
	}
	return x, y
}

// Update sets the X, Y coordinates based on the given number of places to move
// horizontally – relative to the current position – and returns true if and
// only if the final cursor position is at the end of a line on the display.
// Update adjusts the X, Y coordinates to account for line-wrapping.
//
// Update is called after runes have been copied to the output buffer, and the
// cursor needs its coordinates updated to reflect the number of visible glyphs
// that will be drawn to the display. It does not read or write from any output
// buffer or device.
func (c *Cursor) Update(places int) (eol bool) {
	x, y := c.Get()
	w := c.disp.Width()
	x += places
	y += x / w
	x %= w
	x, _ = c.Set(x, y)
	// Normally terminals will advance the current position when writing a
	// character. But that doesn't happen for the last character in a line.
	// However, when writing a character (except a newline) that causes a line
	// wrap, the position will be advanced two places.
	//
	// So, if we are stopping at the end of a line, we need to write a newline so
	// that our cursor can be advanced to the next line.
	return places > 0 && x == 0
}

// Queue appends key sequences to the output buffer which will move the cursor
// in the given directions by the given number of positions, relative to the
// cursor's current position.
//
// Queue does not update the cursor's current X, Y coordinates. The caller is
// responsible for updating the coordinates, which must be done every time Queue
// is called.
func (c *Cursor) Queue(up, down, left, right int) (err error) {
	// 1 unit up can be expressed as ^[[A or ^[A
	// 5 units up can be expressed as ^[[5A
	// - - - - - - - - - - - - - - - - - - - -
	// Up
	if up == 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ctrl.Out.WriteByte('A')
	} else if up > 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ascii.Val = uint32(up)
		c.ascii.WriteTo(c.ctrl.Out)
		c.ctrl.Out.WriteByte('A')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Down
	if down == 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ctrl.Out.WriteByte('B')
	} else if down > 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ascii.Val = uint32(down)
		c.ascii.WriteTo(c.ctrl.Out)
		c.ctrl.Out.WriteByte('B')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Left
	if left == 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ctrl.Out.WriteByte('D')
	} else if left > 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ascii.Val = uint32(left)
		c.ascii.WriteTo(c.ctrl.Out)
		c.ctrl.Out.WriteByte('D')
	}
	// - - - - - - - - - - - - - - - - - - - -
	// Right
	if right == 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ctrl.Out.WriteByte('C')
	} else if right > 1 {
		c.ctrl.Out.WriteByte(key.Escape)
		c.ctrl.Out.WriteByte('[')
		c.ascii.Val = uint32(right)
		c.ascii.WriteTo(c.ctrl.Out)
		c.ctrl.Out.WriteByte('C')
	}
	_, err = c.ctrl.Flush()
	return
}

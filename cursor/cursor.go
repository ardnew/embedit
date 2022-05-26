package cursor

import (
	"strconv"

	"github.com/ardnew/embedit/key"
	"github.com/ardnew/embedit/volatile"
)

// Cursor defines the position of the cursor.
type Cursor struct {
	screen Screener
	X, Y   volatile.Register32 // X=0: left edge; Y=0: first row, current Line
	MaxY   volatile.Register32 // Greatest value of Y so far
	valid  bool
}

// Screener defines the area in which a Cursor can be positioned.
type Screener interface {
	Width() int
	Height() int
	Size() (width, height int)
}

// Configure initializes the Cursor configuration.
func (c *Cursor) Configure(s Screener) *Cursor {
	c.valid = false
	c.screen = s
	return c.init()
}

// init initializes the state of a configured Cursor.
func (c *Cursor) init() *Cursor {
	c.valid = true
	c.X.Set(0)
	c.Y.Set(0)
	c.MaxY.Set(0)
	return c
}

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
	c.X.Set(uint32(x))
	c.Y.Set(uint32(y))
	if y > int(c.MaxY.Get()) {
		c.MaxY.Set(uint32(y))
	}
	return int(c.X.Get()), int(c.Y.Get())
}

func (c *Cursor) move(up, down, left, right int) {
	m := []rune{}

	// 1 unit up can be expressed as ^[[A or ^[A
	// 5 units up can be expressed as ^[[5A

	if up == 1 {
		m = append(m, key.Escape, '[', 'A')
	} else if up > 1 {
		m = append(m, key.Escape, '[')
		m = append(m, []rune(strconv.Itoa(up))...)
		m = append(m, 'A')
	}

	if down == 1 {
		m = append(m, key.Escape, '[', 'B')
	} else if down > 1 {
		m = append(m, key.Escape, '[')
		m = append(m, []rune(strconv.Itoa(down))...)
		m = append(m, 'B')
	}

	if right == 1 {
		m = append(m, key.Escape, '[', 'C')
	} else if right > 1 {
		m = append(m, key.Escape, '[')
		m = append(m, []rune(strconv.Itoa(right))...)
		m = append(m, 'C')
	}

	if left == 1 {
		m = append(m, key.Escape, '[', 'D')
	} else if left > 1 {
		m = append(m, key.Escape, '[')
		m = append(m, []rune(strconv.Itoa(left))...)
		m = append(m, 'D')
	}

	// t.queue(m)
}

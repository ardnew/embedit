package line

import (
	"io"

	"github.com/ardnew/embedit/key"
	"github.com/ardnew/embedit/limit"
	"github.com/ardnew/embedit/volatile"
)

// Types of errors returned by Line methods.
type (
	ErrReceiver  string
	ErrArgument  string
	ErrOverflow  string
	ErrUnderflow string
)

// Line represents a single line of input.
type Line struct {
	Rune  [limit.RunesPerLine]rune
	Pos   volatile.Register32 // Logical cursor position
	valid bool                // Has init been called
}

// Invalid represents an invalid Line.
var Invalid = Line{valid: false}

// Configure initializes the Line configuration.
func (l *Line) Configure() *Line {
	l.valid = false
	return l.init()
}

// init initializes the state of a configured Line.
func (l *Line) init() *Line {
	l.valid = true
	return l.Reset()
}

// Reset sets all runes to key.Null and resets the logical cursor position.
func (l *Line) Reset() *Line {
	for i := range l.Rune {
		l.Rune[i] = key.Null
	}
	l.Pos.Set(0)
	return l
}

func (l *Line) String() string {
	if l.valid {
		return string(l.Rune[0:])
	}
	return ""
}

func (l *Line) Read(a []byte) (n int, err error) {
	if l == nil {
		return 0, ErrReceiver("cannot Read from nil receiver")
	}
	if a == nil {
		return 0, ErrArgument("cannot Read to nil buffer")
	}
	s := string(l.Rune[:])
	c := len(s)
	if c == 0 {
		return 0, io.EOF
	}
	n = copy(a, []byte(s))
	if n == c {
		err = io.EOF
	}
	return
}

func (e ErrReceiver) Error() string  { return "line [receiver]: " + string(e) }
func (e ErrArgument) Error() string  { return "line [argument]: " + string(e) }
func (e ErrOverflow) Error() string  { return "line [overflow]: " + string(e) }
func (e ErrUnderflow) Error() string { return "line [underflow]: " + string(e) }

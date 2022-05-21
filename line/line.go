package line

import (
	"strings"

	"github.com/ardnew/embedit/fifo"
	"github.com/ardnew/embedit/limit"
)

// ErrorRune represents an invalid rune.
const ErrorRune = '\uFFFD'

// Line represents a single line of input. It implements the fifo.Buffer
// interface for use as a fixed-length FIFO.
type Line struct {
	FIFO fifo.State
	Rune [limit.RunesPerLine]rune
	Pos  int
}

// Init initializes all fields of the receiver.
func (l *Line) Init() {
	for i := range l.Rune {
		l.Rune[i] = rune(0)
	}
	l.FIFO.Init(l, len(l.Rune), fifo.DiscardLast)
	l.Pos = 0
}

// Len returns the size of the receiver.
func (l *Line) Len() int {
	return len(l.Rune)
}

// Get returns a Data (of concrete type rune) at the given index and true.
// Returns nil and false if the index is out of bounds.
func (l *Line) Get(i int) (data fifo.Data, ok bool) {
	if l != nil && 0 <= i && i < l.Len() {
		data, ok = l.Rune[i], true
	}
	return
}

// Set sets the rune element at the given index and returns true.
// Returns false if the index is out of bounds or given Data type is not rune.
func (l *Line) Set(i int, data fifo.Data) (ok bool) {
	if l != nil && 0 <= i && i < l.Len() {
		// Don't write to the buffer unless type assertion succeeds.
		var c rune
		if c, ok = data.(rune); ok {
			l.Rune[i] = c
		}
	}
	return
}

func (l *Line) String() string {
	if l == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.Grow(l.FIFO.Len())
	for i := 0; i < l.FIFO.Len(); i++ {
		data, ok := l.FIFO.Get(i)
		if !ok {
			break
		}
		c, ok := data.(rune)
		if !ok || c == rune(0) || c == ErrorRune {
			break
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

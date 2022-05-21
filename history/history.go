package history

import (
	"strings"

	"github.com/ardnew/embedit/fifo"
	"github.com/ardnew/embedit/limit"
	"github.com/ardnew/embedit/line"
)

// History contains previous user-input Lines. It implements the fifo.Buffer
// interface for use as a fixed-length FIFO.
type History struct {
	FIFO fifo.State
	Line [limit.LinesPerHistory]line.Line
	Pos  int
}

// Init initializes all fields of the receiver.
func (h *History) Init() {
	for _, l := range h.Line {
		l.Init()
	}
	h.FIFO.Init(h, len(h.Line), fifo.DiscardFirst)
	h.Pos = 0
}

// Len returns the size of the receiver.
func (h *History) Len() int {
	return len(h.Line)
}

// Get returns a Data (of concrete type rune) at the given index and true.
// Returns nil and false if the index is out of bounds.
func (h *History) Get(i int) (data fifo.Data, ok bool) {
	if h != nil && 0 <= i && i < h.Len() {
		data, ok = h.Line[i], true
	}
	return
}

// Set sets the rune element at the given index and returns true.
// Returns false if the index is out of bounds or given Data type is not rune.
func (h *History) Set(i int, data fifo.Data) (ok bool) {
	if h != nil && 0 <= i && i < h.Len() {
		// Don't write to the buffer unless type assertion succeeds.
		var l line.Line
		if l, ok = data.(line.Line); ok {
			h.Line[i] = l
		}
	}
	return
}

func (h *History) String() string {
	if h == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.Grow(h.FIFO.Len()*(limit.RunesPerLine+4) + 2)
	sb.WriteString("[")
	for i := 0; i < h.FIFO.Len(); i++ {
		data, ok := h.FIFO.Get(i)
		if !ok {
			break
		}
		l, ok := data.(line.Line)
		if !ok {
			break
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		if i == l.Pos {
			sb.WriteRune('<')
		}
		sb.WriteRune('"')
		sb.WriteString(l.String())
		sb.WriteRune('"')
		if i == l.Pos {
			sb.WriteRune('>')
		}
	}
	sb.WriteString("]")
	return sb.String()
}

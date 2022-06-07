package history

import (
	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/line"
	"github.com/ardnew/embedit/volatile"
)

// History contains previous user-input Lines.
type History struct {
	line  [config.LinesPerHistory]line.Line
	pend  line.Line
	head  volatile.Register32
	size  volatile.Register32
	indx  volatile.Register32
	valid bool
}

// Configure initializes the History configuration.
func (h *History) Configure(curs *cursor.Cursor) *History {
	if h == nil {
		return nil
	}
	h.valid = false
	for i := range h.line {
		_ = h.line[i].Configure(curs)
	}
	h.pend.Configure(curs)
	return h.init()
}

// init initializes the state of a configured History.
func (h *History) init() *History {
	h.valid = true
	h.head.Set(0)
	h.size.Set(0)
	h.indx.Set(0)
	return h
}

// Len returns the number of Lines currently stored in History.
func (h *History) Len() int {
	if h == nil {
		return 0
	}
	return int(h.size.Get())
}

// Add appends a Line to the History.
// If the History is filled to capacity, the oldest Line is discarded.
func (h *History) Add(ln line.Line) {
	if h == nil {
		return
	}
	head := (h.head.Get() + 1) % config.LinesPerHistory
	h.head.Set(head)
	h.line[head] = ln
	if size := h.size.Get(); size < config.LinesPerHistory {
		h.size.Set(size + 1)
	}
}

// Get returns the Line passed to the n'th previous call to Add.
// If n=0, the immediately previous Line is returned; if n=1, the Line before
// that, and so on.
// If n<0 or fewer than n+1 lines have been added, ok is false.
func (h *History) Get(n int) (ln *line.Line, ok bool) {
	if h == nil || n < 0 || n >= int(h.size.Get()) {
		return nil, false
	}
	index := int(h.head.Get()) - n
	if index < 0 {
		index += config.LinesPerHistory
	}
	return &h.line[index], true
}

func (h *History) Line() *line.Line {
	return &h.pend
}

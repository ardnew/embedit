package history

import (
	"github.com/ardnew/embedit/config/limits"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/line"
	"github.com/ardnew/embedit/volatile"
)

// History contains previous user-input Lines.
type History struct {
	line  [limits.LinesPerHistory]line.Line
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
	if h.valid {
		// Configure must be called one time only.
		// Use object methods to modify configuration/state.
		return h
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
	h.size.Set(1)
	h.indx.Set(0)
	return h
}

func (h *History) Back() {
	indx, size := h.indx.Get(), h.size.Get()
	if indx < size-1 {
		*h.get(int(indx)) = h.pend
		h.pend.Set(nil)
		indx++
		h.pend = *h.get(int(indx))
		h.pend.Flush()
		h.indx.Set(indx)
	}
}

func (h *History) Forward() {
	indx := h.indx.Get()
	if indx > 0 {
		*h.get(int(indx)) = h.pend
		h.pend.Set(nil)
		indx--
		h.pend = *h.get(int(indx))
		h.pend.Flush()
		h.indx.Set(indx)
	}
}

func (h *History) Line() *line.Line {
	return &h.pend
}

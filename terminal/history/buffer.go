//go:build history
// +build history

package history

import (
	"github.com/ardnew/embedit/config/limits"
	"github.com/ardnew/embedit/terminal/line"
)

// Len returns the number of Lines currently stored in History.
func (h *History) Len() int {
	if h == nil || !h.valid {
		return 0
	}
	return int(h.size.Get())
}

// get returns the Line passed to the n'th previous call to Add.
func (h *History) get(n int) *line.Line {
	if h == nil || !h.valid || n < 0 || n >= int(h.size.Get()) {
		return nil
	}
	index := int(h.head.Get()) - n
	if index < 0 {
		index += limits.LinesPerHistory
	}
	return &h.line[index]
}

// Add appends the pending Line to History.
// If the History is filled to capacity, the oldest Line is discarded.
func (h *History) Add() {
	if h == nil || !h.valid {
		return
	}
	head := h.head.Get()
	// This copies the elements in line.Rune, and it copies the pointer fields by
	// value; i.e., each pointer itself is copied and not dereferenced.
	h.line[head] = h.pend
	// Now it is safe to modify pend without affecting its "snapshot" in h.

	head += 1
	head %= limits.LinesPerHistory
	h.head.Set(head)
	if size := h.size.Get(); size < limits.LinesPerHistory {
		h.size.Set(size + 1)
	}
	// Reset our History pointer
	h.indx.Set(0)
	// Reset the cursor, data, and I/O buffers.
	h.pend.LineFeed()
}

func (h *History) Back() {
	indx, size := h.indx.Get(), h.size.Get()
	if indx < size-1 {
		*h.get(int(indx)) = h.pend
		h.pend.Set(nil)
		indx++
		h.pend = *h.get(int(indx))

		h.pend.Flush()
		h.pend.MoveCursorTo(h.pend.Position())
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
		h.pend.MoveCursorTo(h.pend.Position())
		h.indx.Set(indx)
	}
}

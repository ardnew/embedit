//go:build !history
// +build !history

package history

import (
	"github.com/ardnew/embedit/terminal/line"
)

// Len returns the number of Lines currently stored in History.
func (h *History) Len() int {
	return 0
}

// Get returns the Line passed to the n'th previous call to Add.
// If n=0, the immediately previous Line is returned; if n=1, the Line before
// that, and so on.
// If n<0 or fewer than n+1 lines have been added, ok is false.
func (h *History) Get(n int) (ln *line.Line, ok bool) {
	return nil, false
}

// Add appends the pending Line to History.
// If the History is filled to capacity, the oldest Line is discarded.
func (h *History) Add() {
	if h == nil || !h.valid {
		return
	}
	// Reset the cursor, data, and I/O buffers.
	h.pend.LineFeed()
}

func (h *History) Back() {
}

func (h *History) Forward() {
}

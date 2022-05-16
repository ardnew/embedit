package embedit

import (
	"strings"

	"github.com/ardnew/embedit/fifo"
)

const NilRune = rune(0)

// Line represents a single line of input. It implements the fifo.Buffer
// interface for use as a fixed-length FIFO.
type Line []rune

// Len returns the size of the receiver.
func (rb *Line) Len() int {
	return len(*rb)
}

// Get returns a Data (of concrete type rune) at the given index and true.
// Returns nil and false if the index is out of bounds.
func (rb *Line) Get(i int) (data fifo.Data, ok bool) {
	if rb != nil && 0 <= i && i < rb.Len() {
		data, ok = (*rb)[i], true
	}
	return
}

// Set sets the rune element at the given index and returns true.
// Returns false if the index is out of bounds or given Data type is not rune.
func (rb *Line) Set(i int, data fifo.Data) (ok bool) {
	if rb != nil && 0 <= i && i < rb.Len() {
		// Don't write to the buffer unless type assertion succeeds.
		var c rune
		if c, ok = data.(rune); ok {
			(*rb)[i] = c
		}
	}
	return
}

func (rb *Line) String() string {
	if rb == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < rb.Len(); i++ {
		if (*rb)[i] == NilRune {
			sb.WriteRune('.')
		} else {
			sb.WriteRune((*rb)[i])
		}
	}
	sb.WriteString("]")
	return sb.String()
}

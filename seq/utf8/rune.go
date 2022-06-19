package utf8

import (
	"unicode/utf8"

	"github.com/ardnew/embedit/errors"
)

// Rune extends type rune with unbuffered implementations of several interfaces
// from standard library package "io".
type Rune rune

// invalid defines an invalid Rune.
//
// The conventional value utf8.RuneError is not used because it is considered
// a valid rune by package unicode/utf8 and returns 3-bytes from RuneLen.
// Using a proper invalid UTF-8 value allows us to catch invalid runes and
// strings in a sane manner without potentially using up precious bytes.
var invalid Rune = utf8.MaxRune + 1

// Rune returns r as a Go native rune type.
func (r *Rune) Rune() rune {
	if r == nil {
		return rune(invalid)
	}
	return rune(*r)
}

// Set sets r equal to a.
func (r *Rune) Set(a Rune) {
	if r == nil {
		return
	}
	*r = a
}

// SetRune sets r equal to the Go native rune a.
func (r *Rune) SetRune(a rune) {
	if r == nil {
		return
	}
	*r = Rune(a)
}

// Len returns the number of bytes required to encode r.
// Returns 0 if r is not a valid UTF-8 encoding.
func (r *Rune) Len() (n int) {
	if r == nil {
		return 0
	}
	if n = utf8.RuneLen(rune(*r)); n < 0 {
		return 0
	}
	return
}

// Equals returns true if and only if r and a are the same UTF-8 code point.
func (r *Rune) Equals(a Rune) bool {
	return r.EqualsRune(a.Rune())
}

// EqualsRune returns true if and only if r and a are the same UTF-8 code point.
func (r *Rune) EqualsRune(a rune) bool {
	if r == nil {
		return false
	}
	return rune(*r) == a
}

// IsError returns true if and only if r is nil or equal to RuneError.
func (r *Rune) IsError() bool {
	return r == nil || r.Equals(invalid)
}

// Encode writes into p the UTF-8 encoding of r and returns the number of bytes
// written. Returns 0, ErrWriteOverflow if p is not large enough to hold the
// encoding of r.
func (r *Rune) Encode(p []byte) (n int, err error) {
	if n = r.Len(); n == 0 {
		return 0, &errors.ErrInvalidReceiver
	}
	if p == nil {
		return 0, &errors.ErrInvalidArgument
	}
	pn := len(p)
	if n > pn {
		return 0, &errors.ErrWriteOverflow
	}
	return utf8.EncodeRune(p, rune(*r)), nil
}

// Read implements io.Reader.
func (r *Rune) Read(p []byte) (n int, err error) {
	return r.Encode(p)
}

// RunesLen returns the total number of bytes to encode each rune in p.
// Each rune with an invalid UTF-8 encoding is ignored, i.e., its rune length is
// assumed to be 0.
func RunesLen(p []Rune) (n int) {
	for _, r := range p {
		n += r.Len()
	}
	return
}

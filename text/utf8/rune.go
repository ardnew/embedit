package utf8

import "unicode/utf8"

// Rune extends type rune with unbuffered implementations of io.Copy interfaces.
type Rune rune

const RuneError Rune = utf8.RuneError

// RunesLen returns the total number of bytes to encode each rune in p.
// Each rune with an invalid UTF-8 encoding is ignored, i.e., its rune length is
// assumed to be 0.
func RunesLen(p []Rune) (n int) {
	for _, r := range p {
		n += r.Len()
	}
	return
}

// Rune returns r as native Go type rune.
func (r *Rune) Rune() rune {
	return rune(*r)
}

// Len returns the number of bytes required to encode r.
// Returns 0 if r is not a valid UTF-8 encoding.
func (r *Rune) Len() (n int) {
	u := rune(*r)
	if n = utf8.RuneLen(u); n < 0 {
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
	return rune(*r) == a
}

// IsError returns true if and only if r is equal to RuneError.
func (r *Rune) IsError() bool {
	return r.Equals(RuneError)
}

// Encode writes into p the UTF-8 encoding of r and returns the number of bytes
// written.
// Returns 0, ErrOverflow if p is not large enough to hold the encoding of r.
func (r *Rune) Encode(p []byte) (n int, err error) {
	if n = r.Len(); n == 0 {
		return 0, ErrReceiverEncode
	}
	if p == nil {
		return 0, ErrArgumentEncode
	}
	pn := len(p)
	if n > pn {
		return 0, ErrOverflowEncode
	}
	return utf8.EncodeRune(p, rune(*r)), nil
}

func (r *Rune) Read(p []byte) (n int, err error) {
	return r.Encode(p)
}

// Errors returned by Rune methods.
var (
	ErrReceiverEncode error
	ErrArgumentEncode error
	ErrOverflowEncode error
)

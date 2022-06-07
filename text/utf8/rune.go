package utf8

import "unicode/utf8"

// Rune extends type rune with unbuffered implementations of io.Copy interfaces.
type Rune rune

// RunesLen returns the total number of bytes to encode each rune in p.
// Each rune with an invalid UTF-8 encoding is ignored, i.e., its rune length is
// assumed to be 0.
func RunesLen(p []Rune) (n int) {
	for _, r := range p {
		n += r.Len()
	}
	return
}

// Len returns the number of bytes required to encode r.
// Returns 0 if r is not a valid UTF-8 encoding.
func (r Rune) Len() (n int) {
	u := rune(r)
	if n = utf8.RuneLen(u); n < 0 {
		return 0
	}
	return
}

// Encode writes into p the UTF-8 encoding of r and returns the number of bytes
// written.
// Returns 0, ErrOverflow if p is not large enough to hold the encoding of r.
func (r Rune) Encode(p []byte) (n int, err error) {
	if n = r.Len(); n == 0 {
		return 0, ErrReceiver("cannot Encode from invalid receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Encode into nil buffer")
	}
	pn := len(p)
	if n > pn {
		return 0, ErrOverflow("cannot Encode into undersized buffer")
	}
	return utf8.EncodeRune(p, rune(r)), nil
}

func (r Rune) Read(p []byte) (n int, err error) {
	return r.Encode(p)
}

// Types of errors returned by Rune methods.
type (
	ErrReceiver string
	ErrArgument string
	ErrOverflow string
)

func (e ErrReceiver) Error() string { return "rune [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "rune [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "rune [overflow]: " + string(e) }

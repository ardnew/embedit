package utf8

import "unicode/utf8"

type Rune rune

// Types of errors returned by Rune methods.
type (
	ErrReceiver string
	ErrOverflow string
)

func (r Rune) Read(p []byte) (n int, err error) {
	u := rune(r)
	if n = utf8.RuneLen(u); n <= 0 {
		return 0, ErrReceiver("cannot Read from invalid receiver")
	}
	pn := len(p)
	if n > pn {
		return 0, ErrOverflow("cannot Read into []byte (capacity < required)")
	}
	return utf8.EncodeRune(p, u), nil
}

func (e ErrReceiver) Error() string { return "rune [receiver]: " + string(e) }
func (e ErrOverflow) Error() string { return "rune [overflow]: " + string(e) }

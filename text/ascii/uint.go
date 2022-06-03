package ascii

import (
	"bytes"
	"io"
)

const host32bit = ^uint(0)>>32 == 0

type Uint32 uint32

// WriteUint32 writes the decimal string representation of u to w.
func WriteUint32(w io.Writer, u uint32) (n int, err error) {
	i, err := Uint32(u).WriteTo(w)
	return int(i), err
}

// Read copies the decimal string representation of u into p.
func (u Uint32) Read(p []byte) (n int, err error) {
	i, err := u.WriteTo(bytes.NewBuffer(p))
	return int(i), err
}

// WriteTo writes the decimal string representation of u to w.
func (u Uint32) WriteTo(w io.Writer) (n int64, err error) {
	switch { // Fast paths for small integers
	case u < 10:
		_, err = w.Write(digit[u : u+1])
		return 1, err
	case u < lutMax:
		i, err := w.Write(lut[u*2 : u*2+2])
		return int64(i), err
	}
	var a [10]byte // Maximum length of 32-bit decimal
	i := len(a)
	if host32bit {
		// Convert the lower digits using 32-bit operations
		for ; u >= 1e9; u /= 1e9 {
			r := u % 1e9
			for j := 4; j > 0; j-- {
				v := r % 100 * 2
				r /= 100
				i -= 2
				a[i+1] = lut[v+1]
				a[i+0] = lut[v+0]
			}
			i--
			a[i] = lut[r*2+1]
		}
	}
	for u >= 100 {
		v := u % 100 * 2
		u /= 100
		i -= 2
		a[i+1] = lut[v+1]
		a[i+0] = lut[v+0]
	}
	u *= 2
	i--
	a[i] = lut[u+1]
	if u >= 20 {
		i--
		a[i] = lut[u]
	}
	nw, errw := w.Write(a[i:])
	return int64(nw), errw
}

const lutMax = 100

var (
	digit = []byte(`0123456789`)
	lut   = []byte(`` +
		`00010203040506070809` +
		`10111213141516171819` +
		`20212223242526272829` +
		`30313233343536373839` +
		`40414243444546474849` +
		`50515253545556575859` +
		`60616263646566676869` +
		`70717273747576777879` +
		`80818283848586878889` +
		`90919293949596979899`)
)

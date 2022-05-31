package ascii

// Utoa returns the decimal string representation of u.
func Utoa(u uint32) string {
	return Appu(nil, u)
}

// Appu returns the decimal string representation of u appended to dst.
func Appu(dst []byte, u uint32) string {
	switch { // Fast paths for small integers
	case u < 10:
		return string(append(dst, "0123456789"[u:u+1]...))
	case u < lutMax:
		return string(append(dst, lut[u*2:u*2+2]...))
	}
	var a [10]byte // Maximum length of 32-bit decimal
	n := len(a)
	if host32bit {
		// Convert the lower digits using 32-bit operations
		for ; u >= 1e9; u /= 1e9 {
			r := u % 1e9
			for j := 4; j > 0; j-- {
				v := r % 100 * 2
				r /= 100
				n -= 2
				a[n+1] = lut[v+1]
				a[n+0] = lut[v+0]
			}
			n--
			a[n] = lut[r*2+1]
		}
	}
	for u >= 100 {
		v := u % 100 * 2
		u /= 100
		n -= 2
		a[n+1] = lut[v+1]
		a[n+0] = lut[v+0]
	}
	u *= 2
	n--
	a[n] = lut[u+1]
	if u >= 20 {
		n--
		a[n] = lut[u]
	}
	return string(append(dst, a[n:]...))
}

const host32bit = ^uint(0)>>32 == 0

const (
	lutMax = 100
	lut    = `` +
		`00010203040506070809` +
		`10111213141516171819` +
		`20212223242526272829` +
		`30313233343536373839` +
		`40414243444546474849` +
		`50515253545556575859` +
		`60616263646566676869` +
		`70717273747576777879` +
		`80818283848586878889` +
		`90919293949596979899`
)

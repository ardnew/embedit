package sequence

import (
	"io"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/volatile"
)

// Types of errors returned by Sequence methods.
type (
	ErrReceiver string
	ErrArgument string
	ErrOverflow string
)

// Sequence defines an I/O buffer for Terminal control/data byte sequences.
type Sequence struct {
	Byte  [config.BytesPerSequence]byte
	head  volatile.Register32
	tail  volatile.Register32
	valid bool
}

// Invalid represents an invalid Line.
var Invalid = Sequence{valid: false}

// Configure initializes the Sequence configuration.
func (s *Sequence) Configure() *Sequence {
	if s == nil {
		return nil
	}
	s.valid = false
	return s.init()
}

// init initializes the state of a configured Sequence.
func (s *Sequence) init() *Sequence {
	s.valid = true
	return s.Reset()
}

// Len returns the number of bytes in s.
func (s *Sequence) Len() int {
	if s == nil {
		return 0
	}
	return int(s.tail.Get() - s.head.Get())
}

// Reset sets the Sequence length to 0.
func (s *Sequence) Reset() *Sequence {
	if s == nil {
		return nil
	}
	s.head.Set(0)
	s.tail.Set(0)
	return s
}

// Read copies up to len(p) bytes from s to p and returns the number of bytes
// copied.
func (s *Sequence) Read(p []byte) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot Read from nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Read to nil buffer")
	}
	var ns int
	ns, n = s.Len(), len(p)
	if ns <= n {
		n, err = ns, io.EOF
	}
	h := s.head.Get()
	for i := range p[:n] {
		p[i] = s.Byte[h%config.BytesPerSequence]
		h++
	}
	if err == io.EOF {
		s.Reset()
	} else {
		s.head.Set(h)
	}
	return
}

// Write copies up to len(p) bytes from p to the start of s and returns the
// number of bytes copied.
//
// Write overwrites any bytes present, but stops writing once s is full.
func (s *Sequence) Write(p []byte) (n int, err error) {
	// Local buffer buf is used as the underlying array for long strings that can
	// cause heap allocation. We can take slices from this buffer to iteratively
	// build a desired string (via append builtin)

	if s == nil {
		return 0, ErrReceiver("cannot Write to nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Write from nil buffer")
	}
	n = copy(s.Reset().Byte[0:], p)
	s.tail.Set(uint32(n))
	if np := len(p); n < np {
		err = ErrOverflow("cannot Write entire buffer (truncated)")
	}
	return
}

// Append copies up to len(p) bytes from p to the end of s and returns the
// number of bytes copied.
//
// Unlike the append builtin, it does not extend the length of s to make room
// for all of p. It will only write to the free space in s and then return
// ErrOverflow if all of p could not be copied.
func (s *Sequence) Append(p []byte) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot Append to nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Append from nil buffer")
	}
	np := len(p)
	if np == 0 {
		return
	}
	if s.Len() == 0 {
		n = copy(s.Reset().Byte[0:], p)
		s.tail.Set(uint32(n))
	} else {
		h, t := s.head.Get(), s.tail.Get()
		for _, b := range p {
			if t-h >= config.BytesPerSequence {
				break
			}
			s.Byte[t%config.BytesPerSequence] = b
			t++
			n++
		}
		s.tail.Set(t)
	}
	if n < np {
		err = ErrOverflow("cannot Append entire buffer (truncated)")
	}
	return
}

func (e ErrReceiver) Error() string { return "sequence [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "sequence [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "sequence [overflow]: " + string(e) }

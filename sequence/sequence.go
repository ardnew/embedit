package sequence

import (
	"github.com/ardnew/embedit/encoding/ascii"
	"github.com/ardnew/embedit/limit"
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
	Byte  [limit.BytesPerSequence]byte
	size  volatile.Register32
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
	s.size.Set(0)
	return s.Reset()
}

// Len returns the number of bytes in a Sequence.
func (s *Sequence) Len() int {
	if s == nil {
		return 0
	}
	return int(s.size.Get())
}

// Reset clears all bytes and sets the Sequence length to 0.
func (s *Sequence) Reset() *Sequence {
	if s == nil {
		return nil
	}
	for i := range s.Byte[:s.size.Get()] {
		s.Byte[i] = 0
	}
	s.size.Set(0)
	return s
}

// Append copies up to len(a) bytes from a to the tail of a Sequence.
func (s *Sequence) Append(a []byte) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot Append to nil receiver")
	}
	if a == nil {
		return 0, ErrArgument("cannot Append from nil buffer")
	}
	i := s.Len()
	n = copy(s.Byte[i:], a)
	i += n
	s.size.Set(uint32(i))
	if c := len(a); n < c {
		err = ErrOverflow("limited Append to " + ascii.Utoa(uint32(n)) +
			" of " + ascii.Utoa(uint32(c)) + " total bytes")
	}
	return
}

// Write copies up to len(a) bytes from a to the head of a Sequence.
// Write overwrites any bytes present.
func (s *Sequence) Write(a []byte) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot Write to nil receiver")
	}
	if a == nil {
		return 0, ErrArgument("cannot Write from nil buffer")
	}
	n = copy(s.Byte[0:], a)
	for i := range s.Byte[n:s.Len()] {
		s.Byte[i] = 0 // Reset any trailing bytes.
	}
	s.size.Set(uint32(n))
	if c := len(a); n < c {
		err = ErrOverflow("limited Write to " + ascii.Utoa(uint32(n)) +
			" of " + ascii.Utoa(uint32(c)) + " total bytes")
	}
	return
}

func (e ErrReceiver) Error() string { return "sequence [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "sequence [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "sequence [overflow]: " + string(e) }

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
		return 0, ErrArgument("cannot Read into nil buffer")
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
// Returns ErrOverflow if s is full and all of p could not be copied.
func (s *Sequence) Write(p []byte) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot Write into nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Write from nil buffer")
	}
	n = copy(s.Reset().Byte[0:], p)
	s.tail.Set(uint32(n))
	if np := len(p); n < np {
		err = ErrOverflow("cannot Write entire buffer (short write)")
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
		return 0, ErrReceiver("cannot Append into nil receiver")
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
		err = ErrOverflow("cannot Append entire buffer (short write)")
	}
	return
}

func (s *Sequence) readFrom(r io.Reader, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to readFrom when the
	// elements of s will not be stored contiguously in the backing array.
	// We can do a brief sanity check on the indices to prevent A/V errors,
	// but no attempt is made to normalize or split the range into slices.
	if lo >= hi || lo < 0 || hi > config.BytesPerSequence {
		// The above condition implies implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, ErrArgument("cannot readFrom into invalid slice indices")
	}
	n, err = r.Read(s.Byte[lo:hi])
	// Extend the length of s by the number of bytes copied.
	s.tail.Set(s.tail.Get() + uint32(n))
	return
}

// ReadFrom copies bytes from r to s until all bytes have been copied or an
// error was encountered. Returns the number of bytes successfully copied.
//
// Bytes are copied directly without any buffering, so r and s must not overlap
// if both are implemented as buffers of physical memory.
//
// Calling io.Copy() with dst of type *Sequence will call dst.ReadFrom(src) to
// perform the copy instead of the io.Copy() implementation, which relies on a
// heap-allocated intermediate buffer.
func (s *Sequence) ReadFrom(r io.Reader) (n int64, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot ReadFrom into nil receiver")
	}
	if r == nil {
		return 0, ErrArgument("cannot ReadFrom from nil io.Reader")
	}
	h, t := s.head.Get(), s.tail.Get()
	if t-h == 0 {
		// Sequence is empty, ensure our indices are reset before writing across the
		// entire backing array.
		n0, err0 := s.Reset().readFrom(r, 0, config.BytesPerSequence)
		return int64(n0), err0
	}
	// Convert head and tail to physical array indices to determine if the used
	// elements span a contiguous region of memory in the backing array.
	ih, it := h%config.BytesPerSequence, t%config.BytesPerSequence
	// If the array indices are equal, with head not eqaul to tail (see above),
	// then the backing array is filled to capacity. We have nowhere to store the
	// bytes from r. We can either overwrite the existing Sequence or retain it
	// and return an error. Opting for the latter so that no bytes are lost, and
	// it gives the caller an opportunity to handle the situation.
	if it == ih {
		return 0, ErrOverflow("cannot ReadFrom into full receiver")
	}
	// Tail grows as elements are added to the ring buffer. Thus, if tail is less
	// than head, then the tail index has wrapped around after growing beyond the
	// backing array's high index (capacity-1), but the head index has not yet
	// wrapped around.
	if it > ih {
		// Tail has not overflowed its storage prior to head, which is the normal
		// case, and thus the unused elements potentially exist in two separate
		// contiguous regions of the backing array. The first region (1) spans from
		// the start of the array to the first-in (head) element, and the second
		// region (2) spans from the last-in (tail) element to the end of the array.
		var (
			n1, n2     int
			err1, err2 error
		)
		// (1.) Copy into start of the backing array to head (if range length > 0).
		if ih > 0 {
			if n1, err1 = s.readFrom(r, 0, int(ih)); err1 != nil {
				return int64(n1), err1
			}
		}
		// (2.) Copy into tail to end of the backing array (if range length > 0).
		if it < config.BytesPerSequence {
			n2, err2 = s.readFrom(r, int(it), config.BytesPerSequence)
		}
		return int64(n1 + n2), err2
	}
	// Unused elements form contiguous span in backing array from tail to head.
	nr, errr := s.readFrom(r, int(it), int(ih))
	return int64(nr), errr
}

func (s *Sequence) writeTo(w io.Writer, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to writeTo when the
	// elements of s are not stored contiguously in the backing array.
	// We can do a brief sanity check on the indices to prevent A/V errors,
	// but no attempt is made to normalize or split the range into slices.
	if lo >= hi || lo < 0 || hi > config.BytesPerSequence {
		// The above condition implies implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, ErrArgument("cannot writeTo from invalid slice indices")
	}
	n, err = w.Write(s.Byte[lo:hi])
	// Check if we copied all bytes, regardless of error.
	if n < s.Len() {
		// Short write; not all bytes were copied. Adjust head to refer to the
		// first byte that was not copied.
		s.head.Set(s.head.Get() + uint32(n))
	} else {
		// All bytes have been copied. Reset s to empty.
		_ = s.Reset()
	}
	return
}

// WriteTo copies bytes from s to w until all bytes have been copied or an error
// was encountered. Returns the number of bytes successfully copied.
//
// Bytes are copied directly without any buffering, so w and s must not overlap
// if both are implemented as buffers of physical memory.
//
// Calling io.Copy() with src of type *Sequence will call src.WriteTo(dst) to
// perform the copy instead of the io.Copy() implementation, which relies on a
// heap-allocated intermediate buffer.
func (s *Sequence) WriteTo(w io.Writer) (n int64, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot WriteTo from nil receiver")
	}
	if w == nil {
		return 0, ErrArgument("cannot WriteTo into nil io.Writer")
	}
	h, t := s.head.Get(), s.tail.Get()
	if t-h == 0 {
		// Sequence is empty, writing zero bytes to w.
		return 0, io.EOF
	}
	// Convert head and tail to physical array indices to determine if the used
	// elements span a contiguous region of memory in the backing array.
	ih, it := h%config.BytesPerSequence, t%config.BytesPerSequence
	// Tail grows as elements are added to the ring buffer. Thus, if tail is less
	// than head, then the tail index has wrapped around after growing beyond the
	// backing array's high index (capacity-1), but the head index has not yet
	// wrapped around.
	if it <= ih {
		// Tail has overflowed its storage prior to head, so the elements are not
		// contiguous in the backing array and/or the array is filled to capacity.
		// The first-in (head) elements are at the end of the array, and the last-in
		// (tail) elements are at the start of it. So we will potentially need to
		// copy the elements in two phases:
		// (1.) Copy from head to the end of the backing array.
		n1, err1 := s.writeTo(w, int(ih), config.BytesPerSequence)
		// If the number of bytes written equals the backing array's capacity, then
		// s was filled to capacity and is now empty; nothing to copy in phase 2.
		if err1 != nil || n1 == config.BytesPerSequence {
			return int64(n1), err1
		}
		// (2.) Copy from start of the backing array to tail.
		n2, err2 := s.writeTo(w, 0, int(it))
		return int64(n1 + n2), err2
	}
	// Elements form contiguous span in backing array from head to tail.
	nw, errw := s.writeTo(w, int(ih), int(it))
	return int64(nw), errw
}

func (s *Sequence) ReadByte() (b byte, err error) {
	if s == nil {
		return 0, ErrReceiver("cannot ReadByte from nil receiver")
	}
	h, t := s.head.Get(), s.tail.Get()
	if t-h == 0 {
		// Sequence is empty, reading zero bytes from s.
		_ = s.Reset()
		return 0, io.EOF
	}
	// Copy the byte in head position, and increment head by 1.
	b = s.Byte[h%config.BytesPerSequence]
	s.head.Set(h + 1)
	return
}

func (s *Sequence) WriteByte(b byte) (err error) {
	if s == nil {
		return ErrReceiver("cannot WriteByte into nil receiver")
	}
	h, t := s.head.Get(), s.tail.Get()
	if t-h == 0 {
		// Sequence is empty, we know what the resulting head and tail will be.
		s.Byte[0] = b
		s.head.Set(0)
		s.tail.Set(1)
		return nil
	}
	it := t % config.BytesPerSequence
	// If the array indices are equal, with head not eqaul to tail (see above),
	// then the backing array is filled to capacity. We have nowhere to store the
	// byte. We can either discard head or retain it and return an error. Opting
	// for the latter so that no byte is lost, and it gives the caller an
	// opportunity to handle the situation.
	if it == h%config.BytesPerSequence {
		return ErrOverflow("cannot WriteByte into full receiver")
	}
	// Write the byte into tail position and increment tail by 1.
	s.Byte[it] = b
	s.tail.Set(t + 1)
	return nil
}

func (e ErrReceiver) Error() string { return "sequence [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "sequence [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "sequence [overflow]: " + string(e) }

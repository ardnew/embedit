package sequence

import (
	"bytes"
	"io"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/sequence/eol"
	"github.com/ardnew/embedit/util"
	"github.com/ardnew/embedit/volatile"
)

// Sequence defines an I/O buffer for Terminal control/data byte sequences.
type Sequence struct {
	Byte  [config.BytesPerSequence]byte
	head  volatile.Register32
	tail  volatile.Register32
	mode  eol.Mode
	valid bool
}

// Configure initializes the Sequence configuration.
func (s *Sequence) Configure(mode eol.Mode) *Sequence {
	if s == nil {
		return nil
	}
	s.valid = false
	return s.init()
}

// init initializes the state of a configured Sequence.
func (s *Sequence) init() *Sequence {
	s.valid = true
	return s.reset()
}

// Len returns the number of bytes in s.
func (s *Sequence) Len() int {
	if s == nil {
		return 0
	}
	return int(s.tail.Get() - s.head.Get())
}

func (s *Sequence) reset() *Sequence {
	s.head.Set(0)
	s.tail.Set(0)
	return s
}

// Reset sets the Sequence length to 0.
func (s *Sequence) Reset() {
	if s == nil {
		return
	}
	_ = s.reset()
}

// Read copies up to len(p) unread bytes from s to p and returns the number of
// bytes copied.
func (s *Sequence) Read(p []byte) (n int, err error) {
	if s == nil || p == nil {
		return 0, io.ErrUnexpectedEOF
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
		s.reset()
	} else {
		s.head.Set(h)
	}
	return
}

// Write appends up to len(p) bytes from p to s and returns the number of bytes
// copied.
//
// Write will only write to the free space in s and then return ErrOverflow if
// all of p could not be copied.
func (s *Sequence) Write(p []byte) (n int, err error) {
	if s == nil || p == nil {
		return 0, io.ErrUnexpectedEOF
	}
	np := len(p)
	if np == 0 {
		// Source buffer is empty; there are no bytes from p to copy into s.
		// This isn't considered a Write error; it only means nothing was written.
		return
	}
	if s.Len() == 0 {
		n = copy(s.reset().Byte[0:], p)
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
		err = ErrOverflowWrite
	}
	return
}

func (s *Sequence) readFrom(r io.Reader, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to readFrom when the
	// elements of s will not be stored contiguously in the backing array.
	//
	// We can do a brief sanity check on the indices to prevent A/V errors, but no
	// attempt is made to normalize, split the range into slices, or verify the
	// range starts at tail and spans only the free-space region.
	if lo >= hi || lo < 0 || hi > config.BytesPerSequence {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, ErrArgumentReadFromRange
	}
	// Catch any attempt to return io.EOF and return nil instead.
	// See documentation on io.ReaderFrom, and io.Copy.
	n, err = util.EOFMask{Reader: r}.Read(s.Byte[lo:hi])
	// Extend the length of s by the number of bytes copied.
	s.tail.Set(s.tail.Get() + uint32(n))
	return
}

// ReadFrom copies bytes from r to s until all bytes have been read or an error
// was encountered. Returns the number of bytes successfully copied.
//
// A successful ReadFrom returns err == nil and not err == io.EOF.
// ReadFrom is defined to read from r until all bytes have been read (io.EOF),
// so it does not treat io.EOF from r as an error to be reported.
//
// Bytes are copied directly without any buffering, so r and s must not overlap
// if both are implemented as buffers of physical memory.
func (s *Sequence) ReadFrom(r io.Reader) (n int64, err error) {
	if s == nil {
		return 0, ErrReceiverReadFrom
	}
	if r == nil {
		return 0, ErrArgumentReadFrom
	}
	h, t := s.head.Get(), s.tail.Get()
	if h == t {
		// Sequence is empty, ensure our indices are reset before writing across the
		// entire backing array.
		n0, err0 := s.reset().readFrom(r, 0, config.BytesPerSequence)
		return int64(n0), err0
	}
	// Convert head and tail to physical array indices to determine if the used
	// elements span a contiguous region of memory in the backing array.
	ih, it := h%config.BytesPerSequence, t%config.BytesPerSequence
	// If the array indices are equal, with head not eqaul to tail (see above),
	// then the backing array is filled to capacity. We have nowhere to store the
	// bytes from r. We can either overwrite the existing Sequence or retain it
	// and return an error. Opting for the latter so that no bytes are lost, and
	// it gives the caller an opportunity to remedy the situation.
	if it == ih {
		return 0, ErrOverflowReadFrom
	}
	// When first-in (head) element is located at index 0 (i.e., the start of the
	// backing array), then the unused space spans exactly one region only;
	// namely, the range from last-in (tail) element to end of the array:
	//   (0123456789A) === Array index reference
	//   [HxxxT......]     Free-space forms contiguous span [4..A]
	if ih == 0 {
		nr, errr := s.readFrom(r, int(it), config.BytesPerSequence)
		return int64(nr), errr
	}
	// Tail grows as elements are added to the ring buffer. Thus, if tail is less
	// than head, then the tail index has wrapped around after growing beyond the
	// backing array's high index (capacity-1), but the head index has not yet
	// wrapped around.
	if it > ih {
		// Tail has not overflowed its storage prior to head, which is the normal
		// case, and thus the unused elements potentially exist in two separate
		// contiguous regions of the backing array. The first region (1) spans from
		// the last-in (tail) element to the end of the array, and the second
		// region (2) spans from the start of the array to the first-in (head)
		// element:
		//   (0123456789A) === Array index reference
		//   [...HxxxT...]     Free-space in region 1 [7..A] and region 2 [0..2]
		//   [......HxxxT]     Free-space in region 1 (A) and region 2 [0..5]
		var (
			n1, n2     int
			err1, err2 error
		)
		// (1.) Copy into tail to end of the backing array
		if n1, err1 = s.readFrom(r, int(it), config.BytesPerSequence); err1 != nil {
			return int64(n1), err1
		}
		// (2.) Copy into start of the backing array to head (if region length > 0).
		if ih > 0 {
			n2, err2 = s.readFrom(r, 0, int(ih))
		}
		return int64(n1 + n2), err2

	}
	// Unused elements form contiguous span in backing array from last-in (tail)
	// to first-in (head), which may include the entire front of the backing array
	// when tail is exactly equal to a multiple of array capacity.
	//   (0123456789A) === Array index reference
	//   [xxT......Hx]     Free-space forms contiguous span [2..8]
	//   [T......Hxxx]     Free-space forms contiguous span [0..6]
	nr, errr := s.readFrom(r, int(it), int(ih))
	return int64(nr), errr
}

func (s *Sequence) writeTo(w io.Writer, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to writeTo when the
	// elements of s are not stored contiguously in the backing array.
	//
	// We can do a brief sanity check on the indices to prevent A/V errors,
	// but no attempt is made to normalize or split the range into slices.
	if lo >= hi || lo < 0 || hi > config.BytesPerSequence {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, ErrArgumentWriteToRange
	}
	// Additional bytes written for EOL translation
	// var added int
	switch s.mode {
	case eol.LF:
		n, err = w.Write(s.Byte[lo:hi])
	case eol.CRLF, eol.CR:
		// We need to translate all LF bytes in our Sequence for the configured EOL.
		// Repeatedly write up to the next LF in the given range, then write our
		// configured EOL sequence, and repeat until the range has been covered.
		//
		// We will only return the number of bytes in our Sequence that were copied,
		// not the actual number of bytes written; e.g., "hi\n" contains 3 bytes,
		// but if EOL is CRLF, we will write "hi\r\n" (4 bytes), but return n=3.
		for {
			rem := hi - lo
			if rem <= 0 {
				break
			}
			off := bytes.IndexByte(s.Byte[lo:hi], '\n')
			if off >= 0 {
				rem = off
			}
			var no int
			no, err = w.Write(s.Byte[lo : lo+rem])
			n += no
			if err != nil {
				break
			}
			lo += rem
			if off >= 0 {
				if _, err = s.mode.WriteTo(w); err != nil {
					break
				}
				// Regardless of the number of bytes in our EOL, we are only consuming
				// a single LF byte from our Sequence. So to keep all of our counters
				// happy, we'll just act like we've written a single byte for EOL.
				n, lo = n+1, lo+1
			}
		}
	}
	// Check if we copied all bytes, regardless of error.
	if n < s.Len() {
		// Short write; not all bytes were copied. Adjust head to refer to the
		// first byte that was not copied.
		s.head.Set(s.head.Get() + uint32(n))
	} else {
		// All bytes have been copied. Reset s to empty.
		_ = s.reset()
	}
	return
}

// WriteTo copies bytes from s to w until all bytes have been written or an
// error was encountered. Returns the number of bytes successfully copied.
//
// Bytes are copied directly without any buffering, so w and s must not overlap
// if both are implemented as buffers of physical memory.
func (s *Sequence) WriteTo(w io.Writer) (n int64, err error) {
	if s == nil {
		return 0, ErrReceiverWriteTo
	}
	if w == nil {
		return 0, ErrArgumentWriteTo
	}
	h, t := s.head.Get(), s.tail.Get()
	if h == t {
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
		// The first region (1) spans from the first-in (head) element to the end of
		// the backing array, and the second region (2) spans from the start of the
		// array to the last-in (tail) element:
		//   (0123456789A) === Array index reference
		//   [xxT......Hx]     Elements in region 1 [9..A] and region 2 [0..1]
		//   [T......Hxxx]     Elements in region 1 [7..A] only
		//   [xxxxxHxxxxx]     Elements in region 1 [5..A] and region 2 [0..4]
		//   [Hxxxxxxxxxx]     Elements in region 1 [0..A] only
		// So we will potentially need to copy the elements in two phases:
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
	// Elements form contiguous span in backing array from first-in (head) element
	// to last-in (tail) element:
	//   (0123456789A) === Array index reference
	//   [HxxxT......]     Elements forms contiguous span [0..3]
	//   [......HxxxT]     Elements forms contiguous span [6..9]
	nw, errw := s.writeTo(w, int(ih), int(it))
	return int64(nw), errw
}

// ReadByte returns the next unread byte from s and a nil error.
// If s is empty, returns 0, io.EOF.
//
// To avoid ambiguous validity of the returned byte, ReadByte will always return
// either a valid byte and nil error, or an invalid byte and non-nil error.
// In particular, ReadByte never returns a byte read from s and error == io.EOF.
func (s *Sequence) ReadByte() (b byte, err error) {
	if s == nil {
		return 0, ErrReceiverReadByte
	}
	h, t := s.head.Get(), s.tail.Get()
	switch h {
	case t: // Reading zero bytes from s (empty), return io.EOF.
		_ = s.reset()
		return 0, io.EOF
	case t - 1: // Reading 1 and only byte in s, reset indices.
		_ = s.reset()
	default: // Reading 1 byte from s, reduce length by 1.
		s.head.Set(h + 1)
	}
	// Return the byte from original head position.
	return s.Byte[h%config.BytesPerSequence], nil
}

// WriteByte appends b to s and returns nil.
// If s is full, returns ErrOverflow.
func (s *Sequence) WriteByte(b byte) (err error) {
	if s == nil {
		return ErrReceiverWriteByte
	}
	h, t := s.head.Get(), s.tail.Get()
	if h == t {
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
	// opportunity to remedy the situation.
	if it == h%config.BytesPerSequence {
		return ErrOverflowWriteByte
	}
	// Write the byte into tail position and increment length by 1.
	s.Byte[it] = b
	s.tail.Set(t + 1)
	return nil
}

// WriteEOL appends the configured end of line sequence to s.
func (s *Sequence) WriteEOL() (n int, err error) {
	i, err := s.mode.WriteTo(s)
	return int(i), err
}

// Errors returned by Sequence methods.
var (
	ErrOverflowWrite         error
	ErrArgumentReadFromRange error
	ErrReceiverReadFrom      error
	ErrArgumentReadFrom      error
	ErrOverflowReadFrom      error
	ErrArgumentWriteToRange  error
	ErrReceiverWriteTo       error
	ErrArgumentWriteTo       error
	ErrReceiverReadByte      error
	ErrReceiverWriteByte     error
	ErrOverflowWriteByte     error
)

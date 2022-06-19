package seq

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/ardnew/embedit/config/limits"
	"github.com/ardnew/embedit/errors"
	"github.com/ardnew/embedit/seq/eol"
	"github.com/ardnew/embedit/seq/key"
	"github.com/ardnew/embedit/util"
	"github.com/ardnew/embedit/volatile"
)

// Buffer defines an I/O buffer for Terminal control/data byte sequences.
type Buffer struct {
	Byte  [limits.BytesPerBuffer]byte
	skey  [limits.MaxBytesPerKey]byte
	head  volatile.Register32
	tail  volatile.Register32
	mode  eol.Mode
	valid bool
}

// Configure initializes the Buffer configuration.
func (buf *Buffer) Configure(mode eol.Mode) *Buffer {
	if buf == nil {
		return nil
	}
	if buf.valid {
		// Configure must be called one time only.
		// Use object methods to modify configuration/state.
		return buf
	}
	buf.valid = false
	buf.mode = mode
	return buf.init()
}

// init initializes the state of a configured Buffer.
func (buf *Buffer) init() *Buffer {
	buf.valid = true
	return buf.reset()
}

// Len returns the number of bytes in buf.
func (buf *Buffer) Len() int {
	if buf == nil || !buf.valid {
		return 0
	}
	return int(buf.tail.Get() - buf.head.Get())
}

// Cap returns the byte capacity of buf.
func (buf *Buffer) Cap() int {
	if buf == nil || !buf.valid {
		return 0
	}
	return limits.BytesPerBuffer
}

func (buf *Buffer) reset() *Buffer {
	buf.head.Set(0)
	buf.tail.Set(0)
	return buf
}

// Reset sets the Buffer length to 0.
func (buf *Buffer) Reset() {
	if buf == nil || !buf.valid {
		return
	}
	_ = buf.reset()
}

// Read copies up to len(p) unread bytes from buf to p and returns the number of
// bytes copied.
func (buf *Buffer) Read(p []byte) (n int, err error) {
	if buf == nil || !buf.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	if p == nil {
		return 0, &errors.ErrInvalidArgument
	}
	var ns int
	ns, n = buf.Len(), len(p)
	if ns <= n {
		n, err = ns, io.EOF
	}
	h := buf.head.Get()
	for i := range p[:n] {
		p[i] = buf.Byte[h%limits.BytesPerBuffer]
		h++
	}
	if err == io.EOF {
		buf.reset()
	} else {
		buf.head.Set(h)
	}
	return
}

// Write appends up to len(p) bytes from p to buf and returns the number of
// bytes copied.
//
// Write will only write to the free space in buf and then return
// ErrWriteOverflow if all of p could not be copied.
func (buf *Buffer) Write(p []byte) (n int, err error) {
	if buf == nil || !buf.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	if p == nil {
		return 0, &errors.ErrInvalidArgument
	}
	np := len(p)
	if np == 0 {
		// Source buffer is empty; there are no bytes from p to copy into buf.
		// This isn't considered a Write error; it only means nothing was written.
		return
	}
	if buf.Len() == 0 {
		n = copy(buf.reset().Byte[0:], p)
		buf.tail.Set(uint32(n))
	} else {
		h, t := buf.head.Get(), buf.tail.Get()
		for _, b := range p {
			if t-h >= limits.BytesPerBuffer {
				break
			}
			buf.Byte[t%limits.BytesPerBuffer] = b
			t++
			n++
		}
		buf.tail.Set(t)
	}
	if n < np {
		err = &errors.ErrWriteOverflow
	}
	return
}

func (buf *Buffer) readFrom(r io.Reader, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to readFrom when the
	// elements of buf will not be stored contiguously in the backing array.
	//
	// We can do a brief sanity check on the indices to prevent A/V errors, but no
	// attempt is made to normalize, split the range into slices, or verify the
	// range starts at tail and spans only the free-space region.
	if lo >= hi || lo < 0 || hi > limits.BytesPerBuffer {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, &errors.ErrOutOfRange
	}
	// Catch any attempt to return io.EOF and return nil instead.
	// See documentation on io.ReaderFrom, and io.Copy.
	n, err = util.EOFMask{Reader: r}.Read(buf.Byte[lo:hi])
	// Extend the length of buf by the number of bytes copied.
	buf.tail.Set(buf.tail.Get() + uint32(n))
	return
}

// ReadFrom copies bytes from r to buf until all bytes have been read or an
// error was encountered. Returns the number of bytes successfully copied.
//
// A successful ReadFrom returns err == nil and not err == io.EOF.
// ReadFrom is defined to read from r until all bytes have been read (io.EOF),
// so it does not treat io.EOF from r as an error to be reported.
//
// Bytes are copied directly without any buffering, so r and buf must not
// overlap if both are implemented as buffers of physical memory.
func (buf *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	if buf == nil || !buf.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	if r == nil {
		return 0, &errors.ErrInvalidArgument
	}
	h, t := buf.head.Get(), buf.tail.Get()
	if h == t {
		// Buffer is empty, ensure our indices are reset before writing across the
		// entire backing array.
		n0, err0 := buf.reset().readFrom(r, 0, limits.BytesPerBuffer)
		return int64(n0), err0
	}
	// Convert head and tail to physical array indices to determine if the used
	// elements span a contiguous region of memory in the backing array.
	ih, it := h%limits.BytesPerBuffer, t%limits.BytesPerBuffer
	// If the array indices are equal, with head not eqaul to tail (see above),
	// then the backing array is filled to capacity. We have nowhere to store the
	// bytes from r. We can either overwrite the existing Buffer or retain it and
	// return an error. Opting for the latter so that no bytes are lost, and it
	// gives the caller an opportunity to remedy the situation.
	if it == ih {
		return 0, &errors.ErrReadOverflow
	}
	// When first-in (head) element is located at index 0 (i.e., the start of the
	// backing array), then the unused space spans exactly one region only;
	// namely, the range from last-in (tail) element to end of the array:
	//   (0123456789A) === Array index reference
	//   [HxxxT......]     Free-space forms contiguous span [4..A]
	if ih == 0 {
		nr, errr := buf.readFrom(r, int(it), limits.BytesPerBuffer)
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
		if n1, err1 = buf.readFrom(r, int(it), limits.BytesPerBuffer); err1 != nil {
			return int64(n1), err1
		}
		// (2.) Copy into start of the backing array to head (if region length > 0).
		if ih > 0 {
			n2, err2 = buf.readFrom(r, 0, int(ih))
		}
		return int64(n1 + n2), err2

	}
	// Unused elements form contiguous span in backing array from last-in (tail)
	// to first-in (head), which may include the entire front of the backing array
	// when tail is exactly equal to a multiple of array capacity.
	//   (0123456789A) === Array index reference
	//   [xxT......Hx]     Free-space forms contiguous span [2..8]
	//   [T......Hxxx]     Free-space forms contiguous span [0..6]
	nr, errr := buf.readFrom(r, int(it), int(ih))
	return int64(nr), errr
}

func (buf *Buffer) writeTo(w io.Writer, lo, hi int) (n int, err error) {
	// The caller is responsible for coordinating calls to writeTo when the
	// elements of buf are not stored contiguously in the backing array.
	//
	// We can do a brief sanity check on the indices to prevent A/V errors,
	// but no attempt is made to normalize or split the range into slices.
	if lo >= hi || lo < 0 || hi > limits.BytesPerBuffer {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return 0, &errors.ErrOutOfRange
	}
	// Additional bytes written for EOL translation
	// var added int
	switch buf.mode {
	case eol.LF:
		n, err = w.Write(buf.Byte[lo:hi])
	case eol.CRLF, eol.CR:
		// We need to translate all LF bytes in our Buffer for the configured EOL.
		// Repeatedly write up to the next LF in the given range, then write our
		// configured EOL sequence, and repeat until the range has been covered.
		//
		// We will only return the number of bytes in our Buffer that were copied,
		// not the actual number of bytes written; e.g., "hi\n" contains 3 bytes,
		// but if EOL is CRLF, we will write "hi\r\n" (4 bytes) and return n=3.
		for {
			rem := hi - lo
			if rem <= 0 {
				break
			}
			off := bytes.IndexByte(buf.Byte[lo:hi], '\n')
			if off >= 0 {
				rem = off
			}
			var no int
			no, err = w.Write(buf.Byte[lo : lo+rem])
			n += no
			if err != nil {
				break
			}
			lo += rem
			if off >= 0 {
				if _, err = buf.mode.WriteTo(w); err != nil {
					break
				}
				// Regardless of the number of bytes in our EOL, we are only consuming
				// a single LF byte from our Buffer. So to keep all of our counters
				// happy, we'll just act like we've written a single byte for EOL.
				n, lo = n+1, lo+1
			}
		}
	}
	// Check if we copied all bytes, regardless of error.
	if n < buf.Len() {
		// Short write; not all bytes were copied. Adjust head to refer to the
		// first byte that was not copied.
		buf.head.Set(buf.head.Get() + uint32(n))
	} else {
		// All bytes have been copied. Reset buf to empty.
		_ = buf.reset()
	}
	return
}

// WriteTo copies bytes from buf to w until all bytes have been written or an
// error was encountered. Returns the number of bytes successfully copied.
//
// Bytes are copied directly without any buffering, so w and buf must not
// overlap if both are implemented as buffers of physical memory.
func (buf *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	if buf == nil || !buf.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	if w == nil {
		return 0, &errors.ErrInvalidArgument
	}
	h, t := buf.head.Get(), buf.tail.Get()
	if h == t {
		// Buffer is empty, writing zero bytes to w.
		return 0, io.EOF
	}
	// Convert head and tail to physical array indices to determine if the used
	// elements span a contiguous region of memory in the backing array.
	ih, it := h%limits.BytesPerBuffer, t%limits.BytesPerBuffer
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
		n1, err1 := buf.writeTo(w, int(ih), limits.BytesPerBuffer)
		// If the number of bytes written equals the backing array's capacity, then
		// buf was filled to capacity and is now empty; nothing to copy in phase 2.
		if err1 != nil || n1 == limits.BytesPerBuffer {
			return int64(n1), err1
		}
		// (2.) Copy from start of the backing array to tail.
		n2, err2 := buf.writeTo(w, 0, int(it))
		return int64(n1 + n2), err2
	}
	// Elements form contiguous span in backing array from first-in (head) element
	// to last-in (tail) element:
	//   (0123456789A) === Array index reference
	//   [HxxxT......]     Elements forms contiguous span [0..3]
	//   [......HxxxT]     Elements forms contiguous span [6..9]
	nw, errw := buf.writeTo(w, int(ih), int(it))
	return int64(nw), errw
}

// ReadByte returns the next unread byte from buf and a nil error.
// If buf is empty, returns 0, io.EOF.
//
// To avoid ambiguous validity of the returned byte, ReadByte will always return
// either a valid byte and nil error, or an invalid byte and non-nil error.
// In particular, ReadByte never returns a byte read along with error == io.EOF.
func (buf *Buffer) ReadByte() (b byte, err error) {
	if buf == nil || !buf.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	h, t := buf.head.Get(), buf.tail.Get()
	switch h {
	case t: // Reading zero bytes from buf (empty), return io.EOF.
		_ = buf.reset()
		return 0, io.EOF
	case t - 1: // Reading 1 and only byte in buf, reset indices.
		_ = buf.reset()
	default: // Reading 1 byte from buf, reduce length by 1.
		buf.head.Set(h + 1)
	}
	// Return the byte from original head position.
	return buf.Byte[h%limits.BytesPerBuffer], nil
}

// WriteByte appends b to buf and returns nil.
// If buf is full, returns ErrWriteOverflow.
func (buf *Buffer) WriteByte(b byte) (err error) {
	if buf == nil || !buf.valid {
		return &errors.ErrInvalidReceiver
	}
	h, t := buf.head.Get(), buf.tail.Get()
	if h == t {
		// Buffer is empty, we know what the resulting head and tail will be.
		buf.Byte[0] = b
		buf.head.Set(0)
		buf.tail.Set(1)
		return nil
	}
	it := t % limits.BytesPerBuffer
	// If the array indices are equal, with head not eqaul to tail (see above),
	// then the backing array is filled to capacity. We have nowhere to store the
	// byte. We can either discard head or retain it and return an error. Opting
	// for the latter so that no byte is lost, and it gives the caller an
	// opportunity to remedy the situation.
	if it == h%limits.BytesPerBuffer {
		return &errors.ErrWriteOverflow
	}
	// Write the byte into tail position and increment length by 1.
	buf.Byte[it] = b
	buf.tail.Set(t + 1)
	return nil
}

// WriteEOL appends the configured end of line sequence to buf.
func (buf *Buffer) WriteEOL() (n int, err error) {
	i, err := buf.mode.WriteTo(buf)
	return int(i), err
}

// Parse tries to parse a key sequence from buf.
// If successful, it returns the key r and its size n in bytes.
// Otherwise, it returns key.Error and n=0.
//
// Parse consumes the bytes that contribute to the returned key r.
// If an entire sequence could not be parsed, no bytes are consumed.
func (buf *Buffer) Parse(isPasting bool) (r rune, n int) {
	h, t := buf.head.Get(), buf.tail.Get()
	size := t - h // Number of bytes currently in buf.
	if size == 0 {
		return key.Error, 0
	}
	// Size is the minimum among:
	//   a.) the number bytes in buf.Byte; or
	//   b.) the maximum length of a key sequence (cap(buf.skey)).
	if size > limits.MaxBytesPerKey {
		size = limits.MaxBytesPerKey
	}
	// Copy our source bytes to the temporary buffer []skey so that we can decode
	// it via package "unicode/utf8" without alignment/overflow problems due to
	// our backing array being a circular FIFO.
	for i := uint32(0); i < size; i++ {
		buf.skey[i] = buf.Byte[(h+i)%limits.BytesPerBuffer]
	}
	for i := size; i < limits.MaxBytesPerKey; i++ {
		buf.skey[i] = 0 // Zero out remaining bytes in []skey.
	}
	// Be sure to consume the bytes parsed into r.
	defer func(b *Buffer, head uint32) {
		if n > 0 {
			b.head.Set(head + uint32(n))
		}
	}(buf, h)

	if !isPasting {
		// UTF-8 control codes (ASCII)
		switch buf.skey[0] {
		case key.CtrlA:
			return key.Home, 1
		case key.CtrlB:
			return key.Left, 1
		case key.CtrlE:
			return key.End, 1
		case key.CtrlF:
			return key.Right, 1
		case key.CtrlH:
			return key.Backspace, 1
		case key.CtrlK:
			return key.DeleteTrailing, 1
		case key.CtrlL:
			return key.ClearScreen, 1
		case key.CtrlN:
			return key.Down, 1
		case key.CtrlP:
			return key.Up, 1
		case key.CtrlU:
			return key.DeleteLeading, 1
		case key.CtrlW:
			return key.DeleteWord, 1
		}
	}
	// UTF-8 runes
	if buf.skey[0] != key.Escape {
		if !utf8.FullRune(buf.skey[0:]) {
			return key.Error, 0
		}
		return utf8.DecodeRune(buf.skey[0:])
	}
	// ANSI escape sequences
	if bytes.HasPrefix(buf.skey[0:], key.CSI) {
		if isPasting {
			if bytes.HasPrefix(buf.skey[0:], key.EOP) {
				return key.PasteEnd, len(key.EOP)
			}
		} else {
			switch size {

			case 3:
				// xterm sequences
				switch buf.skey[2] {
				case 'A':
					return key.Up, 3
				case 'B':
					return key.Down, 3
				case 'C':
					return key.Right, 3
				case 'D':
					return key.Left, 3
				case 'H':
					return key.Home, 3
				case 'F':
					return key.End, 3
				}
			case 4:
				// vt sequences
				if buf.skey[3] == '~' {
					switch buf.skey[2] {
					case '1':
						return key.Home, 4
					case '2':
						return key.Insert, 4
					case '3':
						return key.Delete, 4
					case '4':
						return key.End, 4
					case '5':
						return key.PageUp, 4
					case '6':
						return key.PageDown, 4
					case '7':
						return key.Home, 4
					case '8':
						return key.End, 4
					}
				}
			case 5:
				// vt sequences
				if buf.skey[4] == '~' {
					switch buf.skey[2] {
					case '1':
						switch buf.skey[3] {
						case '0':
							return key.F0, 5
						case '1':
							return key.F1, 5
						case '2':
							return key.F2, 5
						case '3':
							return key.F3, 5
						case '4':
							return key.F4, 5
						case '5':
							return key.F5, 5
						case '7':
							return key.F6, 5
						case '8':
							return key.F7, 5
						case '9':
							return key.F8, 5
						}
					case '2':
						switch buf.skey[3] {
						case '0':
							return key.F9, 5
						case '1':
							return key.F10, 5
						case '3':
							return key.F11, 5
						case '4':
							return key.F12, 5
						case '5':
							return key.F13, 5
						case '6':
							return key.F14, 5
						case '8':
							return key.F15, 5
						case '9':
							return key.F16, 5
						}
					case '3':
						switch buf.skey[3] {
						case '1':
							return key.F17, 5
						case '2':
							return key.F18, 5
						case '3':
							return key.F19, 5
						case '4':
							return key.F20, 5
						}
					}
				}
			case 6:
				// xterm sequences
				if buf.skey[2] == '1' && buf.skey[3] == ';' && buf.skey[4] == '3' {
					switch buf.skey[5] {
					case 'C':
						return key.AltRight, 6
					case 'D':
						return key.AltLeft, 6
					}
				}
			}
			if bytes.HasPrefix(buf.skey[0:], key.SOP) {
				return key.PasteStart, len(key.SOP)
			}
		}
	}

	// If we get here then we have a key that we don't recognise, or a partial
	// sequence.
	// It's not clear how one should find the end of a sequence without knowing
	// them all, but it seems [a-zA-Z~] only appears at the end of a sequence.
	for i, c := range buf.skey[0:] {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '~' {
			return key.Unknown, i + 1
		}
	}
	return key.Error, 0
}

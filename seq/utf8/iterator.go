package utf8

// Iterator defines a list of Rune randomly accessible by sequential index and
// terminated by head and tail indices.
//
// This interface provides a []Rune accessor abstraction for structures that
// store Rune-like data in some other type such as the native Go type []rune.
type Iterator interface {
	// RuneHead returns the index of the first Rune element.
	RuneHead() uint32
	// RuneTail returns the index of the last Rune element.
	RuneTail() uint32
	// RuneAt is the normal array-like accessor that returns a Rune for a given
	// 0-based list index.
	RuneAt(i int) *Rune
}

// IterableRune implements Iterator using the native Go type []rune.
//
// Use a construct like the following to convert an existing []rune to Iterable
// without causing a copy/alloc:
//
//	  var aSlice = []rune{...} // Some global
//	  ...
//		   it := Iterable{Iterator: (*IterableRune)(&aSlice)}
type IterableRune []rune

// RuneHead returns 0.
// Implements Iterator for native Go type []rune.
func (ir *IterableRune) RuneHead() uint32 { return 0 }

// RuneTail returns len(*ir).
// Implements Iterator for native Go type []rune.
func (ir *IterableRune) RuneTail() uint32 { return uint32(len(*ir)) }

// RuneAt returns a pointer to (*ir)[i].
// Implements Iterator for native Go type []rune.
func (ir *IterableRune) RuneAt(i int) *Rune { return (*Rune)(&(*ir)[i]) }

var NonIterableRune IterableRune

// Iterable defines a concrete implementation of an Iterator that provides
// type-agnostic methods over the interface.
type Iterable struct {
	Iterator
	pos uint32 // Current element index (1st slice index, e.g., slice[pos:...])
	end uint32 // Last element index +1 (2nd slice index, e.g., slice[...:end])
}

// Reset resets the internal indices of s based on the backing Iterator's
// current head and tail indices.
func (s *Iterable) Reset() *Iterable {
	if s == nil || s.Iterator == nil {
		return nil
	}
	s.pos = s.RuneHead()
	s.end = s.RuneTail()
	return s
}

// Slice sets the internal indices of s based on the backing Iterator's current
// head and tail indices offset by the given lo and hi slice indices.
//
// Both lo and hi are relative to head and tail of s, such that lo=0 always
// refers to s's head index (even if s was previously sliced).
//
// If lo and/or hi are negative, they are treated as unspecified slice indices.
// For example, Slice(-1, N) is equivalent to s[:N], Slice(N, -1) is to s[N:],
// and Slice(-1, -1) is to s[:] (also equivalent to Reset).
func (s *Iterable) Slice(lo, hi int) *Iterable {
	if s.Reset() == nil || s.pos == s.end {
		return nil // Invalid or empty receiver
	}
	if lo < 0 { // From 0 to hi-1
		lo = 0
	}
	if hi < 0 { // From lo to length-1
		hi = int(s.end - s.pos)
	}
	if lo >= hi || uint32(hi) > s.end-s.pos {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		// Don't modify s, just return nil to indicate an invalid slice.
		return nil
	}
	s.end = s.pos + uint32(hi)
	s.pos = s.pos + uint32(lo)
	return s
}

// Next returns the next Rune in s.
//
// If there are no elements remaining in s, returns a Rune r such that
// r.IsError() == true and r.Len() == 0.
func (s *Iterable) Next() (r *Rune) {
	if s == nil || s.Iterator == nil || s.pos == s.end {
		return &invalid
	}
	r = s.RuneAt(int(s.pos))
	s.pos++
	return
}

// Len scans the current range and counts the number of bytes required to encode
// each valid rune. Rune encodings that are invalid UTF-8 are considered to have
// zero bytes. Unlike GlyphCount, runes in escape sequences are included.
//
// Once scanning completes, the receiver's internal indices are reset to their
// original value from when the method was called.
func (s *Iterable) Len() (n int) {
	if s == nil || s.Iterator == nil {
		return
	}
	// Capture head/tail and restore after scanning
	defer func(p *Iterable, h, t uint32) {
		p.pos = h
		p.end = t
	}(s, s.pos, s.end)
	for s.pos != s.end {
		n += s.Next().Len()
	}
	return
}

// GlyphCount scans the current range and counts the number of runes that
// are not within any escape sequence.
//
// Once scanning completes, the receiver's internal indices are reset to their
// original value from when the method was called.
//
// Note that only those runes that are in escape sequences which begin at or
// after the Iterator's first element (at RuneHead) will be excluded from the
// count. An example of this limitation is shown and discussed below.
//
//	Three different Iterators over the same backing array ([9]rune) are shown:
//
//	 [ 'H', 'l', 'o', ESC, '[', '2', 'D', 'e', 'l' ]   // Backing array
//	  ==== ==== ==== ____ ____ ____ ____ ==== ====
//	 { +1   +2   +3   --   --   --   --   +4   +5  }   // (1.) 5 glyphs
//	 { +1   +2   +3   --   -- }                        // (2.) 3 glyphs
//	                          { +1   +2   +3   +4  }   // (3.) 4 glyphs
//
//	The second and third Iterators together form the same sequence as the first
//	Iterator, so their total number of glyphs (3 + 4) should logically equal
//	the first (5).
//
//	However, because the second and third Iterators' bounds were not aligned
//	with the escape sequence's bounds, the final 2 runes of the escape sequence
//	was erroneously counted as ordinary runes in the third Iterator.
func (s *Iterable) GlyphCount() (count int) {
	// Capture head/tail and restore after scanning
	defer func(p *Iterable, h, t uint32) {
		p.pos = h
		p.end = t
	}(s, s.pos, s.end)
	var esc, err bool
	for !err {
		switch r := s.Next(); {
		case r == nil || *r == invalid: // (TBD) Precedes case esc! Is this right?
			err = true
		case esc:
			esc = (*r < 'a' || 'z' < *r) && (*r < 'A' || 'Z' < *r)
		case *r == 0x1B: // ESC (27)
			esc = true
		default:
			count++
		}
	}
	return
}

// Apply scans the current range and evaluates the given function fn with each
// rune as argument. If fn returns false for any rune, Apply returns false
// immediately. Otherwise, fn returned true for all runes, Apply returns true.
// If the Iterable range contains zero runes, Apply returns true.
//
// Once scanning completes, the receiver Iterable's head and tail are restored
// to their original value when the method was entered.
// func (s *Iterable) Apply(fn func(*Rune) bool) (ok bool) {
// 	if s == nil || s.Iterator == nil || fn == nil {
// 		return
// 	}
// 	// Capture head/tail and restore after scanning
// 	defer func(p *Iterable, h, t uint32) {
// 		p.pos = h
// 		p.end = t
// 	}(s, s.pos, s.end)
// 	for s.pos != s.end {
// 		if !fn(s.Next()) {
// 			return
// 		}
// 	}
// 	return true // Even if the Iterable has 0 elements
// }

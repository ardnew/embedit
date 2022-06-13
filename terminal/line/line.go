package line

import (
	"bytes"
	"io"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/errors"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/key"
	"github.com/ardnew/embedit/terminal/wire"
	"github.com/ardnew/embedit/text/utf8"
	"github.com/ardnew/embedit/volatile"
)

// Line represents a single line of input.
type Line struct {
	curs  *cursor.Cursor
	ctrl  *wire.Control
	disp  *display.Display
	Rune  [config.RunesPerLine]utf8.Rune
	posi  volatile.Register32
	head  volatile.Register32
	tail  volatile.Register32
	valid bool
}

// Configure initializes the Line configuration.
func (l *Line) Configure(curs *cursor.Cursor) *Line {
	if l == nil {
		return nil
	}
	l.valid = false
	l.curs = curs
	return l.init()
}

// init initializes the state of a configured Line.
func (l *Line) init() *Line {
	l.valid = true
	if ctrl := l.curs.Control(); ctrl != nil {
		l.ctrl = ctrl
	}
	if disp := l.curs.Display(); disp != nil {
		l.disp = disp
	}
	return l.Reset()
}

// Reset sets the Line length to 0 and resets the cursor position.
func (l *Line) Reset() *Line {
	if l == nil {
		return nil
	}
	l.posi.Set(0)
	l.head.Set(0)
	l.tail.Set(0)
	return l
}

// Len returns the number of bytes in l.
func (l *Line) Len() (n int) {
	if l == nil {
		return 0
	}
	ih := l.head.Get() % config.RunesPerLine
	it := l.tail.Get() % config.RunesPerLine
	if ih > it {
		return utf8.RunesLen(l.Rune[ih:]) + utf8.RunesLen(l.Rune[:it])
	}
	return utf8.RunesLen(l.Rune[ih:it])
}

// RuneCount returns the total number of runes in l.
func (l *Line) RuneCount() (count int) {
	if l == nil {
		return 0
	}
	return int(l.tail.Get() - l.head.Get())
}

// RuneAt returns the Rune at index i in l. Implements utf8.Iterator.
func (l *Line) RuneAt(i int) *utf8.Rune { return &l.Rune[i%config.RunesPerLine] }

// RuneHead returns the head index of l. Implements utf8.Iterator.
func (l *Line) RuneHead() uint32 { return l.head.Get() }

// RuneTail returns the tail index of l. Implements utf8.Iterator.
func (l *Line) RuneTail() uint32 { return l.tail.Get() }

// glyphCount returns the number of runes in l starting at k and ending at k+n-1
// that are not part of an escape sequence.
// If n is negative, returns the number of unescaped runes in l starting at k.
//
// See type key.UnescRuneCount for more details.
//
// The value k = 0 always refers to l.head, i.e., the first rune of l; it does
// not refer to index 0 of the backing array. It is the responsibility of this
// method to compute the appropriate offsets into the circular FIFO and account
// for possible wraparound based on l's current head and tail.
// The caller must not account for these offsets. Otherwise, incorrect indexing
// caused by double correction will almost always occur.
func (l *Line) glyphCount(k, n int) (count int) {
	if s := (utf8.Iterable{Iterator: l}); s.Slice(k, k+n) {
		return s.GlyphCount()
	}
	return
}

// promptGlyphCount returns the number of runes in the user input prompt that
// are not part of an escape sequence.
func (l *Line) promptGlyphCount() (count int) {
	if s := (utf8.Iterable{Iterator: l.disp.PromptIterator()}); s.Reset() {
		return s.GlyphCount()
	}
	return
}

// countToLeftWord returns then number of places from the cursor to the start of
// the previous word.
func (l *Line) countToLeftWord() int {
	origPos := l.position()
	if origPos == 0 {
		return 0
	}
	head := int(l.head.Get())
	pos := origPos - 1
	for pos > 0 {
		if !l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos--
	}
	for pos > 0 {
		if l.RuneAt(head + pos).Equals(' ') {
			pos++
			break
		}
		pos--
	}
	return origPos - pos
}

// countToRightWord returns then number of places from the cursor to the start
// of the next word.
func (l *Line) countToRightWord() int {
	origPos := l.position()
	head := int(l.head.Get())
	eol := l.RuneCount()
	pos := origPos
	for pos < eol {
		if l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos++
	}
	for pos < eol {
		if !l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos++
	}
	return pos - origPos
}

// ErasePrevious erases up to n previous runes from the current cursor position.
// Retained trailing runes are moved left in place of the runes erased.
//
// Appends sequences to the output buffer for repositioning the cursor and
// overwriting the portion of text that changed.
func (l *Line) ErasePrevious(n int) (err error) {
	if l == nil {
		return io.ErrUnexpectedEOF
	}
	if n <= 0 {
		return
	}
	pos := l.position()
	if pos < n {
		n = pos
	}
	pos -= n
	if err = l.moveCursorTo(pos); err != nil {
		return err
	}
	// Overwrite leading runes with trailing runes
	h, t := l.head.Get(), l.tail.Get()-uint32(n)
	if hs, s := h, (utf8.Iterable{Iterator: l}); s.Slice(pos+n, -1) {
		for {
			r := s.Next()
			if r.IsError() {
				break
			}
			l.RuneAt(int(hs + uint32(pos))).Set(*r)
			hs++
		}
	}
	// Erase the trailing runes with spaces
	for i := 0; i < n; i++ {
		l.RuneAt(int(t + uint32(i))).Set(key.Space)
	}
	if l.disp.Echo() {
		// Temporarily adjust head to rewrite only the changed portion of text.
		l.head.Set(h + uint32(pos))
		// Write out the text right-of the deletion, including the 0x20 erasors.
		err = l.flush()
		// Reset head back to the actual beginning of the line.
		l.head.Set(h)
	}
	// Finally truncate tail, set final cursor position, and flush output buffer.
	l.tail.Set(t)
	if e := l.moveCursorTo(pos); err == nil && e != nil {
		err = e
	}
	return
}

// kill appends a line-kill escape sequence to the output buffer that clears the
// line from the current cursor position to the end of the line.
func (l *Line) kill() (err error) {
	l.ctrl.Out.WriteByte(key.Escape)
	l.ctrl.Out.WriteByte('[')
	l.ctrl.Out.WriteByte('K')
	_, err = l.ctrl.Flush()
	return
}

// insert inserts the given key at the current cursor position in l.
func (l *Line) Insert(key rune) (err error) {
	if l == nil {
		return &errors.ErrInvalidReceiver
	}
	h, t := l.head.Get(), l.tail.Get()
	if t-h >= config.RunesPerLine {
		return io.ErrShortWrite
	}
	l.tail.Set(t + 1)
	pos := l.position()
	end := int(t)
	for end-(int(h)+pos) >= 0 {
		l.RuneAt(int(end) + 1).Set(*l.RuneAt(int(end)))
		end--
	}
	l.RuneAt(int(h) + pos).SetRune(key)
	if l.disp.Echo() {
		// Temporarily adjust head to rewrite only the changed portion of text.
		l.head.Set(h + uint32(pos))
		// Write out the text right-of the insertion.
		err = l.flush()
		// Reset head back to the actual beginning of the line.
		l.head.Set(h)
	}
	if e := l.moveCursorTo(int(pos) + 1); err == nil && e != nil {
		err = e
	}
	return
}

// moveCursorTo appends key sequences to the output buffer that move the cursor
// to the given logical position in the text, updating l's logical cursor
// position and the cursor's X, Y coordinates.
func (l *Line) moveCursorTo(pos int) (err error) {
	if pos < 0 {
		pos = 0
	}
	l.posi.Set(uint32(pos))
	if !l.disp.Echo() {
		return
	}
	w := l.disp.Width()
	x := pos + l.promptGlyphCount()
	y := x / w
	x %= w
	xc, yc := l.curs.Get()
	var du, dd, dl, dr int
	if y < yc {
		du = yc - y
	}
	if y > yc {
		dd = y - yc
	}
	if x < xc {
		dl = xc - x
	}
	if x > xc {
		dr = x - xc
	}
	_, _ = l.curs.Set(x, y)
	return l.curs.Move(du, dd, dl, dr)
}

// position returns the logical cursor position in the text of l.
//
// At position 0, the cursor is located on the first rune in l wherever the text
// of l happens to be.
func (l *Line) position() int {
	return int(l.posi.Get())
}

// Set overwrites the text in l and positions the cursor at the end of the line.
func (l *Line) Set(s []rune) (err error) {
	return l.SetPosition(s, -1)
}

// SetPosition overwrites the text in l and sets logical cursor position to pos.
// If pos is negative, the cursor is positioned after the last rune in s.
func (l *Line) SetPosition(s []rune, pos int) (err error) {
	if l == nil {
		return ErrReceiverLineSet
	}
	if len(s) > config.RunesPerLine {
		err = ErrOverflowLineSet
		s = s[:config.RunesPerLine]
	}
	prev := l.RuneCount()
	curr := len(s)
	l.Reset()
	for i := range s {
		l.Rune[i].SetRune(s[i])
	}
	tail := 0
	for tail = curr; tail < prev; tail++ {
		l.Rune[tail].Set(key.Space)
	}
	if l.disp.Echo() {
		if e := l.moveCursorTo(0); err == nil && e != nil {
			err = e
		}
		l.tail.Set(uint32(tail))
		if e := l.flush(); err == nil && e != nil {
			err = e
		}
	}
	l.tail.Set(uint32(curr))
	if pos < 0 {
		// Position cursor at end of line if pos is negative.
		pos = curr
	}
	if e := l.moveCursorTo(pos); err == nil && e != nil {
		err = e
	}
	return
}

// flush moves all buffered bytes in l to the output buffer and advances the
// cursor's current position to the end of the line.
func (l *Line) flush() (err error) {
	width := l.disp.Width()
	h, t := l.head.Get(), l.tail.Get()
	for t-h > 0 {
		free := width - l.curs.X()
		have := int(t - h)
		want := have
		if want > free {
			want = free
		}
		var seen, kept int
		// Copy the bytes in each rune of l to the output buffer, skipping any runes
		// with an invalid encoding.
		for ; kept < want && seen < have; seen++ {
			if _, errc := l.ctrl.Out.ReadFrom(l.RuneAt(int(h) + seen)); errc == nil {
				kept++
			}
		}
		// Update the cursor's coordinates based on the number of valid, visible
		// runes written to the output buffer.
		if l.curs.Update(l.glyphCount(0, seen)) {
			// If the cursor would write beyond the terminal width (line wrap), then
			// also append CR+LF to the output buffer.
			if _, err = l.ctrl.Out.WriteEOL(); err != nil {
				return
			}
		}
		h += uint32(seen)
	}
	_, err = l.ctrl.Flush()
	return
}

// Read copies up to len(p) bytes from l to p and returns the number of bytes
// successfully copied.
//
// If len(p) < l.Len(), and the UTF-8 encoding of the last rune being read
// contains more bytes than the remaining space of p, then that rune will remain
// unread in l, and a non-nil error will be returned.
// func (l *Line) Read(p []byte) (n int, err error) {
// 	if l == nil {
// 		return 0, ErrReceiverLineRead
// 	}
// 	if p == nil {
// 		return 0, ErrArgumentLineRead
// 	}
// 	nl, np := l.Len(), len(p)
// 	if nl <= np {
// 		np, err = nl, io.EOF
// 	}
// 	h, t := l.head.Get(), l.tail.Get()
// 	for n < np {
// 		if h == t {
// 			// If there were any encoding errors, they should have been accounted for
// 			// in l.Len, which serves as a maximum for np (above). Thus, it should not
// 			// be possible to reach the end of l (h == t) before encoding np bytes.
// 			//
// 			// This hasn't been verified in testing, so I'm leaving this check in for
// 			// now, since it adds safety and seems relatively inexpensive.
// 			err = io.EOF
// 			break
// 		}
// 		ih := h % config.RunesPerLine
// 		if ne, erre := l.Rune[ih].Encode(p[n:]); erre == nil {
// 			n += ne // No encode error, increment total bytes encoded (n).
// 		} else {
// 			// If utf8.ErrOverflow, then len(p) < l.Len(), but still an invalid size.
// 			if errt, overflow := erre.(utf8.ErrOverflow); overflow {
// 				err = errt
// 				break
// 			}
// 		}
// 		h++ // Move to next first-in (head) element regardless of encode result.
// 	}
// 	if err == io.EOF {
// 		l.Reset()
// 	} else {
// 		l.head.Set(h)
// 	}
// 	return
// }

// Write appends up to len(p) bytes from p to l and returns the number of bytes
// copied.
//
// Write does not extend the length of l to make room for all of p. It will only
// write to the free space in l and then return an error if all of p could not
// be copied.
//
// To overwrite any existing runes in l, call Reset before calling Write.
func (l *Line) Write(p []byte) (n int, err error) {
	if l == nil {
		return 0, ErrReceiverLineWrite
	}
	if p == nil {
		return 0, ErrArgumentLineWrite
	}
	np := len(p)
	if np == 0 {
		return
	}
	h, t := l.head.Get(), l.tail.Get()
	r := bytes.NewReader(p)
	for r.Len() > 0 {
		if t-h >= config.RunesPerLine {
			break
		}
		cu, nu, erru := r.ReadRune()
		if erru != nil {
			// Only keep the first error generated.
			if err == nil {
				err = erru
			}
			// Skip invalid rune, but continue copying remaining bytes.
			continue
		}
		l.RuneAt(int(t)).SetRune(cu)
		n += nu
		t++
	}
	l.tail.Set(t)
	if r.Len() > 0 {
		err = ErrOverflowLineWrite
	}
	return
}

// Errors returned by Line methods.
var (
	ErrReceiverLineSet   error
	ErrOverflowLineSet   error
	ErrReceiverLineRead  error
	ErrArgumentLineRead  error
	ErrReceiverLineWrite error
	ErrArgumentLineWrite error
	ErrOverflowLineWrite error
)

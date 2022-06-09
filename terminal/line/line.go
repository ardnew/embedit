package line

import (
	"bytes"

	"github.com/ardnew/embedit/config"
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

// glyphCount returns the number of visible glyphs in the slice k:k+n of l.
// If n is negative, returns the number of visible glyphs in l starting at k.
func (l *Line) glyphCount(k, n int) (count int) {
	h, t := l.head.Get(), l.tail.Get()
	if uint32(k) >= t-h {
		return 0 // Offset k is greater than rune count.
	}
	h += uint32(k)
	var g key.GlyphCount
	for t-h > 0 && n != 0 {
		count = g.Scan(l.Rune[h%config.RunesPerLine].Rune())
		h++
		n--
	}
	return count
}

// GlyphCount returns the number of visible glyphs in l.
func (l *Line) GlyphCount() (count int) {
	return l.glyphCount(0, -1)
}

// RuneCount returns the number of runes in l.
func (l *Line) RuneCount() (count int) {
	if l == nil {
		return 0
	}
	return int(l.tail.Get() - l.head.Get())
}

func (l *Line) PromptCount() (count int) {
	if l == nil {
		return 0
	}
	var g key.GlyphCount
	for _, r := range l.disp.Prompt() {
		count = g.Scan(r)
	}
	return count
}

// countToLeftWord returns then number of places from the cursor to the start of
// the previous word.
func (l *Line) countToLeftWord() int {
	origPos := l.pos()
	if origPos == 0 {
		return 0
	}
	head := int(l.head.Get())
	pos := origPos - 1
	for pos > 0 {
		if !l.Rune[(head+pos)%config.RunesPerLine].Equals(' ') {
			break
		}
		pos--
	}
	for pos > 0 {
		if l.Rune[(head+pos)%config.RunesPerLine].Equals(' ') {
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
	origPos := l.pos()
	head := int(l.head.Get())
	eol := l.RuneCount()
	pos := origPos
	for pos < eol {
		if l.Rune[(head+pos)%config.RunesPerLine].Equals(' ') {
			break
		}
		pos++
	}
	for pos < eol {
		if !l.Rune[(head+pos)%config.RunesPerLine].Equals(' ') {
			break
		}
		pos++
	}
	return pos - origPos
}

func (l *Line) ErasePrevious(n int) (err error) {
	if n == 0 {
		return
	}
	pos := l.pos()
	if pos < n {
		n = pos
	}
	pos -= n
	if err = l.setPos(pos); err != nil {
		return err
	}
	// Overwrite leading runes with trailing runes
	h, t := l.head.Get(), l.tail.Get()-uint32(n)
	s := scanner{line: l}
	if hs := h; s.slice(pos+n, -1) {
		for {
			r, ok := s.next()
			if !ok {
				break
			}
			l.Rune[(hs+uint32(pos))%config.RunesPerLine] = r
			hs++
		}
	}
	// Erase the trailing runes with spaces
	for i := 0; i < n; i++ {
		l.Rune[(t+uint32(i))%config.RunesPerLine] = key.Space
	}
	if l.disp.Echo() {
		// Temporarily adjust head to rewrite only the changed portion of text.
		l.head.Set(h + uint32(pos))
		// Write out the text right-of the deletion, including the 0x20 erasors
		if e := l.Queue(); e != nil {
			err = e
		}
		// Reset head back to the actual beginning of the line.
		l.head.Set(h)
	}
	// Finally truncate tail, set final cursor position, and flush output buffer.
	l.tail.Set(t)
	l.setPos(pos)
	return
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

// pos returns the logical cursor position in the text of l.
func (l *Line) pos() int {
	return int(l.posi.Get())
}

// setPos appends key sequences to the output buffer that move the cursor to the
// given logical position in the text, updating l and the cursor's coordinates.
func (l *Line) setPos(pos int) (err error) {
	if pos < 0 {
		pos = 0
	}
	l.posi.Set(uint32(pos))
	if !l.disp.Echo() {
		return
	}
	w := l.disp.Width()
	x := pos + l.PromptCount()
	y := x / w
	x %= w
	var (
		xc, yc         = l.curs.Get()
		du, dd, dl, dr int
	)
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
	return l.curs.Queue(du, dd, dl, dr)
}

// SetPos overwrites the text in l and sets its Cursor's logical position.
// If pos is negative, the Cursor is positioned at the end of the line.
func (l *Line) SetPos(s []rune, pos int) (err error) {
	if l == nil {
		return ErrReceiverLineSet
	}
	if len(s) > config.RunesPerLine {
		err = ErrOverflowLineSet
		s = s[:config.RunesPerLine]
	}
	prevCount := l.RuneCount()
	currCount := len(s)
	l.Reset()
	for i := range s {
		l.Rune[i] = utf8.Rune(s[i])
	}
	padLength := 0
	for padLength = currCount; padLength < prevCount; padLength++ {
		l.Rune[padLength] = utf8.Rune(key.Space)
	}
	if l.disp.Echo() {
		if e := l.setPos(0); err == nil && e != nil {
			err = e
		}
		l.tail.Set(uint32(padLength))
		if e := l.Queue(); err == nil && e != nil {
			err = e
		}
	}
	l.tail.Set(uint32(currCount))
	if pos < 0 {
		// Position cursor at end of line if pos is negative.
		pos = currCount
	}
	if e := l.setPos(pos); err == nil && e != nil {
		err = e
	}
	return
}

// Set overwrites the text in l and positions the cursor at the end of the line.
func (l *Line) Set(s []rune) (err error) {
	return l.SetPos(s, -1)
}

// Queue appends l to the output buffer and advances the cursor's current
// position to the end of the line.
func (l *Line) Queue() (err error) {
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
		for kept < want && seen < have {
			curr := &l.Rune[(int(h)+seen)%config.RunesPerLine]
			seen++
			if _, errc := l.ctrl.Out.ReadFrom(curr); errc == nil {
				kept++
			}
		}
		// Update the cursor's coordinates based on the number of valid, visible
		// runes written to the output buffer.
		if l.curs.Advance(l.glyphCount(0, seen)) {
			// If the cursor would write beyond the terminal width (line wrap), then
			// also append CR+LF to the output buffer.
			if _, err = l.ctrl.Out.Write(key.CRLF); err != nil {
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
// unread in l, and err will be set to utf8.ErrOverflow.
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
// write to the free space in l and then return ErrOverflow if all of p could
// not be copied.
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
		l.Rune[t%config.RunesPerLine] = utf8.Rune(cu)
		n += nu
		t++
	}
	l.tail.Set(t)
	if r.Len() > 0 {
		err = ErrOverflowLineWrite
	}
	return
}

type scanner struct {
	line       *Line
	head, tail uint32
}

func (s *scanner) reset() {
	if s == nil || s.line == nil {
		s.head, s.tail = 0, 0
	} else {
		s.head, s.tail = s.line.head.Get(), s.line.tail.Get()
	}
}

func (s *scanner) slice(lo, hi int) (ok bool) {
	s.reset()
	if s == nil || s.line == nil || s.tail-s.head <= 0 {
		return false // Invalid or empty receiver
	}
	if lo < 0 {
		// From 0 to hi-1
		lo = 0
	}
	if hi < 0 {
		// From lo to length-1
		hi = int(s.tail - s.head)
	}
	if lo >= hi || uint32(hi) > s.tail-s.head {
		// The above condition implies 0<=lo < hi<=N:
		//   If lo<hi and lo>=0, then hi>0 (i.e.: 0<=lo<hi => hi>0).
		//   If lo<hi and hi<=N, then lo<N (i.e.: lo<hi<=N => lo<N).
		return false
	}
	s.tail = s.head + uint32(hi)
	s.head = s.head + uint32(lo)
	return true
}

func (s *scanner) next() (r utf8.Rune, ok bool) {
	if s != nil && s.line != nil && s.tail-s.head > 0 {
		r = s.line.Rune[s.head%config.RunesPerLine]
		s.head++
		return r, true
	}
	return key.Null, false
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

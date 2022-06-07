package line

import (
	"bytes"
	"io"

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

// visibleCount returns the number of visible glyphs in the slice k:k+n of l.
// If n is negative, returns the number of visible glyphs in l starting at k.
func (l *Line) visibleCount(k, n int) (count int) {
	escape := false
	h, t := l.head.Get(), l.tail.Get()
	if uint32(k) >= t-h {
		return 0 // Offset k is greater than rune count.
	}
	h += uint32(k)
	for t-h > 0 && n != 0 {
		r := l.Rune[h%config.RunesPerLine]
		switch {
		case escape:
			escape = (r < 'a' || 'z' < r) && (r < 'A' || 'Z' < r)
		case r == key.Escape:
			escape = true
		default:
			count++
		}
		h++
		n--
	}
	return count
}

// VisibleCount returns the number of visible glyphs in l.
func (l *Line) VisibleCount() (count int) {
	return l.visibleCount(0, -1)
}

// RuneCount returns the number of runes in l.
func (l *Line) RuneCount() (count int) {
	if l == nil {
		return 0
	}
	return int(l.tail.Get() - l.head.Get())
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

// Set overwrites the bytes in l and sets its Cursor's logical position.
// If pos is negative, the Cursor is positioned at the end of the line.
func (l *Line) Set(s []rune, pos int) (err error) {
	if l == nil {
		return Error(ErrReceiver_Line_Set)
	}
	if len(s) > config.RunesPerLine {
		err = Error(ErrOverflow_Line_Set)
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
		if e := l.curs.MoveTo(0); err == nil && e != nil {
			err = e
		}
		l.tail.Set(uint32(padLength))
		if e := l.Queue(); err == nil && e != nil {
			err = e
		}
		l.tail.Set(uint32(currCount))
		if e := l.curs.MoveTo(pos); err == nil && e != nil {
			err = e
		}
	}
	l.setPos(pos)
	// if l.disp.Echo() {
	// 	l.ctrl.Out.Reset()
	// 	if e := l.curs.MoveTo(0); e != nil {
	// 		err = e
	// 	}
	// 	if e := l.WriteLine(s); e != nil && err == nil {
	// 		err = e
	// 	}
	// 	for i := len(s); i < l.RuneCount(); i++ {
	// 		if e := l.WriteLine(key.Blank); e != nil && err == nil {
	// 			err = e
	// 		}
	// 	}
	// 	if e := l.curs.MoveTo(pos); e != nil && err == nil {
	// 		err = e
	// 	}
	// }
	// if _, e := l.Write([]byte(string(s))); e != nil && err == nil {
	// 	err = e
	// }
	// l.setPos(pos)
	return
}

// pos returns the logical position of the Cursor within l.
func (l *Line) pos() int {
	return int(l.posi.Get())
}

// setPos sets the logical position of the Cursor within l.
func (l *Line) setPos(pos int) {
	if pos < 0 {
		pos = 0
	}
	l.posi.Set(uint32(pos))
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
			curr := l.Rune[(int(h)+seen)%config.RunesPerLine]
			seen++
			if _, errc := io.Copy(l.ctrl.Out, curr); errc == nil {
				// size += int(nc)
				kept++
			}
		}
		// Update the cursor's coordinates based on the number of valid, visible
		// runes written to the output buffer.
		if l.curs.Advance(l.visibleCount(int(h), seen)) {
			// If the cursor would write beyond the terminal width (line wrap), then
			// also append CR+LF to the output buffer.
			if _, err = l.ctrl.Out.Write(key.CRLF); err != nil {
				return
			}
		}
		h += uint32(seen)
	}
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
// 		return 0, Error(ErrReceiver_Line_Read)
// 	}
// 	if p == nil {
// 		return 0, Error(ErrArgument_Line_Read)
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
		return 0, Error(ErrReceiver_Line_Write)
	}
	if p == nil {
		return 0, Error(ErrArgument_Line_Write)
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
		err = Error(ErrOverflow_Line_Write)
	}
	return
}

// Types of errors returned by Line methods.
type Error string

const (
	ErrReceiver_Line_Set   = "line [receiver]: cannot Set with nil receiver"
	ErrOverflow_Line_Set   = "line [overflow]: cannot Set entire buffer (truncated)"
	ErrReceiver_Line_Read  = "line [receiver]: cannot Read from nil receiver"
	ErrArgument_Line_Read  = "line [argument]: cannot Read into nil buffer"
	ErrReceiver_Line_Write = "line [receiver]: cannot Write into nil receiver"
	ErrArgument_Line_Write = "line [argument]: cannot Write from nil buffer"
	ErrOverflow_Line_Write = "line [overflow]: cannot Write entire buffer (short write)"
)

func (e Error) Error() string { return string(e) }

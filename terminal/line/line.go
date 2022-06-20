package line

import (
	"bytes"

	"github.com/ardnew/embedit/config/limits"
	"github.com/ardnew/embedit/errors"
	"github.com/ardnew/embedit/seq/ansi"
	"github.com/ardnew/embedit/seq/utf8"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/wire"
	"github.com/ardnew/embedit/volatile"
)

// Line represents a single line of input.
type Line struct {
	curs  *cursor.Cursor
	ctrl  *wire.Control
	disp  *display.Display
	Rune  [limits.RunesPerLine]utf8.Rune
	posi  volatile.Register32
	head  volatile.Register32
	tail  volatile.Register32
	iter  utf8.Iterable
	flush bool
	paste bool
	valid bool
}

// Configure initializes the Line configuration.
func (l *Line) Configure(flush bool, curs *cursor.Cursor) *Line {
	if l == nil {
		return nil
	}
	if l.valid {
		// Configure must be called one time only.
		// Use object methods to modify configuration/state.
		return l
	}
	l.valid = false
	l.flush = flush
	l.curs = curs
	l.iter.Iterator = l
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
	if l == nil || !l.valid {
		return nil
	}
	l.paste = false
	l.posi.Set(0)
	l.head.Set(0)
	l.tail.Set(0)
	l.iter.Reset()
	return l
}

// LineFeed flushes all data to output device and resets the cursor, data, and
// I/O buffers to begin processing a new line.
func (l *Line) LineFeed() {
	if l != nil && l.ctrl != nil && l.curs != nil {
		l.Reset().curs.LineFeed()
	}
}

// EnableAutoFlush enables or disables auto-flush.
func (l *Line) EnableAutoFlush(enable bool) (wasEnabled bool) {
	if l == nil || !l.valid {
		return false
	}
	wasEnabled = l.flush
	l.flush = enable
	return
}

// Len returns the number of bytes in l.
func (l *Line) Len() (n int) {
	if l == nil || !l.valid {
		return 0
	}
	ih := l.head.Get() % limits.RunesPerLine
	it := l.tail.Get() % limits.RunesPerLine
	if ih > it {
		return utf8.RunesLen(l.Rune[ih:]) + utf8.RunesLen(l.Rune[:it])
	}
	return utf8.RunesLen(l.Rune[ih:it])
}

// IsPasted returns true if and only if the entire line consists only of pasted
// data.
func (l *Line) IsPasted() bool { return l.paste }

// SetIsPasted sets the flag that indicates whether the entire line consists
// only of pasted data.
func (l *Line) SetIsPasted(paste bool) { l.paste = paste }

// RuneHead returns the head index of l. Implements utf8.Iterator.
func (l *Line) RuneHead() uint32 { return l.head.Get() }

// RuneTail returns the tail index of l. Implements utf8.Iterator.
func (l *Line) RuneTail() uint32 { return l.tail.Get() }

// RuneAt returns the Rune at index i in l. Implements utf8.Iterator.
func (l *Line) RuneAt(i int) *utf8.Rune { return &l.Rune[i%limits.RunesPerLine] }

// RuneCount returns the total number of runes in l.
func (l *Line) RuneCount() (count int) {
	if l == nil || !l.valid {
		return 0
	}
	return int(l.tail.Get() - l.head.Get())
}

// RuneCountToStartOfWord returns the number of places from the cursor to the
// start of the current or previous word.
func (l *Line) RuneCountToStartOfWord() (n int) {
	from := l.Position()
	if from == 0 {
		return 0
	}
	head := int(l.head.Get())
	pos := from - 1
	found := false
	for pos > 0 {
		// Found a non-space rune, so a current/previous word must exist.
		found = !l.RuneAt(head + pos).Equals(' ')
		if found {
			break
		}
		pos--
	}
	// Return 0 if no previous word exists.
	//  (i.e., leading runes are all white space)
	if l.RuneAt(head).Equals(' ') {
		if !found {
			return 0
		}
	}
	for pos > 0 {
		if l.RuneAt(head + pos).Equals(' ') {
			pos++
			break
		}
		pos--
	}
	return from - pos
}

// RuneCountToStartOfNextWord returns then number of places from the cursor to
// the start of the next word.
func (l *Line) RuneCountToStartOfNextWord() (n int) {
	from := l.Position()
	head := int(l.head.Get())
	eol := l.RuneCount()
	pos := from
	for pos < eol {
		if l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos++
	}
	for pos < eol {
		// We only reach here if a space was found prior to EOL.
		// So if we find a non-space rune after that, a next word must exist.
		if !l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos++
	}
	return pos - from
}

// RuneCountToEndOfWord returns the number of places from the cursor to the end
// of the current word.
func (l *Line) RuneCountToEndOfWord() (n int) {
	from := l.Position()
	head := int(l.head.Get())
	eol := l.RuneCount()
	pos := from
	for pos < eol {
		if l.RuneAt(head + pos).Equals(' ') {
			break
		}
		pos++
	}
	return pos - from
}

// InsertRune inserts key at the current cursor position in l.
func (l *Line) InsertRune(key rune) (err error) {
	if l == nil || !l.valid {
		return &errors.ErrInvalidReceiver
	}
	h, t := l.head.Get(), l.tail.Get()
	if t-h >= limits.RunesPerLine {
		return &errors.ErrWriteOverflow
	}
	l.tail.Set(t + 1)
	pos := l.Position()
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
		err = l.Flush()
		// Reset head back to the actual beginning of the line.
		l.head.Set(h)
	}
	if e := l.MoveCursorTo(int(pos) + 1); err == nil && e != nil {
		err = e
	}
	if l.flush {
		l.ctrl.Flush()
	}
	return
}

// ErasePreviousRuneCount erases up to n previous runes from the current cursor
// position. Retained trailing runes are moved left in place of the runes
// erased.
//
// Appends sequences to the output buffer for repositioning the cursor and
// overwriting the portion of text that changed.
func (l *Line) ErasePreviousRuneCount(n int) (err error) {
	if l == nil || !l.valid {
		return &errors.ErrInvalidReceiver
	}
	if n <= 0 {
		return
	}
	pos := l.Position()
	if pos < n {
		n = pos
	}
	pos -= n
	if err = l.MoveCursorTo(pos); err != nil {
		return err
	}
	// Overwrite leading runes with trailing runes
	h, t := l.head.Get(), l.tail.Get()-uint32(n)
	if hs, s := h, l.iter.Slice(pos+n, -1); s != nil {
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
		l.RuneAt(int(t + uint32(i))).Set(' ')
	}
	if l.disp.Echo() {
		// Temporarily adjust head to rewrite only the changed portion of text.
		l.head.Set(h + uint32(pos))
		// Write out the text right-of the deletion, including the 0x20 erasors.
		err = l.Flush()
		// Reset head back to the actual beginning of the line.
		l.head.Set(h)
	}
	// Finally truncate tail, set final cursor position, and flush output buffer.
	l.tail.Set(t)
	if e := l.MoveCursorTo(pos); err == nil && e != nil {
		err = e
	}
	if l.flush {
		l.ctrl.Flush()
	}
	return
}

func (l *Line) ClearScreen() (err error) {
	_, err = l.ctrl.Out.Write(ansi.CLS)
	if _, e := l.ctrl.Out.Write(ansi.XY0); err == nil && e != nil {
		err = e
	}
	l.curs.Set(0, 0)
	if l.flush {
		l.ctrl.Flush()
	}
	return
}

// glyphCount returns the number of runes in l from k to k+n-1 that are not part
// of an escape sequence. If n is negative, returns the number of unescaped
// runes in l starting at k.
//
// See type Iterable (github.com/ardnew/embedit/seq/utf8) for more details.
//
// The value k = 0 always refers to head, i.e., the first rune of l; it does not
// refer to index 0 of the backing array. It is the responsibility of this
// method to compute the appropriate offsets into the circular FIFO and account
// for possible wraparound based on l's current head and tail. The caller must
// not account for these offsets. Otherwise, incorrect indexing caused by double
// correction will almost always occur.
func (l *Line) glyphCount(k, n int) (count int) {
	if n >= 0 {
		n += k
	}
	return l.iter.Slice(k, n).GlyphCount()
}

// GlyphCount returns the number of runes in l that are not part of an escape
// sequence.
func (l *Line) GlyphCount() (count int) {
	return l.glyphCount(0, -1)
}

// Position returns the logical cursor Position in the text of l.
//
// At Position 0, the cursor is located on the first rune in l wherever the text
// of l happens to be.
func (l *Line) Position() int {
	return int(l.posi.Get())
}

func (l *Line) setPosition(position int) int {
	if position < 0 {
		position = 0
	}
	if end := l.RuneCount(); position > end {
		position = end
	}
	l.posi.Set(uint32(position))
	return position
}

// moveCursorTo appends key sequences to the output buffer that move the cursor
// to the given logical position in the text, adjusted by offset, and updates
// l's logical cursor position and the cursor's X, Y coordinates. The offset is
// typically the number of places in — or width of — the input prompt.
func (l *Line) moveCursorTo(position, offset int) (err error) {
	x := offset + l.setPosition(position)
	if !l.disp.Echo() {
		return
	}
	w := l.disp.Width()
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

// MoveCursorTo appends key sequences to the output buffer that move the cursor
// to the given logical position in the text, updating l's logical cursor
// position and the cursor's X, Y coordinates.
func (l *Line) MoveCursorTo(position int) (err error) {
	return l.moveCursorTo(position, l.disp.GlyphCountInPrompt())
}

// MoveCursor appends sequences to the output buffer that move the cursor by the
// given number of places from the current cursor position, updating l's logical
// cursor position and the cursor's X, Y coordinates.
func (l *Line) MoveCursor(places int) (err error) {
	return l.MoveCursorTo(l.Position() + places)
}

// Set overwrites the text in l and positions the cursor at the end of the line.
func (l *Line) Set(s []rune) (err error) {
	return l.SetAndMoveCursorTo(s, -1)
}

// SetAndMoveCursorTo overwrites the text in l and then sets the logical cursor
// position to pos.
// If pos is negative, the cursor is positioned after the last rune in s.
func (l *Line) SetAndMoveCursorTo(s []rune, position int) (err error) {
	if l == nil || !l.valid {
		return &errors.ErrInvalidReceiver
	}
	if len(s) > limits.RunesPerLine {
		err = &errors.ErrWriteOverflow
		s = s[:limits.RunesPerLine]
	}
	prev := l.RuneCount()
	curr := len(s)
	l.Reset()
	for i := range s {
		l.Rune[i].SetRune(s[i])
	}
	tail := 0
	for tail = curr; tail < prev; tail++ {
		l.Rune[tail].Set(' ')
	}
	if l.disp.Echo() {
		if e := l.MoveCursorTo(0); err == nil && e != nil {
			err = e
		}
		l.tail.Set(uint32(tail))
		if e := l.Flush(); err == nil && e != nil {
			err = e
		}
	}
	l.tail.Set(uint32(curr))
	if position < 0 {
		// Position cursor at end of line if pos is negative.
		position = curr
	}
	if e := l.MoveCursorTo(position); err == nil && e != nil {
		err = e
	}
	if l.flush {
		l.ctrl.Flush()
	}
	return
}

// ShowPrompt appends the user input prompt to the output buffer, positions the
// cursor at the end of the prompt, and flushes the buffer to the output device.
func (l *Line) ShowPrompt() (err error) {
	if l == nil || !l.valid {
		return &errors.ErrInvalidReceiver
	}
	// Iterate over prompt elements as Rune elements (instead of native rune),
	// because it implements an unbuffered io.Reader for copying bytes in each
	// UTF-8 code point.
	s := l.disp.PromptIterable().Reset()
	if s == nil {
		return &errors.ErrInvalidArgument
	}
	for {
		if _, rerr := l.ctrl.Out.ReadFrom(s.Next()); rerr != nil {
			break
		}
	}
	if l.curs.Update(s.Reset().GlyphCount()) {
		// If the cursor would write beyond the terminal width (line wrap), then
		// also append CR+LF to the output buffer.
		_, _ = l.ctrl.Out.WriteEOL()
	}
	return l.Flush()
}

// Flush copies all runes in l to the output buffer and advances the cursor's
// current position to the end of the line.
func (l *Line) Flush() (err error) {
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
	if l.flush {
		l.ctrl.Flush()
	}
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
// 		ih := h % limits.RunesPerLine
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
	if l == nil || !l.valid {
		return 0, &errors.ErrInvalidReceiver
	}
	if p == nil {
		return 0, &errors.ErrInvalidArgument
	}
	np := len(p)
	if np == 0 {
		return
	}
	h, t := l.head.Get(), l.tail.Get()
	r := bytes.NewReader(p)
	for r.Len() > 0 {
		if t-h >= limits.RunesPerLine {
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
		err = &errors.ErrWriteOverflow
	}
	return
}

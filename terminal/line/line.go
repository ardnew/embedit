package line

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/terminal/display"
	"github.com/ardnew/embedit/terminal/key"
	"github.com/ardnew/embedit/terminal/line/cursor"
	"github.com/ardnew/embedit/terminal/wire"
	"github.com/ardnew/embedit/volatile"
)

// Types of errors returned by Line methods.
type (
	ErrReceiver string
	ErrArgument string
	ErrOverflow string
)

// Line represents a single line of input.
type Line struct {
	curs  cursor.Cursor
	Rune  [config.RunesPerLine]rune
	posi  volatile.Register32
	head  volatile.Register32
	tail  volatile.Register32
	valid bool
}

// Invalid represents an invalid Line.
var Invalid = Line{valid: false}

// Configure initializes the Line configuration.
func (l *Line) Configure(disp display.Display, wire wire.ReadWriter) *Line {
	if l == nil {
		return nil
	}
	l.valid = false
	l.curs.Configure(disp, wire)
	return l.init()
}

// init initializes the state of a configured Line.
func (l *Line) init() *Line {
	l.valid = true
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
		return runesLen(l.Rune[ih:]) + runesLen(l.Rune[:it])
	}
	return runesLen(l.Rune[ih:it])
}

// Reset sets the Line length to 0 and resets the cursor position.
func (l *Line) Reset() *Line {
	if l == nil {
		return nil
	}
	l.curs.Reset()
	l.posi.Set(0)
	l.head.Set(0)
	l.tail.Set(0)
	return l
}

// func (l *Line) String() string {
// 	if l != nil && l.valid {
// 		ih := l.head.Get() % config.RunesPerLine
// 		it := l.tail.Get() % config.RunesPerLine
// 		if ih > it {
// 			return string(l.Rune[ih:]) + string(l.Rune[:it])
// 		}
// 		return string(l.Rune[ih:it])
// 	}
// 	return ""
// }

// Set overwrites the bytes in l and sets its Cursor's logical position.
func (l *Line) Set(s []rune, pos int) (err error) {
	if l == nil {
		return ErrReceiver("cannot Set with nil receiver")
	}
	if l.curs.Echo() {
		if e := l.curs.MoveTo(0); e != nil {
			err = e
		}
		if e := l.curs.WriteLine(s); e != nil && err == nil {
			err = e
		}
		for i := len(s); i < l.Len(); i++ {
			if e := l.curs.WriteLine([]rune(key.Blank)); e != nil && err == nil {
				err = e
			}
		}
		if e := l.curs.MoveTo(pos); e != nil && err == nil {
			err = e
		}
	}
	if _, e := l.Write([]byte(string(s))); e != nil && err == nil {
		err = e
	}
	l.SetPos(pos)
	return
}

// Cursor returns a reference to the Cursor.
func (l *Line) Cursor() *cursor.Cursor {
	if l == nil {
		return nil
	}
	return &l.curs
}

// Pos returns the logical position of the Cursor within l.
func (l *Line) Pos() int {
	if l == nil {
		return 0
	}
	return int(l.posi.Get())
}

// SetPos sets the logical position of the Cursor within l.
func (l *Line) SetPos(pos int) {
	if l == nil {
		return
	}
	if pos < 0 {
		pos = 0
	}
	l.posi.Set(uint32(pos))
}

// Read copies up to len(p) bytes from l to p and returns the number of bytes
// copied.
func (l *Line) Read(p []byte) (n int, err error) {
	if l == nil {
		return 0, ErrReceiver("cannot Read from nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Read into nil buffer")
	}
	var nl int
	nl, n = l.Len(), len(p)
	if nl <= n {
		n, err = nl, io.EOF
	}
	h, i := l.head.Get(), 0
	for i < n {
		ih := h % config.RunesPerLine
		i += utf8.EncodeRune(p[i:], l.Rune[ih])
		h++
	}
	if err == io.EOF {
		l.Reset()
	} else {
		l.head.Set(h)
	}
	return
}

// Write copies up to len(p) bytes from p to the start of l and returns the
// number of bytes copied.
//
// Write overwrites any bytes present, but stops writing once l is full.
func (l *Line) Write(p []byte) (n int, err error) {
	if l == nil {
		return 0, ErrReceiver("cannot Write into nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Write from nil buffer")
	}
	np := len(p)
	if np == 0 {
		return
	}
	_ = l.Reset()
	var t uint32
	r := bytes.NewReader(p)
	for r.Len() > 0 {
		if t > config.RunesPerLine {
			break
		}
		cu, nu, erru := r.ReadRune()
		if erru != nil {
			// Skip invalid rune, but continue copying remaining bytes.
			err = erru
			continue
		}
		l.Rune[t] = cu
		n += nu
		t++
	}
	l.tail.Set(t)
	if r.Len() > 0 {
		err = ErrOverflow("cannot Write entire buffer (short write)")
	}
	return
}

// Append copies up to len(p) bytes from p to the end of l and returns the
// number of bytes copied.
//
// Unlike the append builtin, it does not extend the length of l to make room
// for all of p. It will only write to the free space in l and then return
// ErrOverflow if all of p could not be copied.
func (l *Line) Append(p []byte) (n int, err error) {
	if l == nil {
		return 0, ErrReceiver("cannot Append into nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Append from nil buffer")
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
			// Skip invalid rune, but continue copying remaining bytes.
			err = erru
			continue
		}
		l.Rune[t%config.RunesPerLine] = cu
		n += nu
		t++
	}
	l.tail.Set(t)
	if r.Len() > 0 {
		err = ErrOverflow("cannot Append entire buffer (short write)")
	}
	return
}

func (e ErrReceiver) Error() string { return "line [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "line [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "line [overflow]: " + string(e) }

// runesLen returns the number of bytes in p.
func runesLen(p []rune) (n int) {
	for _, r := range p {
		n += utf8.RuneLen(r)
	}
	return
}

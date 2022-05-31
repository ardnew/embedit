package line

import (
	"io"

	"github.com/ardnew/embedit/sys"
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
	Rune  [sys.RunesPerLine]rune
	curs  cursor.Cursor
	posi  volatile.Register32 // Logical position of Cursor.
	head  volatile.Register32
	tail  volatile.Register32
	valid bool // Has init been called
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
func (l *Line) Len() int {
	if l == nil {
		return 0
	}
	return len(l.String())
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

func (l *Line) String() string {
	if l != nil && l.valid {
		ih := l.head.Get() % sys.RunesPerLine
		it := l.tail.Get() % sys.RunesPerLine
		if ih > it {
			return string(l.Rune[ih:]) + string(l.Rune[:it])
		}
		return string(l.Rune[ih:it])
	}
	return ""
}

// Set overwrites the bytes in l and sets its Cursor's logical position.
func (l *Line) Set(s []rune, pos int) (err error) {
	if l == nil {
		return ErrReceiver("cannot Set with nil receiver")
	}
	e := []error{}
	// Always attempt to perform all operations, and save any errors returned
	// along the way. Once we've finished, we return the first non-nil error
	// that was added to the slice.
	if l.curs.Echo() {
		e = append(e, l.curs.MoveTo(0))
		e = append(e, l.curs.WriteLine(s))
		for i := len(s); i < l.Len(); i++ {
			e = append(e, l.curs.WriteLine([]rune(key.Blank)))
		}
		e = append(e, l.curs.MoveTo(pos))
	}
	_, ew := l.Write([]byte(string(s)))
	l.SetPos(pos)
	// Check if any of the appended errors are non-nil. Return the first, if any.
	for _, err = range append(e, ew) {
		if err != nil {
			return
		}
	}
	return
}

// Pos returns the logical position of the Cursor within l.
func (l *Line) Cursor() *cursor.Cursor {
	if l == nil {
		return nil
	}
	return &l.curs
}

// Pos returns the logical position of the Cursor within l.
func (l *Line) Pos() int {
	return int(l.posi.Get())
}

// Pos returns the logical position of the Cursor within l.
func (l *Line) SetPos(pos int) {
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
		return 0, ErrArgument("cannot Read to nil buffer")
	}
	var nl int
	nl, n = l.Len(), len(p)
	if nl <= n {
		n, err = nl, io.EOF
	}
	h, i := l.head.Get(), 0
	for i < n {
		ih := h % sys.RunesPerLine
		i += copy(p[i:], []byte(string(l.Rune[ih:ih+1])))
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
		return 0, ErrReceiver("cannot Write to nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Write from nil buffer")
	}
	n = copy([]byte(string(l.Reset().Rune[0:])), p)
	l.tail.Set(uint32(n))
	if np := len(p); n < np {
		err = ErrOverflow("cannot Write entire buffer (truncated)")
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
		return 0, ErrReceiver("cannot Append to nil receiver")
	}
	if p == nil {
		return 0, ErrArgument("cannot Append from nil buffer")
	}
	np := len(p)
	if np == 0 {
		return
	}
	if l.Len() == 0 {
		n = copy([]byte(string(l.Reset().Rune[0:])), p)
		l.tail.Set(uint32(n))
	} else {
		n = np
		h, t := l.head.Get(), l.tail.Get()
		for i, c := range string(p) {
			if t-h >= sys.RunesPerLine {
				n = i
				break
			}
			l.Rune[t%sys.RunesPerLine] = c
			t++
		}
		l.tail.Set(t)
	}
	if n < np {
		err = ErrOverflow("cannot Append entire buffer (truncated)")
	}
	return
}

func (e ErrReceiver) Error() string { return "line [receiver]: " + string(e) }
func (e ErrArgument) Error() string { return "line [argument]: " + string(e) }
func (e ErrOverflow) Error() string { return "line [overflow]: " + string(e) }

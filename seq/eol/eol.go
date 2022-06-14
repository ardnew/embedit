package eol

import (
	"io"

	"github.com/ardnew/embedit/seq/ascii"
)

// Mode defines end-of-line sequence conventions.
type Mode byte

// Constants of enumerated type Mode.
const (
	LF Mode = iota
	CRLF
	CR
)

// Platform aliases of Mode constants.
const Unix, DOS, Mac = LF, CRLF, CR

// ASCII byte sequences of enumerated type Mode.
var seq = [...][]byte{
	{ascii.LF},           // LF
	{ascii.CR, ascii.LF}, // CRLF
	{ascii.CR},           // CR
}

// WriteTo implements io.WriterTo.
func (m Mode) WriteTo(w io.Writer) (n int64, err error) {
	i, err := w.Write(seq[m])
	return int64(i), err
}

package eol

import "io"

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
	{0x0A},       // LF
	{0x0D, 0x0A}, // CRLF
	{0x0D},       // CR
}

// WriteTo implements io.WriterTo.
func (m Mode) WriteTo(w io.Writer) (n int64, err error) {
	i, err := w.Write(seq[m])
	return int64(i), err
}

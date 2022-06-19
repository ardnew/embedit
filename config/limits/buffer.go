package limits

import "unicode/utf8"

// maxBytesPerRune defines the maximum number of bytes in a UTF-8 encoded rune.
const MaxBytesPerRune = utf8.UTFMax

// BytesPerBuffer defines the maximum number of bytes in a buffer used for
// reading/writing terminal control/data byte sequences.
const BytesPerBuffer = MaxBytesPerRune * RunesPerLine // Max bytes per 1 line

// MaxBytesPerKey defines the maximum number of bytes in a key code sequence.
// This is the size of the buffer used for parsing each key from an I/O buffer.
//
// This buffer must be large enough to hold the byte sequence of a single key
// code sequence; it need not hold an entire escape sequence consisting of
// multiple key codes.
//
// Note that this probably isn't large enough for every possible valid key code
// byte sequence, but we need to place a reasonable upper bound on this space.
//
// Using ansi.go as reference (github.com/ardnew/embedit/seq/key), the widest
// keycode appears to be 16-bit
const MaxBytesPerKey = 2 * MaxBytesPerRune // TBD: Is there a _correct_ value?

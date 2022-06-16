//go:build !debug && !test
// +build !debug,!test

package config

const (
	// RunesPerLine defines the maximum number of runes in a line of input.
	RunesPerLine = 256
	// LinesPerHistory defines the maximum number of lines stored in history.
	// Old lines are discarded as more than LinesPerHistory are added.
	LinesPerHistory = 32
	// BytesPerBuffer defines the maximum number of bytes in a buffer used for
	// reading/writing terminal control/data byte sequences.
	BytesPerBuffer = 4 * RunesPerLine // 4-byte maximum UTF-8 rune size
	// MaxBytesPerKey defines the maximum number of bytes in a keycode sequence.
	// This is the size of the buffer used for parsing a key from an I/O buffer.
	// Note that this probably isn't large enough for every possible valid keycode
	// byte sequence, but we need to place a reasonable upperbounds on this space.
	// Currently only two of these buffers exist (each In and Out seq.Buffer in
	// wire.Control) for each Embedit object.
	MaxBytesPerKey = 8
)

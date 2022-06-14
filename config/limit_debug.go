//go:build debug || test
// +build debug test

package config

const (
	// RunesPerLine defines the maximum number of runes in a line of input.
	RunesPerLine = 8
	// LinesPerHistory defines the maximum number of lines stored in history.
	// Old lines are discarded as more than LinesPerHistory are added.
	LinesPerHistory = 5
	// BytesPerBuffer defines the maximum number of bytes in a buffer used for
	// reading/writing terminal control/data byte sequences.
	BytesPerBuffer = 4 * RunesPerLine // 4-byte maximum UTF-8 rune size
	// MaxBytesPerKey defines the maximum number of bytes in a keycode sequence.
	// This is the size of the buffer used for parsing a key from an I/O buffer.
	MaxBytesPerKey = 8
)

package limit

const (
	// RunesPerLine defines the maximum number of runes in a line of input.
	RunesPerLine = 256
	// LinesPerHistory defines the maximum number of lines stored in history.
	// Old lines are discarded as more than LinesPerHistory are added.
	LinesPerHistory = 32
	// BytesPerSequence defines the maximum number of bytes in a buffer used for
	// reading/writing terminal control/data byte sequences.
	BytesPerSequence = 4 * RunesPerLine // 4-byte maximum UTF-8 rune size
)

package limit

const (
	// RunesPerLine defines the maximum number of runes in a line of input.
	// New runes are discarded in favor of old runes.
	RunesPerLine = 96
	// LinesPerHistory defines the maximum number of lines stored in history.
	// Old lines are discarded in favor of new lines.
	LinesPerHistory = 64
)

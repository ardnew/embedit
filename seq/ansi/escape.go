package ansi

// Common escape sequences.
var (
	CSI = []byte{Escape, '['}                     // Ctrl seq intro
	SOP = []byte{Escape, '[', '2', '0', '0', '~'} // Start of paste
	EOP = []byte{Escape, '[', '2', '0', '1', '~'} // End of paste
	CLS = []byte{Escape, '[', '2', 'J'}           // Clear screen
	XY0 = []byte{Escape, '[', 'H'}                // Set cursor X=0 Y=0
	KIL = []byte{Escape, '[', 'K'}                // Clear line right
	DEL = []byte{' ', Escape, '[', 'D'}           // Delete next rune
)

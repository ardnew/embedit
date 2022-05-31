package display

// Display defines the interface of a terminal display's viewport.
type Display interface {
	Width() int
	Height() int
	Size() (width, height int)

	Echo() bool // True if input keystrokes are echoed to the display.

	Prompt() []rune
}

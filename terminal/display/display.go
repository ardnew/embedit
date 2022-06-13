package display

import (
	"github.com/ardnew/embedit/config"
	"github.com/ardnew/embedit/text/utf8"
	"github.com/ardnew/embedit/volatile"
)

// Display defines a terminal display's viewport.
type Display struct {
	prompt []rune
	width  volatile.Register32
	height volatile.Register32
	echo   volatile.Register8
	valid  bool
}

// Configure initializes the Display configuration.
func (d *Display) Configure(width, height int, prompt []rune, echo bool) *Display {
	if d == nil {
		return nil
	}
	d.valid = false
	d.SetSize(width, height)
	d.SetPrompt(prompt)
	d.SetEcho(echo)
	return d.init()
}

// init initializes the state of a configured Display.
func (d *Display) init() *Display {
	d.valid = true
	return d
}

// Width returns the Display width.
func (d *Display) Width() int {
	if d == nil {
		return 0
	}
	return int(d.width.Get())
}

// Height returns the Display height.
func (d *Display) Height() int {
	if d == nil {
		return 0
	}
	return int(d.height.Get())
}

// Size returns the Display width and height.
func (d *Display) Size() (width, height int) {
	if d == nil {
		return 0, 0
	}
	return int(d.width.Get()), int(d.height.Get())
}

// SetSize sets the Display width and height.
func (d *Display) SetSize(width, height int) {
	if d != nil {
		if width <= 0 {
			width = config.DefaultWidth
		}
		if height <= 0 {
			height = config.DefaultHeight
		}
		d.width.Set(uint32(width))
		d.height.Set(uint32(height))
	}
}

// Echo returns true if and only if input keystrokes are echoed to output.
func (d *Display) Echo() bool {
	return d != nil && d.echo.Get() != 0
}

// SetEcho sets echo true if and only if input keystrokes are echoed to output.
func (d *Display) SetEcho(echo bool) {
	if d != nil {
		if echo {
			d.echo.Set(1)
		} else {
			d.echo.Set(0)
		}
	}
}

// Prompt returns the user input prompt.
func (d *Display) Prompt() []rune {
	if d == nil || d.prompt == nil {
		return config.DefaultPrompt
	}
	return d.prompt
}

// PromptIterator returns the user input prompt as utf8.RuneIterator.
func (d *Display) PromptIterator() utf8.Iterator {
	if d == nil || d.prompt == nil {
		return (*utf8.IterableRune)(&config.DefaultPrompt)
	}
	return (*utf8.IterableRune)(&d.prompt)
}

// SetPrompt sets the user input prompt.
func (d *Display) SetPrompt(prompt []rune) {
	if d != nil {
		if prompt == nil {
			for i := range d.prompt {
				d.prompt[i] = 0
			}
		} else {
			d.prompt = prompt
		}
	}
}

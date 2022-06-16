package embedit

import (
	"io"

	"github.com/ardnew/embedit/terminal"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/line"
)

// Embedit defines the state and configuration of a line-buffered, commandline
// user interface with some capabilities of a modern terminal.
//
// It requires no dynamic memory allocation. Size limitations are defined by
// compile-time constants in package config.
//
// Refer to the examples to see how to allocate and configure the object for
// common use cases.
type Embedit struct {
	term  terminal.Terminal
	valid bool
}

// Config defines the configuration parameters of an Embedit.
type Config struct {
	RW     io.ReadWriter
	Prompt []rune
	Width  int
	Height int
}

// New allocates a new Embedit and returns a pointer to that object.
//
// The object remains uninitialized and unusable until Configure has been called
func New() Embedit { return Embedit{} }

// Configure initializes the Embedit configuration.
func (e *Embedit) Configure(config Config) *Embedit {
	e.valid = false
	_ = e.term.Configure(config.RW, config.Prompt, config.Width, config.Height)
	return e.init()
}

// init initializes the state of a configured Embedit.
func (e *Embedit) init() *Embedit {
	e.valid = true
	return e
}

// Terminal returns the terminal.
func (e *Embedit) Terminal() *terminal.Terminal {
	if e == nil || !e.valid {
		return nil
	}
	return &e.term
}

func (e *Embedit) Cursor() *cursor.Cursor {
	if e == nil || !e.valid {
		return nil
	}
	return e.term.Cursor()
}

// Line returns the Terminal's active user input line.
func (e *Embedit) Line() *line.Line {
	if e == nil || !e.valid {
		return nil
	}
	return e.term.Line()
}

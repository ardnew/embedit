package embedit

import (
	"io"

	"github.com/ardnew/embedit/terminal"
	"github.com/ardnew/embedit/terminal/cursor"
	"github.com/ardnew/embedit/terminal/line"
)

// Types of errors returned by Embedit methods.
type (
	ErrReceiver string
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

// Configure initializes the Embedit configuration.
func (e *Embedit) Configure(config Config) *Embedit {
	e.valid = false
	_ = e.term.Configure(config.RW, config.Prompt, config.Width, config.Height)
	return e
}

// init initializes the state of a configured Embedit.
func (e *Embedit) init() *Embedit {
	e.valid = true
	return e
}

func (e *Embedit) Cursor() *cursor.Cursor {
	if e == nil {
		return nil
	}
	return e.term.Cursor()
}

// Line returns the Terminal's active user input line.
func (e *Embedit) Line() *line.Line {
	if e == nil {
		return nil
	}
	return e.term.Line()
}

func (e ErrReceiver) Error() string { return "embedit [receiver]: " + string(e) }

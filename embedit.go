package embedit

import (
	"io"

	"github.com/ardnew/embedit/history"
	"github.com/ardnew/embedit/line"
	"github.com/ardnew/embedit/terminal"
)

// Embedit defines the state and configuration of a line-buffered, commandline
// user interface with some capabilities of a modern terminal.
//
// It requires no dynamic memory allocation. Size limitations are defined by
// compile-time constants in package limit.
//
// Refer to the examples to see how to allocate and configure the object for
// common use cases.
type Embedit struct {
	term  terminal.Terminal
	hist  history.History
	line  line.Line
	valid bool // Has init been called
}

// Config defines the configuration parameters of an Embedit.
type Config struct {
	RW     io.ReadWriter
	Width  int
	Height int
}

// Configure initializes the Embedit configuration.
func (e *Embedit) Configure(c Config) *Embedit {
	e.valid = false
	_ = e.term.Configure(c.RW, c.Width, c.Height)
	_ = e.hist.Configure()
	_ = e.line.Configure()
	return e
}

// init initializes the state of a configured Embedit.
func (e *Embedit) init() *Embedit {
	e.valid = true
	return e
}

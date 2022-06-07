// Package wire defines an API for transmitting serial data.
package wire

import "io"

type Reader interface {
	io.Reader
	io.ByteReader
	io.WriterTo
	Reset()
}

type Writer interface {
	io.Writer
	io.ByteWriter
	io.ReaderFrom
	Reset()
}

type Sweller interface {
	Swell() (int, error)
}

type Flusher interface {
	Flush() (int, error)
}

type Controller interface {
	Sweller
	Flusher
}

type Control struct {
	Controller
	In  Reader
	Out Writer
}

// MakeControl returns a Control object initialized with the given Controller
// and I/O buffers.
func (c *Control) Configure(ctrl Controller, in Reader, out Writer) *Control {
	c.Controller = ctrl
	c.In = in
	c.Out = out
	return c
}

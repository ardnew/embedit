// Package wire defines an API for transmitting serial data.
package wire

import "io"

type Reader interface {
	io.Reader
	ReadWire() (int, error)
}

type Writer interface {
	io.Writer
	WriteWire() (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}

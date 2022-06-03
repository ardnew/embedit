// Package wire defines an API for transmitting serial data.
package wire

import "io"

type Reader interface {
	io.Reader
	io.ReaderFrom
	io.ByteReader
	ReadWire() (int, error)
}

type Writer interface {
	io.Writer
	io.WriterTo
	io.ByteWriter
	WriteWire() (int, error)
}

type ReadWriter interface {
	Reader
	Writer
}

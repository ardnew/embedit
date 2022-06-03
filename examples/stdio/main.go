package main

import (
	"errors"
	"io"
	"os"

	"github.com/ardnew/embedit"
	"github.com/ardnew/embedit/sys"
)

// Static storage for our main object.
var em embedit.Embedit

// A simple io.ReadWriter used in the embedit.Config object.
var rw = &struct {
	io.Reader
	io.Writer
}{os.Stdin, os.Stdout}

type mainFunc func() error

func main() {
	if err := run(app); err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Stderr.Write([]byte{'\n'})
	}
}

func app() error {
	em.Configure(embedit.Config{RW: rw, Width: 80, Height: 24})

	f := sys.MakeFdio(int(os.Stdin.Fd()))
	if !f.Valid() || !f.Raw() {
		return errors.New("cannot attach terminal to stdin")
	}
	defer f.Restore()

	for i := 0; i < 10000; i++ {
		em.Line().Set([]rune("hello there"), 5)
		// time.Sleep(1 * time.Second)
		em.Line().Cursor().Move(0, 2, 0, 0)
		// time.Sleep(1 * time.Second)
		em.Line().Cursor().Move(2, 0, 0, 10)
	}

	return nil
}

package main

import (
	"io"
	"os"
	"time"

	"github.com/ardnew/embedit"
	"github.com/ardnew/embedit/sysio"
)

// Static storage for our main object.
var em embedit.Embedit

// A simple io.ReadWriter used in the embedit.Config object.
var rw = &struct {
	io.Reader
	io.Writer
}{os.Stdin, os.Stdout}

func main() {
	em.Configure(embedit.Config{RW: rw, Width: 80, Height: 24})

	f := sysio.MakeFdio(int(os.Stdin.Fd()))
	if !f.Valid() || !f.Raw() {
		return
	}
	defer f.Restore()

	em.Line().Set([]rune("hello there"), 5)
	time.Sleep(1 * time.Second)
	em.Line().Cursor().Move(0, 2, 0, 0)
	time.Sleep(1 * time.Second)
	em.Line().Cursor().Move(2, 0, 0, 10)

	profile()
}

type stdio struct {
	io.Reader
	io.Writer
}

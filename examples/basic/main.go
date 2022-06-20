package main

import (
	"io"
	"os"

	"github.com/ardnew/embedit"
	"github.com/ardnew/embedit/sys"
)

const pkgName = "stdio"

// Static storage for our main object.
var em embedit.Embedit

// A simple io.ReadWriter used in the embedit.Config object.
var rw = &struct {
	io.Reader
	io.Writer
}{os.Stdin, os.Stdout}

func main() {
	f := sys.MakeFdio(int(os.Stdin.Fd()))
	if !f.Valid() || !f.Raw() {
		panic(os.ErrInvalid)
	}
	defer f.Restore()

	em.Configure(embedit.Config{RW: rw, Width: 80, Height: 24})
	for {
		if em.Terminal().ReadLine() != nil {
			return
		}
	}
}

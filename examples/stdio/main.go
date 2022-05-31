package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardnew/embedit"
)

var (
	em embedit.Embedit
	cf = embedit.Config{
		RW:     &stdio{os.Stdin}, //, os.Stdout},
		Width:  80,
		Height: 24,
	}
)

func main() {
	em.Configure(cf)

	em.Line().Set([]rune("hello there"), 8)
	em.Line().Cursor().Move(1, 2, 3, 4)
	em.Line().Cursor().Move(6, 7, 8, 9)
	for {
		time.Sleep(time.Minute)
	}
}

type stdio struct {
	io.Reader
	// io.Writer
}

func (s *stdio) Write(p []byte) (n int, err error) {
	fmt.Printf("%#v\n", p)
	fmt.Printf("%s", p)
	return len(p), nil
}

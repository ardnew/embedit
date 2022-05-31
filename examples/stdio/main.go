package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardnew/embedit"
	"github.com/ardnew/embedit/examples/stdio/sysio"
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

	state := sysio.NewState(int(os.Stdin.Fd()))
	if state != nil {
		if err := state.Raw(); err != nil {
			fmt.Println("error:", err)
			return
		}
		defer state.Restore()

		for {
			em.Line().Set([]rune("hello there"), 5)
			time.Sleep(2 * time.Second)
			em.Line().Cursor().Move(0, 2, 0, 0)
			time.Sleep(2 * time.Second)
			em.Line().Cursor().Move(2, 0, 0, 10)
			time.Sleep(10 * time.Second)
		}
	}
}

type stdio struct {
	io.Reader
	// io.Writer
}

func (s *stdio) Write(p []byte) (n int, err error) {
	// fmt.Printf("%#v\n", p)
	fmt.Printf("%s", p)
	return len(p), nil
}

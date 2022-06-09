package main

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

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

// Common options, regardless of profile/trace mode.
var options = struct {
	n int           // iterations
	t time.Duration // step delay
}{
	n: 10,
	t: 0 * time.Second,
}

// mainFunc is the prototype for a pseudo-"main" function. A function with this
// signature is passed to (and called by) function run, which is conditionally
// defined according to build tags.
type mainFunc func() error

// main simply calls function run, which is defined according to the given build
// tags. Function run initializes its particular runtime support and then calls
// function app (below), which is effectively the real main function.
// Available tags are:
//
// trace: Traces program execution with package "runtime/trace"
//   - go build -tags='trace' -gcflags='-m -l' ./examples/stdio/
//   - ./stdio [-o=stdio.trace]
//
// pprof: Profiles program performance with package "runtime/pprof"
//   - go build -tags='pprof' -gcflags='-m -l' ./examples/stdio/
//   - ./stdio [-o=stdio.profile]
//   - Interactive (CLI): go tool pprof stdio.profile
//   - Interactive (Web): go tool pprof -http=localhost:8080 stdio.profile
//
// pprof_http: pprof but with real-time web server (requires tag pprof)
//   - go build -tags='pprof,pprof_http' -gcflags='-m -l' ./examples/stdio/
//   - ./stdio [-addr=localhost:8080]
//   - Navigation: http://localhost:8080/debug/pprof
//   - Export, e.g., heap profile: curl -sK -v http://localhost:8080/debug/pprof/heap > heap.profile
func main() { log.Println(run(Main)) }

// Main is the real "main" function, which is called by function run in order to
// wrap any trace/profile support capabilities requested via build tags.
func Main() error {
	// Put the terminal to which stdin is attached into "raw" mode. This prevents
	// any line buffering or line disciplines implemented by the user's terminal.
	//
	// This is required to allow Embedit to process keypresses and other events
	// immediately when they occur. Otherwise, the program would not receive any
	// input until the terminal sends it (typically on Return, CR/LF).
	f := sys.MakeFdio(int(os.Stdin.Fd()))
	if !f.Valid() || !f.Raw() {
		return errors.New("cannot attach terminal to stdin")
	}
	defer f.Restore()

	em.Configure(embedit.Config{RW: rw, Width: 80, Height: 24})

	em.Line().SetPos([]rune("hello there"), 7)
	// time.Sleep(options.t)
	em.Line().ErasePrevious(3)
	// time.Sleep(options.t)
	em.Line().Set([]rune("wat"))
	// time.Sleep(options.t)

	// em.Line().SetPos([]rune("hello there"), 9)
	// time.Sleep(options.t)
	// em.Line().ErasePrevious(3)
	// time.Sleep(options.t)
	// em.Line().Set([]rune("wat"))
	// time.Sleep(options.t)

	return nil
}

package main

import (
	"io"
	"os"
	"time"

	"github.com/ardnew/embedit"
	"github.com/ardnew/embedit/sys"
	"github.com/ardnew/embedit/terminal/key"
)

const binName = "analysis"

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
	n: 10000,
	t: 1 * time.Millisecond,
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
//   - go build -tags='history,trace' -gcflags='github.com/ardnew/embedit/...=-m -l' ./examples/analysis/
//   - ./analysis [-o=analysis.trace]
//
// pprof: Profiles program performance with package "runtime/pprof"
//   - go build -tags='history,pprof' -gcflags='github.com/ardnew/embedit/...=-m -l' ./examples/analysis/
//   - ./analysis [-o=analysis.profile]
//   - Interactive (CLI): go tool pprof -alloc_space -lines -nodefraction=0 -edgefraction=0 analysis.profile
//   - Interactive (Web): go tool pprof -http=:8080 -alloc_space -lines -nodefraction=0 -edgefraction=0 analysis.profile
//
// pprof_http: pprof but with real-time web server (requires tag pprof)
//   - go build -tags='history,pprof,pprof_http' -gcflags='github.com/ardnew/embedit/...=-m -l' ./examples/analysis/
//   - ./analysis [-addr=localhost:8080]
//   - Navigation: http://localhost:8080/debug/pprof
//   - Export, e.g., heap profile: curl -sK -v http://localhost:8080/debug/pprof/heap > heap.profile
func main() { run(Main) }

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
		return os.ErrInvalid
	}
	defer f.Restore()

	em.Configure(embedit.Config{RW: rw, Width: 80, Height: 24, AutoFlush: true})

	for i := 0; i < options.n; i++ {
		em.Line().InsertRune('A')
		em.Line().SetAndMoveCursorTo([]rune("hello testing there"), 10)
		em.Terminal().HandleKey(key.ClearScreen)
		em.Line().SetAndMoveCursorTo([]rune("there testing hello"), 8)
		em.Terminal().HandleKey(key.Enter)
		em.Line().SetAndMoveCursorTo([]rune("  hello testing there"), -1)
		em.Terminal().HandleKey(key.Enter)
		em.Terminal().HandleKey(key.Up)
		em.Terminal().HandleKey(key.Up)
		em.Terminal().HandleKey(key.Down)
		em.Line().MoveCursorTo(7)
		em.Terminal().HandleKey(key.DeleteWord)
		em.Terminal().HandleKey(key.DeleteWord)
		em.Terminal().HandleKey(key.EndOfFile)
		em.Terminal().HandleKey(key.AltRight)
		em.Terminal().HandleKey(key.EndOfFile)
		em.Terminal().HandleKey(key.End)
		em.Terminal().HandleKey(key.Left)
		em.Terminal().HandleKey(key.Backspace)
		em.Terminal().HandleKey(key.AltLeft)
		em.Terminal().HandleKey(key.Home)
		em.Terminal().HandleKey(key.AltRight)
		em.Terminal().HandleKey(key.Right)
		em.Terminal().HandleKey(key.Kill)
		em.Terminal().HandleKey(key.AltLeft)
		em.Terminal().HandleKey(key.Left)
		em.Terminal().HandleKey(key.KillPrevious)
		em.Line().InsertRune('X')
		em.Line().Set([]rune("wat"))
		time.Sleep(options.t)
	}
	return nil
}

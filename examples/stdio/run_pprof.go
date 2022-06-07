//go:build pprof && !pprof_http
// +build pprof,!pprof_http

package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/pprof"
)

var flags = struct {
	o string // output file
}{
	o: pkgName + ".profile",
}

func init() {
	runtime.MemProfileRate = 1
}

func run(fn mainFunc) (err error) {
	if err = parseFlags(); err != nil {
		return
	}
	if err = fn(); err != nil {
		return
	}
	var f *os.File
	if f, err = os.Create(flags.o); err != nil {
		return
	}
	defer func() { err = f.Close() }()
	runtime.GC() // Get up-to-date statistics
	return pprof.WriteHeapProfile(f)
}

func parseFlags() (err error) {
	fs := flag.NewFlagSet(pkgName, flag.ExitOnError)

	fs.IntVar(&options.n, "n", options.n,
		"Number of `iterations`")

	fs.DurationVar(&options.t, "t", options.t,
		"Wait `delay` between steps")

	fs.StringVar(&flags.o, "o", flags.o,
		"Write output to `file`")

	return fs.Parse(os.Args[1:])
}

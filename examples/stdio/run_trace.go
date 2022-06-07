//go:build trace
// +build trace

package main

import (
	"flag"
	"os"
	"runtime/trace"
)

var flags = struct {
	o string
}{
	o: pkgName + ".trace",
}

func run(fn mainFunc) (err error) {
	if err = parseFlags(); err != nil {
		return
	}
	var f *os.File
	if f, err = os.Create(flags.o); err != nil {
		return
	}
	defer func() { err = f.Close() }()
	if err = trace.Start(f); err != nil {
		return
	}
	defer trace.Stop()
	return fn()
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

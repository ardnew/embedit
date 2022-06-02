//go:build pprof
// +build pprof

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
)

func run(app mainFunc) (err error) {
	err = app()

	var f *os.File
	if f, err = os.Create("mem.pro"); err != nil {
		return
	}
	defer func() { err = f.Close() }()

	runtime.GC() // get up-to-date statistics
	return pprof.WriteHeapProfile(f)
}

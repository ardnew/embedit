//go:build trace
// +build trace

package main

import (
	"os"
	"runtime/trace"
)

func run(app mainFunc) (err error) {
	var f *os.File
	if f, err = os.Create("trace.out"); err != nil {
		return
	}
	defer func() { err = f.Close() }()

	if err = trace.Start(f); err != nil {
		return
	}
	defer trace.Stop()

	return app()
}

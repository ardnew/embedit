//go:build windows
// +build windows

package sysio

import "golang.org/x/sys/windows"

type fdio struct {
	mode uint32
}

func makeFdio(fd int) Fdio {
	f := Fdio{fd: fd}
	f.valid = f.read()
	return f
}

func (f *Fdio) read() (ok bool) {
	return windows.GetConsoleMode(windows.Handle(f.fd), &f.mode) == nil
}

func (f *Fdio) write(mode uint32) bool {
	return windows.SetConsoleMode(windows.Handle(f.fd), mode) == nil
}

func (f *Fdio) raw() bool {
	// Using f.read() here will cause f.restore() to always restore to the state
	// prior to f.raw() being called, instead of using the initial state obtained
	// when makeFdio() was called.
	if !f.valid {
		return false
	}
	// Copy the receiver's mode instead of manipulating it. This allows us to
	// restore the previous state at a later time.
	return f.write(f.mode &^ (
		windows.ENABLE_ECHO_INPUT |
		windows.ENABLE_PROCESSED_INPUT |
		windows.ENABLE_LINE_INPUT |
		windows.ENABLE_PROCESSED_OUTPUT,
	))
}

func (f *Fdio) restore() error {
	if !f.valid {
		return false
	}
	// The mode field of Fdio is never modified except for initialization or when
	// users call Save directly.
	// Hence, we can restore to a prior state at any time.
	return f.set(f.mode)
}

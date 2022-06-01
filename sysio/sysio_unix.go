//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package sysio

import (
	"github.com/ardnew/embedit/sysio/sys"
)

type fdio struct {
	termios sys.Termios
}

func makeFdio(fd int) Fdio {
	f := Fdio{fd: fd}
	f.valid = f.read()
	return f
}

func (f *Fdio) read() (ok bool) {
	f.termios, ok = sys.GetTermios(f.fd)
	return
}

func (f *Fdio) write(ts sys.Termios) bool {
	return sys.SetTermios(f.fd, ts)
}

func (f *Fdio) raw() bool {
	// Using f.read() here will cause f.restore() to always restore to the state
	// prior to f.raw() being called, instead of using the initial state obtained
	// when makeFdio() was called.
	if !f.valid {
		return false
	}
	// Raw copies the receiver's termios instead of manipulating it, which allows
	// us to restore the previous state at a later time.
	return f.write(f.termios.Raw())
}

func (f *Fdio) restore() bool {
	if !f.valid {
		return false
	}
	// The termios field of Fdio is never modified except for initialization or
	// when users call Save directly.
	// Hence, we can restore to a prior state at any time.
	return f.write(f.termios)
}

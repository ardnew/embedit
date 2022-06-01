//go:build darwin || freebsd
// +build darwin freebsd

package unix

import "golang.org/x/sys/unix"

const (
	getTermios = unix.TIOCGETA
	setTermios = unix.TIOCSETA
)

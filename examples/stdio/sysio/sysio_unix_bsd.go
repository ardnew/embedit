//go:build darwin || freebsd
// +build darwin freebsd

package sysio

import "golang.org/x/sys/unix"

const (
	ioctlReadTermios  = unix.TIOCGETA
	ioctlWriteTermios = unix.TIOCSETA
)

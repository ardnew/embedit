//go:build linux
// +build linux

package unix

import "golang.org/x/sys/unix"

const (
	getTermios = unix.TCGETS
	setTermios = unix.TCSETS
)

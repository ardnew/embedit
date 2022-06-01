//go:build darwin && !386
// +build darwin,!386

package sys

const (
	ioctlReadTermios  = 0x40487413 // TIOCGETA
	ioctlWriteTermios = 0x80487414 // TIOCSETA
)

//go:build (darwin && 386) || freebsd
// +build darwin,386 freebsd

package sys

const (
	ioctlReadTermios  = 0x402C7413 // TIOCGETA
	ioctlWriteTermios = 0x802C7414 // TIOCSETA
)

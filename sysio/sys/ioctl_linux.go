//go:build linux
// +build linux

package sys

const (
	ioctlReadTermios  = 0x5401 // TCGETS
	ioctlWriteTermios = 0x5402 // TCSETS
)

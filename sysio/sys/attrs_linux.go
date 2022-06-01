//go:build linux
// +build linux

package sys

const (
	// IFLAG
	ignbrk = 0x1
	brkint = 0x2
	parmrk = 0x8
	istrip = 0x20
	inlcr  = 0x40
	igncr  = 0x80
	icrnl  = 0x100
	ixon   = 0x400
	// OFLAG
	opost = 0x1
	// CFLAG
	cs8    = 0x30
	csize  = 0x30
	parenb = 0x100
	// LFLAG
	isig   = 0x1
	icanon = 0x2
	echo   = 0x8
	echonl = 0x40
	iexten = 0x8000
	// CC
	vmin  = 0x6
	vtime = 0x5
)

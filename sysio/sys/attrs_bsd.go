//go:build darwin || freebsd
// +build darwin freebsd

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
	ixon   = 0x200
	// OFLAG
	opost = 0x1
	// CFLAG
	cs8    = 0x300
	csize  = 0x300
	parenb = 0x1000
	// LFLAG
	echo   = 0x8
	echonl = 0x10
	isig   = 0x80
	icanon = 0x100
	iexten = 0x400
	// CC
	vmin  = 0x10
	vtime = 0x11
)

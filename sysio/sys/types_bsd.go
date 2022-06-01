//go:build (darwin && 386) || (darwin && arm) || freebsd
// +build darwin,386 darwin,arm freebsd

package sys

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}

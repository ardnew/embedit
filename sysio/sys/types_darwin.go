//go:build darwin && !386 && !arm
// +build darwin,!386,!arm

package sys

type Termios struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]uint8
	_      [4]byte
	Ispeed uint64
	Ospeed uint64
}

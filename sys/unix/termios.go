//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package unix

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Termios is a wrapper for the type provided by golang.org/x/sys/unix.
type Termios struct {
	unix.Termios
}

// Raw returns a copy of the receiver Termios modified for raw terminal mode.
func (t Termios) Raw() Termios {
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK |
		unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	t.Oflag &^= unix.OPOST
	t.Cflag &^= unix.CSIZE | unix.PARENB
	t.Cflag |= unix.CS8
	t.Lflag &^= unix.ECHO | unix.ECHONL |
		unix.ISIG | unix.ICANON | unix.IEXTEN
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0
	return t
}

func SetTermios(fd int, value Termios) (ok bool) {
	ok = ioctl(fd, setTermios, uintptr(unsafe.Pointer(&value.Termios))) == retOK
	runtime.KeepAlive(value)
	return
}

func GetTermios(fd int) (value Termios, ok bool) {
	ok = ioctl(fd, getTermios, uintptr(unsafe.Pointer(&value.Termios))) == retOK
	return
}

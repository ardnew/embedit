//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package sys

import (
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const retOK = 0

func ioctl(fd int, req uint, arg uintptr) (ret uintptr) {
	_, _, e1 := syscall.Syscall(sysIoctl, uintptr(fd), uintptr(req), uintptr(arg))
	return uintptr(e1)
}

// SetTermios performs an ioctl on fd with a *Termios.
func SetTermios(fd int, value unix.Termios) (ok bool) {
	ok = ioctl(fd, ioctlWriteTermios, uintptr(unsafe.Pointer(&value))) == retOK
	runtime.KeepAlive(value)
	return
}

func GetTermios(fd int) (value Termios, ok bool) {
	ok = ioctl(fd, ioctlReadTermios, uintptr(unsafe.Pointer(&value))) == retOK
	return
}

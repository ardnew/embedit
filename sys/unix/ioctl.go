//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package unix

import "golang.org/x/sys/unix"

const retOK = 0

func ioctl(fd int, req uint, arg uintptr) (ret uintptr) {
	_, _, e1 := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	return uintptr(e1)
}

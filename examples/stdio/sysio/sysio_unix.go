//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package sysio

import "golang.org/x/sys/unix"

type state struct {
	termios unix.Termios
}

func newState(fd int) *State {
	st := &State{fd: fd}
	ts, err := st.get()
	if st.valid = (err == nil); !st.valid {
		return nil
	}
	st.termios = *ts
	return st
}

func (st *State) get() (*unix.Termios, error) {
	return unix.IoctlGetTermios(st.fd, ioctlReadTermios)
}

func (st *State) set(ts *unix.Termios) error {
	return unix.IoctlSetTermios(st.fd, ioctlWriteTermios, ts)
}

func (st *State) raw() error {
	if !st.valid {
		return unix.EINVAL
	}
	// Copy the receiver's termios instead of manipulating it. This allows us to
	// restore the previous state at a later time.
	termios := st.termios
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	return st.set(&termios)
}

func (st *State) restore() error {
	if !st.valid {
		return unix.EINVAL
	}
	// The termios field of State is never modified except for initialization.
	// Hence, we can restore to this initial state at any time.
	return st.set(&st.termios)
}

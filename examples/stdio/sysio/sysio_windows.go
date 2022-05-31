//go:build windows
// +build windows

package sysio

import "golang.org/x/sys/windows"

type state struct {
	mode uint32
}

func newState(fd int) *State {
	st := &State{fd: fd}
	mode, err := st.get()
	if st.valid = (err == nil); !st.valid {
		return nil
	}
	st.mode = mode
	return st
}

func (st *State) get() (mode uint32, err error) {
	err = windows.GetConsoleMode(windows.Handle(st.fd), &mode)
	return
}

func (st *State) set(mode uint32) error {
	return windows.SetConsoleMode(windows.Handle(st.fd), mode)
}

func (st *State) raw() error {
	if !st.valid {
		return windows.ERROR_INVALID_HANDLE
	}
	// Copy the receiver's mode instead of manipulating it. This allows us to
	// restore the previous state at a later time.
	return st.set(st.mode &^ (
		windows.ENABLE_ECHO_INPUT |
		windows.ENABLE_PROCESSED_INPUT |
		windows.ENABLE_LINE_INPUT |
		windows.ENABLE_PROCESSED_OUTPUT,
	))
}

func (st *State) restore() error {
	if !st.valid {
		return windows.ERROR_INVALID_HANDLE
	}
	// The mode field of State is never modified except for initialization.
	// Hence, we can restore to this initial state at any time.
	return st.set(st.mode)
}

package sysio

// State contains the state of a terminal.
type State struct {
	state
	fd    int
	valid bool // Has the system initialized this state?
}

// NewState returns the state of the given file descriptor.
func NewState(fd int) *State {
	return newState(fd)
}

// MakeRaw puts the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
func (st *State) Raw() error {
	return st.raw()
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func (st *State) Restore() error {
	return st.restore()
}

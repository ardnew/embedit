package paste

// State represents the state of bracketed paste mode.
type State byte

// Constants values of enumerated type State.
const (
	Active State = iota
	Inactive
)

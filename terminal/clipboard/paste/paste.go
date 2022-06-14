package paste

// State represents the state of bracketed paste mode.
type State byte

// Constants values of enumerated type State.
const (
	Active State = iota
	Inactive
)

// IsActive returns true if and only if s is Active.
func (s State) IsActive() bool {
	return s == Active
}

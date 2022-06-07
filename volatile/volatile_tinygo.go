//go:build tinygo
// +build tinygo

package volatile

import "runtime/volatile"

// Registers are implemented in the TinyGo runtime package "runtime/volatile".
type (
	Register8  = volatile.Register8
	Register32 = volatile.Register32
)

//go:build tinygo
// +build tinygo

package volatile

import "runtime/volatile"

// Register32 is implemented in the TinyGo runtime package "runtime/volatile".
type Register32 = volatile.Register32

//go:build !tinygo

package volatile

// Register32 defines an unsigned 32-bit integer with atomic mutability.
type Register32 uint32

// Get returns the uint32 value of r.
func (r Register32) Get() uint32 {
	return uint32(r)
}

// Set sets the uint32 value of r to v.
func (r *Register32) Set(v uint32) {
	*r = Register32(v)
}

package sysio

// Fdio contains the state of a terminal connected to a file descriptor.
type Fdio struct {
	fd int
	fdio
	valid bool
}

// MakeFdio returns the state of the terminal connected to the given file
// descriptor.
func MakeFdio(fd int) Fdio {
	return makeFdio(fd)
}

// Fd returns the file descriptor to which a terminal was configured to connect.
func (f *Fdio) Fd() int {
	return f.fd
}

// Valid returns true if and only if the state of a terminal connected to f's
// file descriptor was read successfully.
func (f *Fdio) Valid() bool {
	return f.valid
}

// Save stores the current state of the terminal connected to f's file
// descriptor. This state can later be restored via receiver's Restore method.
//
// Calling Save will overwrite the state stored with MakeFdio or any prior call
// to Save.
func (f *Fdio) Save() bool {
	f.valid = f.read()
	return f.valid
}

// Raw puts the terminal connected to f's file descriptor into raw mode.
// The terminal's prior mode can be restored via receiver's Restore method.
func (f *Fdio) Raw() bool {
	return f.raw()
}

// Restore puts the terminal connected to f's file descriptor into its prior
// mode; i.e., the mode read during MakeFdio or Save, whichever is called last.
func (f *Fdio) Restore() bool {
	return f.restore()
}

//go:build linux || darwin || freebsd
// +build linux darwin freebsd

package sys

// Raw returns a copy of the receiver Termios modified for raw terminal mode.
func (t Termios) Raw() Termios {
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	t.Iflag &^= ignbrk | brkint | parmrk | istrip | inlcr | igncr | icrnl | ixon
	t.Oflag &^= opost
	t.Cflag &^= csize | parenb
	t.Cflag |= cs8
	t.Lflag &^= echo | echonl | isig | icanon | iexten
	t.Cc[vmin] = 1
	t.Cc[vtime] = 0
	return t
}

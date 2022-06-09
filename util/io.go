package util

import "io"

// EOFMask wraps an io.Reader to hide io.EOF errors from the caller.
//
// Either wrap individual calls with EOFMask{r}.Read(...), or replace the
// Reader (e.g., r = EOFMask{r}) to mask all Read calls in that scope.
type EOFMask struct{ io.Reader }

// Read ignores io.EOF errors returned from its Reader's implementation, and
// returns nil instead. The returned number of bytes are unmodified.
//
// See documentation on io.ReaderFrom, and io.Copy.
func (i EOFMask) Read(p []byte) (n int, err error) {
	if n, err = i.Reader.Read(p); err == io.EOF {
		err = nil
	}
	return
}

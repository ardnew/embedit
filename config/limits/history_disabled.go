//go:build !history
// +build !history

package limits

// LinesPerHistory defines the maximum number of lines stored in history.
// Old lines are discarded as more than LinesPerHistory are added.
//
// By omitting build tag "history", you can disable the line history capability,
// and as a result eliminate nearly all unessential statically-allocated memory
// in Terminal objects.
//
// This may be desirable for use cases where only a single line of input is
// being read (e.g., password prompt, complex user dialog), or when the package
// is being used strictly for its terminal emulation to manipulate the cursor
// or display (line drawing, progress meter, interactive menu, etc.).
const LinesPerHistory = 0

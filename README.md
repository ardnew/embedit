# embedit
##### CLI line editing for embedded devices

`embedit` is a Go module for emulating a subset of ANSI/VT100 terminal capabilities. The principle design goals are:
1. Usable with standard Go and TinyGo
2. No dynamic memory allocation
   - For operation on baremetal and embedded systems such as microcontrollers
3. Minimal dependencies (no `fmt`, `strconv`, or any third-party packages)
   - Required to guarantee goals 1 & 2

Much of the original source code was based on [golang.org/x/term](https://cs.opensource.google/go/x/term) by the standard Go authors.

## Usage
##### TBD
 - Line navigation (<kbd>←</kbd>/<kbd>^B</kbd>, <kbd>→</kbd>/<kbd>^F</kbd>, <kbd>Home</kbd>/<kbd>^A</kbd>, <kbd>End</kbd>/<kbd>^E</kbd>)
 - History (<kbd>↑</kbd>/<kbd>^P</kbd>, <kbd>↓</kbd>/<kbd>^N</kbd>)
 - Editing (<kbd>⌫ Backspace</kbd>/<kbd>^H</kbd>, <kbd>Delete</kbd>/<kbd>^D</kbd>)
 - Clipboard cut (word <kbd>^W</kbd>, line <kbd>^U</kbd>, right <kbd>^K</kbd>), paste (<kbd>^Y</kbd>)
 - Control redraw (<kbd>^L</kbd>)

# embedit
##### CLI line editing for embedded devices

`embedit` is a Go module for reading and editing a line of input (similar to [readline](https://en.wikipedia.org/wiki/GNU_Readline)). The principle design goals are:
1. Usable with standard Go and TinyGo
2. No dynamic memory allocation
   - For operation on baremetal and embedded systems such as microcontrollers
3. Minimal dependencies (no `fmt`, `strconv`, or any third-party packages)
   - Required to guarantee goals 1 & 2

Much of the original source code was based on [golang.org/x/term](https://cs.opensource.google/go/x/term) by the standard Go authors.

## Usage
##### TBD

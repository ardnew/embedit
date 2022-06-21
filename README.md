# embedit
##### CLI line editing for embedded devices

`embedit` is a Go module for reading and editing a line of input (similar to [readline](https://en.wikipedia.org/wiki/GNU_Readline)). The principle design goals are:
1. Usable with standard Go and TinyGo
2. No dynamic memory allocation (see [analysis below](#heap-profile))
   - For operation on baremetal and embedded systems such as microcontrollers
3. Minimal dependencies (no `fmt`, `strconv`, or any third-party packages)
   - Required to guarantee goals 1 & 2

Much of the original source code was based on [`golang.org/x/term`](https://cs.opensource.google/go/x/term) by the standard Go authors.

## Usage

See [`examples`](examples) for a few different usages. The following gif shows [`examples/basic`](examples/basic) in action.

![examples/basic](examples/basic/embedit-basic.gif)

## Heap Profile

The following tools are used to verify no dynamic memory allocation is performed:

 - [pprof](https://github.com/google/pprof) from standard Go
 - [trace](https://pkg.go.dev/runtime/trace) from standard Go
 - TinyGo's built-in escape analysis compile-time option (`-print-allocs`)

The last is straight-forward, so only `pprof` and `trace` are discussed here. 

The [`examples/analysis`](examples/analysis) demo is used to exercise the package API. Different build tags will configure the individual profiling tools. See the comments in [`examples/analysis/main.go`](examples/analysis/main.go) for details.

So far, runtime heap allocation can only be eliminated from the package code and its direct dependencies. In standard Go, the runtime itself will perform heap allocation internally to create and manage core goroutines and system threads (e.g., [`G`s and `M`s](https://go.dev/src/runtime/HACKING)).

The following graph shows the total heap memory allocated throughout the lifetime of the program. 

![Heap profile (examples/analysis)](docs/analysis.pprof.svg)

As you can see, most of the memory is attributed to the internal runtime constructs mentioned above. All remaining allocations are related to the main package of the `analysis` example such as the profiler tooling and command-line flag parser. None of these are directly related to the `embedit` package itself.

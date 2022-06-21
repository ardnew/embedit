#!/bin/sh

# Build the executable with pprof support
#   (writes escape analysis to analysis.pprof.build)
go build -o=./analysis.pprof -tags=pprof,history -gcflags='github.com/ardnew/embedit/...=-l -m' ../examples/analysis/ &>analysis.pprof.build

# Run the executable and write the pprof profile to analysis.pprof.profile
./analysis.pprof -o=analysis.pprof.profile -n=10000 -t=0

# Generate the pprof reports (tree, peek, and call-graph SVG)
go tool pprof -alloc_space -add_comment="n = 10000, t = 0" -filefunctions -nodefraction=0 -edgefraction=0 -source_path=/usr/local/go/dev/src:.. -output=analysis.pprof.tree -tree analysis.pprof analysis.pprof.profile
go tool pprof -alloc_space -add_comment="n = 10000, t = 0" -filefunctions -nodefraction=0 -edgefraction=0 -source_path=/usr/local/go/dev/src:.. -output=analysis.pprof.peek -peek=. analysis.pprof analysis.pprof.profile
go tool pprof -alloc_space -add_comment="n = 10000, t = 0" -filefunctions -nodefraction=0 -edgefraction=0 -source_path=/usr/local/go/dev/src:.. -output=analysis.pprof.svg -svg analysis.pprof analysis.pprof.profile
go tool pprof -alloc_space -add_comment="n = 10000, t = 0" -filefunctions -nodefraction=0 -edgefraction=0 -source_path=/usr/local/go/dev/src:.. -output=analysis.pprof.png -png analysis.pprof analysis.pprof.profile

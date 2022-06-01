//go:build pprof
// +build pprof

package main

import (
	"net/http"
	_ "net/http/pprof"
)

func profile() {
	http.ListenAndServe("localhost:8080", nil)
}

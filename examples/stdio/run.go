//go:build !trace && !pprof && !pprof_http
// +build !trace,!pprof,!pprof_http

package main

func run(app mainFunc) error { return app() }

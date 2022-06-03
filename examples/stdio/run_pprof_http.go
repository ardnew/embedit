//go:build pprof_http
// +build pprof_http

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func run(app mainFunc) (err error) {
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()

	err = app()

	return
}

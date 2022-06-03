//go:build pprof && pprof_http
// +build pprof,pprof_http

package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var flags = struct {
	addr string
}{
	addr: "localhost:8080",
}

func run(fn mainFunc) (err error) {
	if err = parseFlags(); err != nil {
		return
	}
	done := make(chan bool, 1)
	go func(c chan bool) {
		log.Println(http.ListenAndServe(flags.addr, nil))
		c <- true
	}(done)
	err = fn()
	<-done
	return
}

func parseFlags() (err error) {
	fs := flag.NewFlagSet(pkgName, flag.ExitOnError)

	fs.StringVar(&flags.addr, "addr", flags.addr,
		"Bind pprof http server to `address:port`")

	return fs.Parse(os.Args[1:])
}

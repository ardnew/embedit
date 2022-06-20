package main

import (
	"machine"

	"github.com/ardnew/embedit"
)

// Static storage for our main object.
var em embedit.Embedit

func main() {
	// Use the target device's default serial port (machine.Serial) for physical
	// read/write byte transfers.
	//
	// The machine.Serial device is configured at compile-time in TinyGo with
	// command-line flag -serial=(none|uart|usb).
	em.Configure(embedit.Config{RW: machine.Serial, Width: 80, Height: 24})
	for {
		em.Terminal().ReadLine()
	}
}

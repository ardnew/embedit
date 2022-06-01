//go:build (linux && !amd64 && !arm64) || darwin || freebsd
// +build linux,!amd64,!arm64 darwin freebsd

package sys

const sysIoctl = 54

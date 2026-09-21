//go:build linux

package main

import "syscall"

// The two requests that read and write a terminal's settings on this
// system.
const (
	getAttributes = syscall.TCGETS
	setAttributes = syscall.TCSETS
)

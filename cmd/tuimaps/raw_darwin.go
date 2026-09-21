//go:build darwin

package main

import "syscall"

// The two requests that read and write a terminal's settings on this
// system. They are spelt differently from system to system and mean the
// same thing.
const (
	getAttributes = syscall.TIOCGETA
	setAttributes = syscall.TIOCSETA
)

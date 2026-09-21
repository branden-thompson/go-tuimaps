//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// winsize is what the terminal answers with: its size in cells, and in
// pixels where it knows that, which this app does not use.
type winsize struct {
	rows, cols     uint16
	xpixel, ypixel uint16
}

// fromTerminal asks the terminal how big it is. A run whose output is a
// file or a pipe has no terminal to ask and says so rather than guessing.
func fromTerminal() (int, int, bool) {
	var size winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if errno != 0 || size.cols == 0 || size.rows == 0 {
		return 0, 0, false
	}
	return int(size.cols), int(size.rows), true
}

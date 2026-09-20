//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// consoleInfo is what the Windows console answers with. Only the window
// itself is read: the buffer behind it may be far taller than the window,
// and a map drawn to the buffer's height would not fit on the screen.
type consoleInfo struct {
	size              [2]int16
	cursor            [2]int16
	attributes        uint16
	left, top         int16
	right, bottom     int16
	maximumWindowSize [2]int16
}

// fromTerminal asks the console how big its window is.
func fromTerminal() (int, int, bool) {
	kernel := syscall.NewLazyDLL("kernel32.dll")
	ask := kernel.NewProc("GetConsoleScreenBufferInfo")
	var info consoleInfo
	ok, _, _ := ask.Call(os.Stdout.Fd(), uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return 0, 0, false
	}
	cols := int(info.right) - int(info.left) + 1
	rows := int(info.bottom) - int(info.top) + 1
	if cols < 1 || rows < 1 {
		return 0, 0, false
	}
	return cols, rows, true
}

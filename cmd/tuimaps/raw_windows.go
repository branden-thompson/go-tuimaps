//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// The console modes this app changes. Reading a key as it is pressed means
// turning off the line editor and the echo; drawing a map means asking the
// console to read the escape sequences the library writes.
const (
	processedInput     = 0x0001
	lineInput          = 0x0002
	echoInput          = 0x0004
	virtualTerminalIn  = 0x0200
	virtualTerminalOut = 0x0004
	newlineAutoReturn  = 0x0008
	enableProcessedOut = 0x0001
	enableWrapAtEolOut = 0x0002
)

// state is what the console's input and output were set to when the app
// found them, kept so that they can be put back exactly.
type state struct {
	in, out uint32
}

// raw hands each key over as it is pressed and turns on the console's
// handling of escape sequences.
func raw() (*state, error) {
	in, err := consoleMode(os.Stdin)
	if err != nil {
		return nil, err
	}
	out, err := consoleMode(os.Stdout)
	if err != nil {
		return nil, err
	}
	kept := &state{in: in, out: out}
	wanted := in&^(processedInput|lineInput|echoInput) | virtualTerminalIn
	if err := setConsoleMode(os.Stdin, wanted); err != nil {
		return nil, err
	}
	if err := setConsoleMode(os.Stdout, out|virtualTerminalOut|enableProcessedOut|enableWrapAtEolOut); err != nil {
		setConsoleMode(os.Stdin, in)
		return nil, err
	}
	return kept, nil
}

// unraw puts the console back as it was found.
func unraw(kept *state) {
	if kept == nil {
		return
	}
	setConsoleMode(os.Stdin, kept.in)
	setConsoleMode(os.Stdout, kept.out)
}

// consoleMode is what one handle is set to.
func consoleMode(file *os.File) (uint32, error) {
	if file == nil {
		return 0, syscall.EINVAL
	}
	var mode uint32
	ask := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleMode")
	ok, _, err := ask.Call(file.Fd(), uintptr(unsafe.Pointer(&mode)))
	if ok == 0 {
		return 0, err
	}
	return mode, nil
}

// setConsoleMode sets one handle.
func setConsoleMode(file *os.File, mode uint32) error {
	if file == nil {
		return syscall.EINVAL
	}
	tell := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
	ok, _, err := tell.Call(file.Fd(), uintptr(mode))
	if ok == 0 {
		return err
	}
	return nil
}

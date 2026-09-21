//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// state is the terminal's settings as they were found, kept whole so that
// they can be put back exactly rather than guessed at.
type state struct {
	was syscall.Termios
}

// raw hands each key over as it is pressed: no line buffering, no echo, and
// no signal characters - the app answers the interrupt character itself, so
// that there is one way out and it always puts the terminal back (L-16).
func raw() (*state, error) {
	var settings syscall.Termios
	if err := attributes(getAttributes, &settings); err != nil {
		return nil, err
	}
	kept := &state{was: settings}
	settings.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	settings.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	settings.Cc[syscall.VMIN] = 1
	settings.Cc[syscall.VTIME] = 0
	if err := attributes(setAttributes, &settings); err != nil {
		return nil, err
	}
	return kept, nil
}

// unraw puts the settings back as they were found.
func unraw(kept *state) {
	if kept == nil {
		return
	}
	settings := kept.was
	attributes(setAttributes, &settings)
}

// attributes reads or writes the terminal's settings.
func attributes(what uintptr, settings *syscall.Termios) error {
	if settings == nil {
		return syscall.EINVAL
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdin.Fd(), what, uintptr(unsafe.Pointer(settings)))
	if errno != 0 {
		return errno
	}
	return nil
}

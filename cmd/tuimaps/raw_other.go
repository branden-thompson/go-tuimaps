//go:build !darwin && !linux && !windows

package main

import "errors"

// state is nothing on a system whose terminal this app does not know how to
// take over.
type state struct{}

// raw says plainly that the interactive map cannot run here. The modes that
// need no terminal - one frame, or the description in words - still do.
func raw() (*state, error) {
	return nil, errors.New("this system's terminal is one the app cannot take over; --headless and --describe still work")
}

// unraw has nothing to put back.
func unraw(_ *state) {}

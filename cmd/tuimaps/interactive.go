package main

import (
	"errors"
	"io"
)

// interactive is the map a person moves about with the keys. It is built in
// the next step of this work package - the key map, the chrome, the app's
// own pump and the draw loop - and until then the app says so plainly
// rather than opening a terminal it cannot yet give back.
func interactive(_ settings, _, errs io.Writer) int {
	return complain(errs, errors.New(
		"the interactive map is not built yet; --headless draws one frame and --describe says where things are"),
		exitFailed)
}

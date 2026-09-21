package main

import (
	"os"
	"strconv"
)

// The size a map is drawn at when there is no terminal to ask and nothing
// was said: upstream's own, and the size of a terminal that has never been
// resized.
const (
	defaultCols = 80
	defaultRows = 24
)

// terminalSize is how big the map should be drawn. The library draws the
// exact rectangle it is given and reserves nothing of its own (P-53), so
// this is the app's decision and the app's alone.
//
// What the person wrote wins; then what the terminal says it is; then what
// the environment says, which is what a terminal that cannot be asked - a
// pipe, a test, a continuous-integration run - is left with.
func terminalSize() (int, int) {
	if cols, rows, ok := fromTerminal(); ok {
		return cols, rows
	}
	cols, rows := fromEnvironment()
	if cols > 0 && rows > 0 {
		return cols, rows
	}
	return defaultCols, defaultRows
}

// fromEnvironment reads COLUMNS and LINES, which a shell sets and a person
// can set by hand.
func fromEnvironment() (int, int) {
	cols, err := strconv.Atoi(os.Getenv("COLUMNS"))
	if err != nil || cols < 1 {
		return 0, 0
	}
	rows, err := strconv.Atoi(os.Getenv("LINES"))
	if err != nil || rows < 1 {
		return 0, 0
	}
	return cols, rows
}

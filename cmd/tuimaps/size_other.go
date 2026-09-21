//go:build !darwin && !linux && !windows

package main

// fromTerminal has no way to ask on this system, and says so rather than
// answering a size nobody measured. The environment and the default stand
// behind it.
func fromTerminal() (int, int, bool) { return 0, 0, false }

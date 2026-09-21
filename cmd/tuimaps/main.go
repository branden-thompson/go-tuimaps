// Command tuimaps draws a map of the world in a terminal.
//
// It is the library's first host, and it is written the way any host is: it
// owns the terminal, the keys and the clock, it runs the library's work on
// its own goroutines, and it asks the library only for cells and for facts.
// Nothing here is a special arrangement the library makes for it.
package main

import (
	"fmt"
	"io"
	"os"
)

// The exit codes. A person reading them in a script should not have to
// guess: nothing wrong, something the person wrote, something that failed.
const (
	exitFine    = 0
	exitMistake = 2
	exitFailed  = 1
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the whole app, with its two ends handed in so that a test can read
// what it wrote.
func run(args []string, out, errs io.Writer) int {
	if out == nil || errs == nil {
		return exitFailed // there is nowhere to say anything, so nothing is said
	}
	s, err := parse(args)
	if err != nil {
		return complain(errs, err, exitMistake)
	}
	if s.help {
		fmt.Fprint(out, helpText())
		return exitFine
	}
	if err := s.check(); err != nil {
		return complain(errs, err, exitMistake)
	}
	switch {
	case s.purge || s.verify:
		return maintain(s, out, errs)
	case s.describe:
		return describing(s, out, errs)
	case s.headless:
		return headless(s, out, errs)
	}
	return interactive(s, out, errs)
}

// complain writes one clear line about what went wrong. Nothing the app
// prints is text a terminal could be made to act on (FR-34).
func complain(errs io.Writer, err error, code int) int {
	if err == nil {
		return code
	}
	text := err.Error()
	if !clean(text) {
		text = "the app was given something it cannot print back"
	}
	fmt.Fprintln(errs, "tuimaps: "+text)
	return code
}

// check is what can be said about the flags only once they are all read:
// which of them cannot be asked for together.
func (s settings) check() error {
	if (s.purge || s.verify) && s.cacheRoot == "" {
		return errNoCacheToMaintain()
	}
	if s.describe && len(s.places) == 0 && s.scenario == 0 {
		return errNothingToDescribe()
	}
	return nil
}

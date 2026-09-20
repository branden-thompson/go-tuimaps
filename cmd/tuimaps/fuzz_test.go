package main

import (
	"math"
	"strings"
	"testing"
)

// FuzzCommandLine hands the app's one untrusted input - the command line -
// whatever a person or a script can put there. Nothing it is given may make
// it panic, and whatever it answers must hold: a complaint is text a
// terminal can be trusted with, and settings that come back are settings
// the rest of the app can use without checking them again.
func FuzzCommandLine(f *testing.F) {
	for _, seed := range []string{
		"", "--help", "--headless --offline --size 149x38",
		"--describe --scenario 1 --place Home@-84.51,33.82",
		"--purge --no-cache", "--lang pt-BR", "--style /nowhere",
		"--size 69x12 --no-colour --safe-ramps --reduce-motion",
		"--scenario 5", "--place 0,0", "--size 0x0", "-h",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, line string) {
		args := strings.Split(line, " ")
		s, err := parse(args)
		if err != nil {
			if !clean(err.Error()) {
				t.Fatalf("the complaint about %q is not text a terminal can be trusted with: %q", line, err)
			}
			if strings.TrimSpace(err.Error()) == "" {
				t.Fatalf("%q was refused with nothing said about why", line)
			}
			return
		}
		if s.cols < 0 || s.rows < 0 {
			t.Fatalf("%q gave the size %dx%d", line, s.cols, s.rows)
		}
		if (s.cols == 0) != (s.rows == 0) {
			t.Fatalf("%q gave a size with one side and not the other: %dx%d", line, s.cols, s.rows)
		}
		if s.scenario != 0 && !drawable(s.scenario) {
			t.Fatalf("%q gave scenario %d, which this release cannot draw", line, s.scenario)
		}
		if len(s.language) > languageMax || s.language == "" {
			t.Fatalf("%q gave the language %q", line, s.language)
		}
		for _, p := range s.places {
			if math.Abs(p.At.Lon) > 180 || math.Abs(p.At.Lat) > 90 || p.Name == "" {
				t.Fatalf("%q gave the place %+v", line, p)
			}
		}
		if s.noCache && s.cacheRoot != "" {
			t.Fatalf("%q turned the cache off and kept a root for it", line)
		}
		if err := s.check(); err != nil && !clean(err.Error()) {
			t.Fatalf("%q: the complaint is not clean: %q", line, err)
		}
	})
}

// drawable reports whether a scenario number is one this release draws.
func drawable(n int) bool {
	for _, can := range m1Scenarios() {
		if can == n {
			return true
		}
	}
	return false
}

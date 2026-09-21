package main

import (
	"strings"
	"testing"
)

// TestHelpStatesBrailleNeed is plan task 13.10 (D-57): a terminal cannot be
// asked whether its font has braille, so the app says so itself. A person
// whose font has none sees boxes and must be told why before deciding the
// program is broken.
func TestHelpStatesBrailleNeed(t *testing.T) {
	help := helpText()
	if !strings.Contains(help, "braille") {
		t.Error("the help does not mention braille at all")
	}
	if !strings.Contains(help, "font") {
		t.Error("the help mentions braille without saying it is a matter of the font")
	}
}

// TestNetworkOnByDefaultAndSaidInHelp is plan task 13.6 (D-65): the library
// reaches nothing until it is told to, and the app tells it to. That is a
// difference a person must be able to find out without reading the source,
// so the help says it, and says how to turn it off.
func TestNetworkOnByDefaultAndSaidInHelp(t *testing.T) {
	was, err := parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if was.offline {
		t.Fatal("the app is offline before anything asked it to be")
	}
	help := helpText()
	for _, word := range []string{"--offline", defaultSource} {
		if !strings.Contains(help, word) {
			t.Errorf("the help does not name %q", word)
		}
	}
	if !strings.Contains(strings.ToLower(help), "on by default") {
		t.Error("the help does not say that the network is on by default")
	}
}

// TestHelpListsAccessibilitySwitches is plan task 13.13 (PL-AX-5): every
// accessibility switch is in the help, and each has both a flag and a key -
// a person who cannot use one must be able to use the other (NFR-15).
func TestHelpListsAccessibilitySwitches(t *testing.T) {
	help := helpText()
	for _, switched := range accessibility() {
		if !strings.Contains(help, switched.what) {
			t.Errorf("the help does not name the %s switch", switched.what)
		}
		if switched.flag != "" && !strings.Contains(help, switched.flag) {
			t.Errorf("%s has no flag in the help", switched.what)
		}
		if switched.key != "" && !strings.Contains(help, "key "+switched.key) {
			t.Errorf("%s has no key in the help", switched.what)
		}
		if switched.env != "" && !strings.Contains(help, switched.env) {
			t.Errorf("%s is set in the environment and the help does not name it", switched.what)
		}
		if switched.flag == "" && switched.key == "" && switched.env == "" {
			t.Errorf("%s is neither a flag, a key nor a setting in the environment", switched.what)
		}
	}
	for _, want := range []string{"safe ramps", "reduce motion", "no colour", "describe", "NO_COLOR"} {
		found := false
		for _, switched := range accessibility() {
			if switched.what == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is not one of the switches the app lists", want)
		}
	}
}

// TestPrintsOnlyCleanText is plan task 13.11: everything the app puts on a
// terminal is text a terminal can be trusted with. The frame comes from the
// library already cleaned (FR-34); what the app writes of its own is held to
// the same bar here.
func TestPrintsOnlyCleanText(t *testing.T) {
	if !clean(helpText()) {
		t.Error("the help text is not clean")
	}
	for _, line := range strings.Split(helpText(), "\n") {
		if strings.Contains(line, "\x1b") {
			t.Errorf("the help carries an escape sequence: %q", line)
		}
	}
	// And the checker itself is worth nothing if it passes anything.
	for _, bad := range []string{"\x1b[2J", "a\x07b", "a\x00b", "ab", "a‮b", "\xff"} {
		if clean(bad) {
			t.Errorf("%q passed as clean text", bad)
		}
	}
	for _, good := range []string{"", "plain words", "a line\nand another\n", "149x38", "degrees Celsius"} {
		if !clean(good) {
			t.Errorf("%q was called unclean", good)
		}
	}
}

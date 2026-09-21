package main

import (
	"strconv"
	"strings"
	"testing"
)

// TestFlags is the app's whole switchboard: every flag the design names,
// what it means when it is left out, and a clear word when it is written
// wrongly.
func TestFlags(t *testing.T) {
	// What the app is before anything is said to it.
	was, err := parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if was.offline {
		t.Error("the app starts offline; the network is on by default (D-65)")
	}
	if was.headless || was.describe || was.purge || was.verify {
		t.Error("a mode is chosen with no flag to choose it")
	}
	if was.safeRamps || was.reduceMotion || was.noColour {
		t.Error("an accessibility switch is on with nothing asking for it")
	}
	if was.cols != 0 || was.rows != 0 {
		t.Error("a size is set before the terminal has been asked")
	}
	if was.language != "en" {
		t.Errorf("the label language starts as %q, want en", was.language)
	}
	if was.cacheRoot == "" {
		t.Error("the app keeps no tiles on disk by default; FR-21b asks for a cache the app can empty")
	}

	// Each switch, said.
	on, err := parse([]string{"--offline", "--no-cache", "--safe-ramps", "--reduce-motion",
		"--no-colour", "--headless", "--lang", "de", "--size", "149x38", "--style", "/no/such/style.json"})
	if err != nil {
		t.Fatal(err)
	}
	if !on.offline || !on.noCache || !on.safeRamps || !on.reduceMotion || !on.noColour || !on.headless {
		t.Errorf("a flag was not taken: %+v", on)
	}
	if on.language != "de" {
		t.Errorf("the language is %q", on.language)
	}
	if on.cols != 149 || on.rows != 38 {
		t.Errorf("the size is %dx%d, want 149x38", on.cols, on.rows)
	}
	if on.stylePath != "/no/such/style.json" {
		t.Errorf("the style path is %q", on.stylePath)
	}
	if on.cacheRoot != "" {
		t.Error("--no-cache left a cache root")
	}

	// The American spelling is taken too: a flag a person writes the other
	// way round is not a mistake worth stopping for (D-83).
	other, err := parse([]string{"--no-color"})
	if err != nil || !other.noColour {
		t.Errorf("--no-color: %v, %+v", err, other)
	}

	// The two sizes M1 is judged at.
	for _, size := range []struct {
		text       string
		cols, rows int
	}{{"149x38", 149, 38}, {"69x12", 69, 12}, {"69X12", 69, 12}} {
		got, err := parse([]string{"--size", size.text})
		if err != nil {
			t.Fatalf("--size %s: %v", size.text, err)
		}
		if got.cols != size.cols || got.rows != size.rows {
			t.Errorf("--size %s gave %dx%d", size.text, got.cols, got.rows)
		}
	}
}

// TestFlagMistakesAreSaidPlainly: every way of writing a flag wrongly is
// answered with what was wrong and what to write instead (NFR-20's rule,
// applied to the app's own input).
func TestFlagMistakesAreSaidPlainly(t *testing.T) {
	for _, c := range []struct {
		args []string
		says string
	}{
		{[]string{"--size", "big"}, "149x38"},
		{[]string{"--size", "0x38"}, "at least 1"},
		{[]string{"--size", "149x"}, "149x38"},
		{[]string{"--scenario", "9"}, "1 to 7"},
		{[]string{"--scenario", "0"}, "1 to 7"},
		{[]string{"--place", "nowhere"}, "lon,lat"},
		{[]string{"--place", "Home@200,0"}, "longitude"},
		{[]string{"--place", "Home@0,99"}, "latitude"},
		{[]string{"--lang", "a language, not a code"}, "code"},
		{[]string{"--nonsense"}, "nonsense"},
	} {
		_, err := parse(c.args)
		if err == nil {
			t.Errorf("%v was taken without complaint", c.args)
			continue
		}
		if !strings.Contains(err.Error(), c.says) {
			t.Errorf("%v: %q does not say %q", c.args, err, c.says)
		}
		if !clean(err.Error()) {
			t.Errorf("%v: the complaint is not text a terminal can be trusted with: %q", c.args, err)
		}
	}
}

// TestPlacesGivenOnTheCommandLine: a place is a name and a position, and a
// position alone names itself.
func TestPlacesGivenOnTheCommandLine(t *testing.T) {
	got, err := parse([]string{"--place", "Home@-84.51,33.82", "--place", "-84.39,33.75"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.places) != 2 {
		t.Fatalf("%d places, want 2", len(got.places))
	}
	if got.places[0].Name != "Home" || got.places[0].At.Lon != -84.51 || got.places[0].At.Lat != 33.82 {
		t.Errorf("the first place is %+v", got.places[0])
	}
	if got.places[1].Name == "" {
		t.Error("a place given as a position alone has no name at all")
	}
}

// TestScenarioFlag is plan task 13.12: the scenario flag names one of the M1
// scenarios, which is what lets HUM LEAD judge M1a live and gives --describe
// something to describe.
func TestScenarioFlag(t *testing.T) {
	for _, n := range m1Scenarios() {
		got, err := parse([]string{"--scenario", strconv.Itoa(n)})
		if err != nil {
			t.Fatalf("--scenario %d: %v", n, err)
		}
		if got.scenario != n {
			t.Errorf("--scenario %d gave %d", n, got.scenario)
		}
	}
	// Scenario 5 is wind, which this release does not draw (D-44). It is
	// refused by name rather than by "out of range", which would be a lie.
	_, err := parse([]string{"--scenario", "5"})
	if err == nil {
		t.Fatal("--scenario 5 was taken; wind is not in this release")
	}
	if !strings.Contains(err.Error(), "wind") {
		t.Errorf("--scenario 5: %q does not say what scenario 5 is", err)
	}
}

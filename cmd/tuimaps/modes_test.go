package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// ran runs the app as the command line would, and gives back what it wrote
// and what it exited with. Nothing here reaches a network: every case is
// --offline.
func ran(t *testing.T, args ...string) (out, errs string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, &o, &e)
	return o.String(), e.String(), code
}

// TestHeadlessFlag is plan task 13.4: one complete frame to standard
// output, and nothing else. It is how a map is put in a file, a pipe or a
// bug report.
func TestHeadlessFlag(t *testing.T) {
	out, errs, code := ran(t, "--headless", "--offline", "--size", "69x12")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	if errs != "" {
		t.Errorf("headless wrote to standard error: %q", errs)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 12 {
		t.Fatalf("%d rows, want 12", len(lines))
	}
	if strings.TrimSpace(colourless(out)) == "" {
		t.Error("the frame is blank")
	}
	// A frame drawn from the built-in tiles alone is complete: there is
	// nothing left to fetch (D-65).
	if strings.Contains(out, "no tiles") {
		t.Error("headless drew nothing from the tiles built into the app")
	}
	// The other size M1 is judged at.
	out, _, code = ran(t, "--headless", "--offline", "--size", "149x38")
	if code != 0 {
		t.Fatal(code)
	}
	if got := len(strings.Split(strings.TrimRight(out, "\n"), "\n")); got != 38 {
		t.Errorf("%d rows, want 38", got)
	}
}

// TestSizeFlag is the first half of plan task 13.14: --size sets the map's
// size in every mode that has no terminal to ask, so that M1's two sizes
// can be produced exactly.
func TestSizeFlag(t *testing.T) {
	for _, size := range []struct {
		text string
		rows int
	}{{"149x38", 38}, {"69x12", 12}} {
		out, _, code := ran(t, "--headless", "--offline", "--size", size.text)
		if code != 0 {
			t.Fatalf("--size %s: exit %d", size.text, code)
		}
		if got := len(strings.Split(strings.TrimRight(out, "\n"), "\n")); got != size.rows {
			t.Errorf("--size %s drew %d rows", size.text, got)
		}
	}
	// And a size the terminal cannot hold is refused before anything is
	// drawn, rather than drawn wrongly.
	_, errs, code := ran(t, "--headless", "--offline", "--size", "0x0")
	if code == 0 {
		t.Error("a map of no size was drawn")
	}
	if !strings.Contains(errs, "at least 1") {
		t.Errorf("the complaint was %q", errs)
	}
}

// TestDescribeMode is plan task 13.5 (FR-5, D-52): --describe prints what
// the map says in words and draws no map at all. It is the path a person
// who does not read the picture takes.
func TestDescribeMode(t *testing.T) {
	out, errs, code := ran(t, "--describe", "--offline", "--scenario", "1")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errs)
	}
	if strings.Contains(out, "\x1b") {
		t.Error("the description carries an escape sequence")
	}
	if !clean(out) {
		t.Error("the description is not text a terminal can be trusted with")
	}
	for _, r := range out {
		if r >= 0x2800 && r <= 0x28FF {
			t.Fatal("the description is drawn with braille; it is meant to be read aloud")
		}
	}
	for _, want := range []string{"Home", "Downtown", "flood-warning", "inside", "outside", "kilometres"} {
		if !strings.Contains(out, want) {
			t.Errorf("the description does not say %q:\n%s", want, out)
		}
	}
	// The place given on the command line is described too.
	out, _, code = ran(t, "--describe", "--offline", "--scenario", "1", "--place", "Marietta@-84.55,33.95")
	if code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out, "Marietta") {
		t.Errorf("a place named on the command line is not described:\n%s", out)
	}
}

// TestDescribeWithNothingToDescribe: a person who asks for a description
// and has named no place is told what to do, not given an empty answer.
func TestDescribeWithNothingToDescribe(t *testing.T) {
	out, errs, code := ran(t, "--describe", "--offline")
	if code == 0 {
		t.Errorf("an empty description was called a success: %q", out)
	}
	if !strings.Contains(errs, "--place") {
		t.Errorf("the complaint does not say how to name a place: %q", errs)
	}
}

// TestScenarioModesLoadTheirData is plan task 13.12: each M1 scenario the
// release can draw loads its places and its overlays, in both modes, so
// that M1a can be judged live and M1b's data has something to describe.
func TestScenarioModesLoadTheirData(t *testing.T) {
	for _, n := range m1Scenarios() {
		name := "--scenario " + string(rune('0'+n))
		out, errs, code := ran(t, "--describe", "--offline", "--scenario", string(rune('0'+n)))
		if code != 0 {
			t.Errorf("%s: exit %d: %s", name, code, errs)
			continue
		}
		if strings.TrimSpace(out) == "" {
			t.Errorf("%s described nothing", name)
		}
		frame, errs, code := ran(t, "--headless", "--offline", "--scenario", string(rune('0'+n)), "--size", "69x12")
		if code != 0 {
			t.Errorf("%s headless: exit %d: %s", name, code, errs)
			continue
		}
		if got := len(strings.Split(strings.TrimRight(frame, "\n"), "\n")); got != 12 {
			t.Errorf("%s headless drew %d rows", name, got)
		}
	}
}

// TestScenariosAreTheOnesTheKeyWasComputedFor: the scenario files inside
// the app are copies, and a copy that drifts from the one the answer key
// was computed against would make the app say something M1b never checked.
func TestScenariosAreTheOnesTheKeyWasComputedFor(t *testing.T) {
	canonical := filepath.Join("..", "..", "06_docs", "02_features", "go-tuimaps", "02-analysis", "scenarios")
	for _, n := range m1Scenarios() {
		name := "scenario-" + string(rune('0'+n)) + ".json"
		theirs, err := os.ReadFile(filepath.Join(canonical, name))
		if err != nil {
			t.Fatal(err)
		}
		mine, err := scenarioFiles.ReadFile("scenarios/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(theirs, mine) {
			t.Errorf("%s inside the app is not the file the answer key was computed against", name)
		}
	}
}

// TestStyleFlag is plan task 13.15 (FR-20, D-94): the app reads a style
// file and hands its bytes to the library, which never opens one itself. A
// file that is missing or cannot be read is a clear word and no map.
func TestStyleFlag(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "own.json")
	body := `{"layers":[{"id":"sea","type":"fill","source-layer":"water","paint":{"fill-color":"#102080"}}]}`
	if err := os.WriteFile(good, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	mine, _, code := ran(t, "--headless", "--offline", "--size", "69x12", "--style", good)
	if code != 0 {
		t.Fatalf("a style of one's own: exit %d", code)
	}
	theirs, _, _ := ran(t, "--headless", "--offline", "--size", "69x12")
	if mine == theirs {
		t.Error("the style made no difference to the map")
	}

	missing := filepath.Join(dir, "nowhere.json")
	_, errs, code := ran(t, "--headless", "--offline", "--style", missing)
	if code == 0 {
		t.Error("a style file that is not there was passed over in silence")
	}
	if !strings.Contains(errs, "nowhere.json") {
		t.Errorf("the complaint does not name the file: %q", errs)
	}

	broken := filepath.Join(dir, "broken.json")
	if err := os.WriteFile(broken, []byte("layers:"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errs, code = ran(t, "--headless", "--offline", "--style", broken)
	if code == 0 {
		t.Error("a style that cannot be read was drawn with anyway")
	}
	if !clean(errs) {
		t.Errorf("the complaint is not clean text: %q", errs)
	}
}

// TestOfflineFlag is plan task 13.6 (D-65): --offline reaches nothing, and
// says so where it matters - the map still draws, from the tiles built in.
func TestOfflineFlag(t *testing.T) {
	out, _, code := ran(t, "--headless", "--offline", "--size", "69x12")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.TrimSpace(colourless(out)) == "" {
		t.Error("offline drew nothing at all")
	}
	// The settings an offline run makes: no source, and no cache to keep
	// what was never fetched.
	s, err := parse([]string{"--offline"})
	if err != nil {
		t.Fatal(err)
	}
	if sourceFor(s) != "" {
		t.Errorf("an offline run names the source %q", sourceFor(s))
	}
	on, err := parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if sourceFor(on) != defaultSource {
		t.Errorf("a run that was told nothing names %q, want the default source", sourceFor(on))
	}
}

// TestNoCacheAndPurgeFlags is plan task 13.7 (FR-21b): the disk cache is
// the app's to offer and the person's to empty. Cache file names are a
// record of where someone has looked, so turning it off must work, and
// emptying it must say what it did.
func TestNoCacheAndPurgeFlags(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir) // where the app keeps tiles on Linux
	t.Setenv("HOME", dir)           // and on macOS
	if got, _ := parse([]string{"--no-cache"}); got.cacheRoot != "" {
		t.Errorf("--no-cache left the cache at %q", got.cacheRoot)
	}
	out, errs, code := ran(t, "--purge", "--offline")
	if code != 0 {
		t.Fatalf("--purge: exit %d: %s", code, errs)
	}
	if !strings.Contains(out, "cache") {
		t.Errorf("--purge said %q", out)
	}
	out, errs, code = ran(t, "--verify", "--offline")
	if code != 0 {
		t.Fatalf("--verify: exit %d: %s", code, errs)
	}
	if !strings.Contains(out, "cache") {
		t.Errorf("--verify said %q", out)
	}
	// With the cache turned off, there is nothing to empty, and that is
	// said rather than pretended.
	_, errs, code = ran(t, "--purge", "--no-cache", "--offline")
	if code == 0 {
		t.Error("emptying a cache that is turned off was called a success")
	}
	if !strings.Contains(errs, "--no-cache") {
		t.Errorf("the complaint was %q", errs)
	}
}

// TestSafeRampsAndReduceMotionFlags is plan task 13.8's flag half: the two
// accessibility switches reach the library. Their keys are the other half,
// and live with the key map.
func TestSafeRampsAndReduceMotionFlags(t *testing.T) {
	plain, _, code := ran(t, "--headless", "--offline", "--size", "69x12", "--no-colour")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(plain, "\x1b") {
		t.Error("--no-colour drew colour sequences")
	}
	colourful, _, _ := ran(t, "--headless", "--offline", "--size", "69x12")
	if !strings.Contains(colourful, "\x1b") {
		t.Error("a run with colour drew none")
	}
	// NO_COLOR in the environment does the same, with no flag at all
	// (NFR-15).
	t.Setenv("NO_COLOR", "1")
	byEnvironment, _, _ := ran(t, "--headless", "--offline", "--size", "69x12")
	if strings.Contains(byEnvironment, "\x1b") {
		t.Error("NO_COLOR was set and the map drew colour anyway")
	}
	t.Setenv("NO_COLOR", "")
	for _, flag := range []string{"--safe-ramps", "--reduce-motion"} {
		if _, _, code := ran(t, "--headless", "--offline", "--size", "69x12", flag); code != 0 {
			t.Errorf("%s: exit %d", flag, code)
		}
	}
}

// TestHelpIsPrintedAndNothingElseHappens: --help writes the help and exits
// without opening, fetching or drawing anything.
func TestHelpIsPrintedAndNothingElseHappens(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		out, errs, code := ran(t, arg)
		if code != 0 {
			t.Errorf("%s: exit %d", arg, code)
		}
		if errs != "" {
			t.Errorf("%s wrote to standard error: %q", arg, errs)
		}
		if !strings.Contains(out, "braille") || !strings.Contains(out, "--offline") {
			t.Errorf("%s printed something that is not the help", arg)
		}
	}
	_, errs, code := ran(t, "--nonsense")
	if code == 0 {
		t.Error("an unknown flag was ignored")
	}
	if !strings.Contains(errs, "--help") {
		t.Errorf("the complaint does not point at the help: %q", errs)
	}
}

// TestEverythingPrintedIsCleanText is plan task 13.11 over every mode: what
// the app writes of its own is held to the same bar as what the library
// hands out (FR-34).
func TestEverythingPrintedIsCleanText(t *testing.T) {
	for _, args := range [][]string{
		{"--help"},
		{"--describe", "--offline", "--scenario", "1"},
		{"--describe", "--offline", "--scenario", "4"},
		{"--nonsense"},
		{"--size", "big"},
		{"--headless", "--offline", "--size", "69x12", "--no-colour"},
	} {
		out, errs, _ := ran(t, args...)
		for what, text := range map[string]string{"what it printed": out, "what it complained": errs} {
			if !utf8.ValidString(text) {
				t.Errorf("%v: %s is not valid text", args, what)
			}
			if !clean(colourless(text)) {
				t.Errorf("%v: %s is not clean: %q", args, what, text)
			}
		}
	}
}

// colourless is the text of something drawn, with the escape sequences
// taken out - the colours the library emits, and the few the app sends to
// the terminal itself. A sequence ends at its last byte, which is anything
// from @ to ~; taking it to end at "m" would eat the row behind it.
func colourless(text string) string {
	var b strings.Builder
	for len(text) > 0 {
		before, after, found := strings.Cut(text, "\x1b[")
		b.WriteString(before)
		if !found {
			break
		}
		end := strings.IndexFunc(after, func(r rune) bool { return r >= '@' && r <= '~' })
		if end < 0 {
			break
		}
		text = after[end+1:]
	}
	return b.String()
}

// noon is a fixed instant, so that a frame drawn in a test is the same
// frame every time.
var noon = time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

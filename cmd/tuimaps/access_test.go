package main

// access_test.go — D-135: the demo app's accessibility findings, each held
// by a test. The keys panel shows every key, escape sequences are read
// whole, every key row names play, toggles say what they did, the status
// row composes, failures are said, colour comes back as it started, answers
// are worded from the legend, a loop opens stopped, and the words are plain.

import (
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// statusOf is the status row as drawn.
func statusOf(t *testing.T, a *app, out *shown) string {
	t.Helper()
	out.Reset()
	if err := a.drawn(noon); err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(out.String(), "\r\n")
	return colourless(rows[len(rows)-1])
}

// TestTheKeysPanelShowsEveryKey (finding 1): on a 24-row terminal the ?
// panel is the key table, every key on it and nothing cut; a panel that does
// not fit names the command that prints it whole.
func TestTheKeysPanelShowsEveryKey(t *testing.T) {
	a, out := upFor(t)
	a.act(keyHelp)
	out.Reset()
	if err := a.drawn(noon); err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(colourless(out.String()), "\r\n")
	panel := strings.Join(rows[:mapRows(a.rows)], "\n")
	for _, want := range []string{"quit", "zoom in", "zoom out", "pan", "names", "water", "whole world", "markers",
		"Tab", "Shift+Tab", "play", "[ ]", "right now", "safe ramps", "reduce motion", "colour", "describe", "these keys"} {
		if !strings.Contains(panel, want) {
			t.Errorf("the keys panel does not show %q:\n%s", want, panel)
		}
	}
	if strings.Contains(panel, "more lines") {
		t.Errorf("the keys panel is cut on a 24-row terminal:\n%s", panel)
	}
	for _, line := range strings.Split(keysText(), "\n") {
		if len([]rune(line)) > 69 {
			t.Errorf("a line of the key table is %d wide, wider than a 69-column terminal: %q", len([]rune(line)), line)
		}
	}
	for _, row := range keyTable {
		if len(row[0]) >= 11 || len(row[1]) >= 23 || len(row[2]) >= 10 {
			t.Errorf("the key table's row %q leaves less than two spaces between its columns", row)
		}
	}
	if !strings.Contains(helpText(), keysText()) {
		t.Error("--help does not carry the key table")
	}
	long := strings.Split(strings.Repeat("words\n", 30), "\n")
	if got := fitted(long, 10, 69, "tuimaps --help"); !strings.Contains(got[9], "tuimaps --help") {
		t.Errorf("the overflow line is %q; want it to name tuimaps --help", got[9])
	}
}

// TestEscapeSequencesAreReadWhole (finding 2): a CSI or SS3 sequence is read
// to its end, arrows with modifiers are arrows, a sequence the app has no
// use for is passed over whole, Alt and a key is no key, and Esc is Esc only
// when nothing follows it in the read. A sequence split across reads is
// held until it is whole.
func TestEscapeSequencesAreReadWhole(t *testing.T) {
	for _, c := range []struct {
		name string
		sent string
		want []string
	}{
		{"control and right", "\x1b[1;5C", []string{keyRight}},
		{"shift and up", "\x1b[1;2A", []string{keyUp}},
		{"SS3 arrows", "\x1bOA\x1bOD", []string{keyUp, keyLeft}},
		{"F1, SS3", "\x1bOP", nil},
		{"F5, CSI", "\x1b[15~", nil},
		{"a modified F-key", "\x1b[15;5~", nil},
		{"Alt and q", "\x1bq", nil},
		{"Alt and a letter, then a key", "\x1bxn", []string{keyNames}},
		{"shift and tab", "\x1b[Z", []string{keyFocusBack}},
		{"home, three ways", "\x1b[H\x1b[1~\x1bOH", []string{keyNow, keyNow, keyNow}},
		{"zero", "0", []string{keyNow}},
		{"a sequence then a key", "\x1b[1;5Cq", []string{keyRight, keyQuit}},
		{"escape alone", "\x1b", []string{keyEscape}},
		{"escape twice", "\x1b\x1b", []string{keyEscape, keyEscape}},
	} {
		got := decode([]byte(c.sent))
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	var r keyReader
	for _, c := range []struct {
		sent string
		want []string
	}{
		{"\x1b[1;", nil}, {"5C", []string{keyRight}},
		{"\x1bO", nil}, {"A", []string{keyUp}},
		{"\x1b[", nil}, {"Z", []string{keyFocusBack}},
		{"n", []string{keyNames}},
	} {
		if got := r.read([]byte(c.sent)); strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("split read %q: %v, want %v", c.sent, got, c.want)
		}
	}
	// A sequence that never ends is dropped rather than held for ever.
	var endless keyReader
	endless.read([]byte("\x1b[" + strings.Repeat("1;", 40)))
	if got := endless.read([]byte("q")); strings.Join(got, ",") != keyQuit {
		t.Errorf("after an endless sequence, q read as %v", got)
	}
}

// TestEveryKeyRowNamesPlay (finding 3): every length of the key row names
// play, the longer ones the steps; the row for 69 columns fits them.
func TestEveryKeyRowNamesPlay(t *testing.T) {
	for _, cols := range []int{200, 100, 69, 30} {
		row := keyRow(cols)
		if !strings.Contains(row, "p play") || !strings.Contains(row, "q quit") || !strings.Contains(row, "? keys") {
			t.Errorf("the key row at %d columns is %q; want quit, play and keys", cols, row)
		}
		if cols >= 69 && !strings.Contains(row, "[ ] step") {
			t.Errorf("the key row at %d columns is %q; want the steps", cols, row)
		}
		if len(row) > cols {
			t.Errorf("the key row at %d columns is %d long", cols, len(row))
		}
	}
}

// TestShiftTabFocusesBackwards (finding 3): Shift+Tab walks the places the
// other way, through no focus, as Tab does forwards.
func TestShiftTabFocusesBackwards(t *testing.T) {
	a, _ := upFor(t, tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: 1, Lat: 1}}, tuimaps.Place{Name: "Work", At: tuimaps.LonLat{Lon: 2, Lat: 2}})
	var went []int
	for range 3 {
		a.act(keyFocusBack)
		went = append(went, a.focus)
	}
	if went[0] != 1 || went[1] != 0 || went[2] != -1 {
		t.Errorf("Shift+Tab went %v; want 1, 0, -1", went)
	}
}

// TestRightNowKeyReturnsTheLoop (finding 3): 0 returns a loop to its newest
// observed frame and stops it; with no loop it says how to get one.
func TestRightNowKeyReturnsTheLoop(t *testing.T) {
	m, err := build(settings{offline: true, noColour: true, cols: 100, rows: 30, loopDir: outbreak, language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	a := newApp(m, settings{offline: true}, &shown{}, 100, 30)
	a.act(keyStepBack)
	a.act(keyStepBack)
	a.act(keyPlay)
	a.act(keyNow)
	st := m.Loop()
	if st.Playing || st.Index != st.Count-1 {
		t.Errorf("after 0 the loop is at frame %d of %d, playing %v; want the newest, stopped", st.Index+1, st.Count, st.Playing)
	}
	if !strings.Contains(a.said, "right now") {
		t.Errorf("0 said %q", a.said)
	}
	b, _ := upFor(t)
	b.act(keyNow)
	if !strings.Contains(b.said, "--loop") {
		t.Errorf("0 with no loop said %q; want how to load one", b.said)
	}
}

// TestTogglesSayWhatTheyDid (finding 4): each switch says in the status row
// what it turned on or off.
func TestTogglesSayWhatTheyDid(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	for _, c := range []struct{ key, first, second string }{
		{keyNames, "names off", "names on"},
		{keyWater, "water off", "water on"},
		{keyMarkers, "markers off", "markers on"},
		{keySafeRamps, "safe ramps on", "safe ramps off"},
		{keyReduce, "reduce motion on", "reduce motion off"},
		{keyColour, "colour off", "colour on"},
	} {
		a, out := upFor(t)
		a.act(c.key)
		if got := statusOf(t, a, out); !strings.Contains(got, c.first) {
			t.Errorf("%s pressed once: the status row is %q; want %q", c.key, got, c.first)
		}
		a.act(c.key)
		if got := statusOf(t, a, out); !strings.Contains(got, c.second) {
			t.Errorf("%s pressed twice: the status row is %q; want %q", c.key, got, c.second)
		}
	}
}

// TestTheStatusRowComposes (finding 5): a focused place and what was said
// are added to the loop's state and the map's, never in place of them, and
// the row stays within the terminal.
func TestTheStatusRowComposes(t *testing.T) {
	a, _ := upFor(t, tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.5, Lat: 33.8}})
	a.offline = true
	a.act(keyFocus)
	a.said = "the pan was refused"
	if err := a.m.Zoom(9); err != nil {
		t.Fatal(err)
	}
	got := a.statusRow(tuimaps.Frame{Status: tuimaps.Sharpening})
	for _, want := range []string{"the pan was refused", "focused: Home", "--offline"} {
		if !strings.Contains(got, want) {
			t.Errorf("the status row is %q; want it to carry %q", got, want)
		}
	}
	if line := cut(got, 69); len([]rune(line)) > 69 {
		t.Errorf("the status row is %d wide", len([]rune(line)))
	}
	m, err := build(settings{offline: true, noColour: true, cols: 100, rows: 30, loopDir: outbreak, language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	b := newApp(m, settings{offline: true}, &shown{}, 100, 30)
	b.said = "the pan was refused"
	if got := b.statusRow(tuimaps.Frame{Status: tuimaps.Complete}); !strings.Contains(got, "frame 12 of 12") || !strings.Contains(got, "refused") {
		t.Errorf("the status row with a loop is %q; want what was said and the loop's state", got)
	}
}

// TestFailuresAreSaid (finding 6): online, a tile that failed is said with a
// way to resolve it until the map is whole; offline, the notice is for a
// zoom beyond the built-in map or no tiles at all, and any other incomplete
// frame is loading.
func TestFailuresAreSaid(t *testing.T) {
	a, _ := upFor(t)
	if got := a.statusRow(tuimaps.Frame{Status: tuimaps.Sharpening}); !strings.Contains(got, "loading") || strings.Contains(got, "sharpening") {
		t.Errorf("an incomplete frame online says %q; want loading", got)
	}
	a.heard([]tuimaps.Warning{{Kind: tuimaps.TileFailed, Count: 3}})
	got := a.statusRow(tuimaps.Frame{Status: tuimaps.Sharpening})
	if !strings.Contains(got, "could not be fetched") || !strings.Contains(got, "connection") {
		t.Errorf("a failed tile online says %q; want it said with a way to resolve it", got)
	}
	a.statusRow(tuimaps.Frame{Status: tuimaps.Complete})
	if got := a.statusRow(tuimaps.Frame{Status: tuimaps.Complete}); strings.Contains(got, "could not be fetched") {
		t.Errorf("a whole map still says %q", got)
	}
	a.offline = true
	if err := a.m.Zoom(2); err != nil {
		t.Fatal(err)
	}
	if got := a.statusRow(tuimaps.Frame{Status: tuimaps.Sharpening}); strings.Contains(got, "--offline") || !strings.Contains(got, "loading") {
		t.Errorf("offline within the built-in zoom, an incomplete frame says %q; want loading", got)
	}
	if got := a.statusRow(tuimaps.Frame{Status: tuimaps.NoTiles}); !strings.Contains(got, "no map at this zoom") {
		t.Errorf("offline with no tiles says %q", got)
	}
	if err := a.m.Zoom(6); err != nil {
		t.Fatal(err)
	}
	if got := a.statusRow(tuimaps.Frame{Status: tuimaps.Sharpening}); !strings.Contains(got, "no map at this zoom") {
		t.Errorf("offline beyond the built-in zoom says %q", got)
	}
}

// TestATileFailureReachesTheStatusRow (finding 6): a source that cannot be
// reached fails its tiles, and the drawn status row says so.
func TestATileFailureReachesTheStatusRow(t *testing.T) {
	a, out := upFor(t)
	if err := a.m.Source("http://127.0.0.1:9/"); err != nil {
		t.Fatal(err)
	}
	if err := a.m.Zoom(6); err != nil {
		t.Fatal(err)
	}
	if _, err := a.m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := statusOf(t, a, out); !strings.Contains(got, "could not be fetched") {
		t.Errorf("with every tile failing the status row is %q", got)
	}
}

// TestColourComesBackAsItStarted (finding 7): C turned off and on again puts
// back the depth the app started with, which under NO_COLOR is none.
func TestColourComesBackAsItStarted(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	a, _ := upFor(t)
	if !a.noColour {
		t.Error("under NO_COLOR the colour switch starts on colour")
	}
	a.act(keyColour)
	if !strings.Contains(a.said, "NO_COLOR") {
		t.Errorf("under NO_COLOR, colour on said %q; want why it stays off", a.said)
	}
	a.act(keyColour)
	if got := a.depth(); got != tuimaps.NoColour {
		t.Errorf("under NO_COLOR, C twice draws at depth %v; want none", got)
	}
	t.Setenv("NO_COLOR", "")
	b, _ := upFor(t)
	b.act(keyColour)
	if got := b.depth(); got != tuimaps.NoColour {
		t.Errorf("C once draws at depth %v; want none", got)
	}
	b.act(keyColour)
	if got := b.depth(); got != tuimaps.Truecolor {
		t.Errorf("C twice draws at depth %v; want the library's own choice", got)
	}
}

// legendForTests is a radar, a temperature and a wind legend, as Legend
// hands them out.
var legendForTests = []tuimaps.LegendEntry{
	{ID: "radar", Unit: "dBZ", Preset: "radar", Classes: []tuimaps.Class{{Label: "under 5"}, {Label: "5 to 10"}, {Label: "10 to 20"}, {Label: "20 and above"}}},
	{ID: "temperature", Unit: "°C", Preset: "temperature", Classes: []tuimaps.Class{{Label: "under 10"}, {Label: "10 to 20"}, {Label: "20 and above"}}},
	{ID: "wind", Unit: "mph", Preset: "wind", Classes: []tuimaps.Class{{Label: "under 10"}, {Label: "10 to 20"}, {Label: "20 and above"}}},
	{ID: "heat", Unit: "C", Preset: "temperature", Classes: []tuimaps.Class{{Label: "under 15"}, {Label: "15 to 20"}}},
}

// TestAnswersAreWordedFromTheLegend (finding 8): an image's class and a
// field's band are said as the legend's range, wind says where it blows
// from and no "change across this view", and radar with no reading says so
// and names what is heavier.
func TestAnswersAreWordedFromTheLegend(t *testing.T) {
	for _, c := range []struct {
		a          tuimaps.Answer
		want, gone string
	}{
		{tuimaps.Answer{Overlay: "radar", Form: "image", Class: 2, Heavier: "east", HeavierAt: 293, Unit: "kilometres"},
			"10 to 20 dBZ, the nearest rain heavier than here is 293 kilometres to the east", "class"},
		{tuimaps.Answer{Overlay: "radar", Form: "image", NoData: true, Heavier: "east", HeavierAt: 293, Unit: "kilometres"},
			"no reading here, the nearest rain is 293 kilometres to the east", "no data"},
		{tuimaps.Answer{Overlay: "radar", Form: "image", Class: 3},
			"20 dBZ and above, with no heavier rain anywhere on it", "class"},
		{tuimaps.Answer{Overlay: "temperature", Form: "field", Value: 15, ValueUnit: "degrees Celsius", Band: 1, Rises: "south-east"},
			"15 degrees Celsius, in the legend's 10 to 20 °C, rising to the south-east", "band"},
		{tuimaps.Answer{Overlay: "temperature", Form: "field", Value: 4, ValueUnit: "degrees Celsius", Band: 0},
			"4 degrees Celsius, in the legend's under 10 °C, with no change across this view", "band"},
		{tuimaps.Answer{Overlay: "wind", Form: "field", Value: 15, ValueUnit: "mph", Band: 1, From: "south-west"},
			"from the south-west at 15 mph, in the legend's 10 to 20 mph", "change"},
		{tuimaps.Answer{Overlay: "heat", Form: "field", Value: 15, ValueUnit: "degrees Celsius", Band: 1, Rises: "south"},
			"15 degrees Celsius, in the legend's 15 to 20 degrees Celsius, rising to the south", " C,"},
		{tuimaps.Answer{Overlay: "wind", Form: "field", NoData: true},
			"no reading here", "no data"},
		{tuimaps.Answer{Overlay: "other", Form: "field", Value: 3, Band: 7},
			"3, band 7, with no change across this view", "legend"},
	} {
		got := sentence(c.a, legendForTests)
		if got != c.want || strings.Contains(got, c.gone) {
			t.Errorf("said %q\n want %q", got, c.want)
		}
	}
}

// TestDescribeSaysNoClassNumbers (finding 8): the description of the
// scenarios says ranges, never a bare class or band number.
func TestDescribeSaysNoClassNumbers(t *testing.T) {
	for _, n := range []string{"3", "4"} {
		out, errs, code := ran(t, "--describe", "--offline", "--scenario", n, "--size", "100x30")
		if code != exitFine {
			t.Fatalf("scenario %s: exit %d: %s", n, code, errs)
		}
		if strings.Contains(out, "band ") || strings.Contains(out, "class ") || strings.Contains(out, "no data") {
			t.Errorf("scenario %s says a number for a range:\n%s", n, out)
		}
	}
}

// TestALoopOpensStopped (finding 9): --loop opens at "right now" with
// playback on and nothing moving until p is pressed; the help says
// --reduce-motion also stops loops.
func TestALoopOpensStopped(t *testing.T) {
	m, err := build(settings{offline: true, noColour: true, cols: 100, rows: 30, loopDir: outbreak, language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	st := m.Loop()
	if st.Playback != tuimaps.PlaybackOn || st.Playing || st.Index != st.Count-1 {
		t.Errorf("a loop opens with playback %v, playing %v, at frame %d of %d; want on, stopped, at the newest", st.Playback, st.Playing, st.Index+1, st.Count)
	}
	for _, line := range strings.Split(helpText(), "\n") {
		if strings.HasPrefix(line, "  --reduce-motion") && !strings.Contains(line, "loop") {
			t.Errorf("the help's --reduce-motion line is %q; want it to say loops stop", line)
		}
	}
}

// TestTheWordsArePlain (finding 10): focus is "focused: Home"; an alert
// whose label repeats its overlay is named once and its severity is the
// same word in both lists; an alert says its times with their zone.
func TestTheWordsArePlain(t *testing.T) {
	a, out := upFor(t, tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.5, Lat: 33.8}})
	a.act(keyFocus)
	if got := statusOf(t, a, out); !strings.Contains(got, "focused: Home") || strings.Contains(got, ">>") {
		t.Errorf("the status row with a place focused is %q; want focused: Home", got)
	}
	issued := time.Date(2026, 5, 1, 14, 5, 0, 0, time.UTC)
	until := time.Date(2026, 5, 1, 16, 0, 0, 0, time.UTC)
	r := tuimaps.Report{
		Places: []tuimaps.PlaceReport{{Place: "Home", Alerts: []tuimaps.PlaceAlert{{Overlay: "flood-warning", Label: "flood-warning",
			Severity: tuimaps.SeveritySevere, Where: tuimaps.Inside, Distance: 7.5, Unit: "kilometres", Compass: "east", Valid: issued}}}},
		Alerts: []tuimaps.AlertShown{{Overlay: "flood-warning", Label: "flood-warning", Severity: tuimaps.SeveritySevere, Valid: issued, Expires: until}},
	}
	said := strings.Join(reported(r, nil), "\n")
	for _, want := range []string{
		"  flood-warning, severe: inside this area, the nearest edge is 7.5 kilometres to the east; valid from 14:05 UTC",
		"  flood-warning, severe, valid from 14:05 UTC until 16:00 UTC",
	} {
		if !strings.Contains(said, want) {
			t.Errorf("the report does not say %q:\n%s", want, said)
		}
	}
	if strings.Contains(said, "flood-warning: flood-warning") || strings.Contains(said, "SEVERE") {
		t.Errorf("the report names an alert twice or shouts its severity:\n%s", said)
	}
	r.Places[0].Alerts[0].Label = "Flood Warning"
	if got := reported(r, nil)[1]; !strings.HasPrefix(got, "  flood-warning: Flood Warning, severe: inside") {
		t.Errorf("an alert with a label of its own reads %q", got)
	}
}

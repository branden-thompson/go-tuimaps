package tuimaps_test

import (
	"context"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

var colours = regexp.MustCompile("\x1b\\[[0-9;]*m")

// world is a settled map of the world at a size, drawn from the embedded
// tiles and reaching nothing.
func world(t *testing.T, cols, rows int) *tuimaps.Map {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(cols, rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m
}

// drawn is one frame of a map as two strings of its own: the frame as it is
// emitted, and its text with the colour sequences taken out. They are copies
// because a frame is only valid until the next Render (contract, section 5).
func drawn(t *testing.T, m *tuimaps.Map, cols, rows int) (raw, text string) {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	raw = strings.Join(f.Lines, "\n")
	return raw, colours.ReplaceAllString(raw, "")
}

// TestLookSettingsChangeTheNextFrame is plan task 12.17 (contract, the Look
// group): each setting takes effect at the next Render, and each changes the
// frame in the way the contract says it does.
func TestLookSettingsChangeTheNextFrame(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	was, _ := drawn(t, m, cols, rows)

	// The palette: a token given a colour of the host's own.
	unknown, err := m.SetPalette(map[string]tuimaps.RGB{"water.fill": {R: 10, G: 20, B: 90}, "not.a.token": {}})
	if err != nil {
		t.Fatal(err)
	}
	if len(unknown) != 1 || unknown[0] != "not.a.token" {
		t.Errorf("the names that are no token came back as %v", unknown)
	}
	now, _ := drawn(t, m, cols, rows)
	if now == was {
		t.Error("the frame is unchanged after a palette was set")
	}

	// The colour depth: the host's hint decides what sequences are emitted.
	m.ColourDepth(tuimaps.NoColour)
	plainRaw, plainText := drawn(t, m, cols, rows)
	if colours.MatchString(plainRaw) {
		t.Error("colour sequences with the depth set to none")
	}
	m.ColourDepth(tuimaps.Colours256)
	at256, _ := drawn(t, m, cols, rows)
	if !strings.Contains(at256, "\x1b[") {
		t.Error("no colour sequences at 256 colours")
	}

	// The layers: the water switched off takes the shore with it.
	m.ColourDepth(tuimaps.NoColour)
	m.Layers(tuimaps.WaterLayer, false)
	_, withoutWater := drawn(t, m, cols, rows)
	if withoutWater == plainText {
		t.Error("switching the water layer off left the frame as it was")
	}
	m.Layers(tuimaps.WaterLayer, true)
	_, backAgain := drawn(t, m, cols, rows)
	if backAgain != plainText {
		t.Error("switching the water layer back on did not give the frame back")
	}

	// Safe ramps and reduce motion are settings the frame may not show on a
	// map with no overlay and no marker, but they must be callable and must
	// not fail.
	m.SafeRamps(true)
	m.ReduceMotion(true)
	if _, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, time.Time{}); err != nil {
		t.Fatalf("with safe ramps and reduce motion on: %v", err)
	}
}

// TestGroundPaintedOrDeclared is the ground half of 12.17 (D-64): by default
// the library paints every cell's background; a host that declares its own
// ground is not painted over.
func TestGroundPaintedOrDeclared(t *testing.T) {
	const cols, rows = 80, 24
	m := world(t, cols, rows)
	painted, _ := drawn(t, m, cols, rows)
	if !strings.Contains(painted, "\x1b[48;2;") {
		t.Error("the library paints the ground by default, and no background colour was set")
	}
	if err := m.Ground(tuimaps.RGB{R: 20, G: 20, B: 24}); err != nil {
		t.Fatal(err)
	}
	declared, _ := drawn(t, m, cols, rows)
	if declared == painted {
		t.Error("declaring the ground changed nothing")
	}
}

// TestPaletteSwapKeepsTiles is plan task 08.22 (FR-15), moved here because it
// is the public call that swaps a palette: the change counter moves, the
// frame changes, and not one tile is read again to do it.
func TestPaletteSwapKeepsTiles(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	drawn(t, m, cols, rows)
	if m.Pending() != 0 {
		t.Fatalf("%d jobs pending before the palette is swapped", m.Pending())
	}
	was := m.Changed()
	if _, err := m.SetPalette(map[string]tuimaps.RGB{"water.fill": {R: 1, G: 2, B: 3}}); err != nil {
		t.Fatal(err)
	}
	if m.Changed() == was {
		t.Error("the change counter did not move when the palette was set")
	}
	if m.Pending() != 0 {
		t.Errorf("%d jobs pending after a palette swap; nothing is decoded again for a colour", m.Pending())
	}
	res, err := m.Settle(context.Background())
	if err != nil || res.Ran != 0 {
		t.Errorf("settle after a palette swap ran %d jobs: %+v, %v", res.Ran, res, err)
	}
}

// TestChangedCounterMovesOnlyWhenTheFrameWould is the counter half of 12.20:
// it moves when a redraw would differ and stands still when it would not.
func TestChangedCounterMovesOnlyWhenTheFrameWould(t *testing.T) {
	const cols, rows = 80, 24
	m := world(t, cols, rows)
	drawn(t, m, cols, rows)
	was := m.Changed()
	drawn(t, m, cols, rows)
	if m.Changed() != was {
		t.Error("the counter moved when nothing changed")
	}
	m.SafeRamps(true)
	if m.Changed() == was {
		t.Error("the counter did not move when the look changed")
	}
	was = m.Changed()
	m.SafeRamps(true) // the same setting again
	if m.Changed() != was {
		t.Error("the counter moved when a setting was set to what it already was")
	}
}

// TestLookOnAClosedMap: every look setting answers the closed kind once the
// map is closed, and none of them panics.
func TestLookOnAClosedMap(t *testing.T) {
	m := world(t, 40, 12)
	m.Close()
	if _, err := m.SetPalette(nil); !isKind(err, fault.Closed) {
		t.Errorf("SetPalette on a closed map: %v", err)
	}
	if err := m.Ground(tuimaps.RGB{}); !isKind(err, fault.Closed) {
		t.Errorf("Ground on a closed map: %v", err)
	}
	m.SafeRamps(true)
	m.ReduceMotion(true)
	m.ColourDepth(tuimaps.NoColour)
	m.Layers(tuimaps.WaterLayer, false)
	if err := m.LabelLanguage("fr"); !isKind(err, fault.Closed) {
		t.Errorf("LabelLanguage on a closed map: %v", err)
	}
}

// TestLabelLanguageChecked: a language is a short code, and anything else is
// refused with a kind from the closed list (D-82).
func TestLabelLanguageChecked(t *testing.T) {
	m := world(t, 40, 12)
	drawn(t, m, 40, 12)
	if m.Pending() != 0 {
		t.Fatalf("%d jobs pending before the language changes", m.Pending())
	}
	if err := m.LabelLanguage("de"); err != nil {
		t.Errorf("a language code: %v", err)
	}
	// A language is part of a tile's cache key, so the tiles on hand are
	// another language's and the view wants them again (D-82).
	drawn(t, m, 40, 12)
	if m.Pending() == 0 {
		t.Error("nothing is wanted after the label language changed; a language is part of a tile's key")
	}
	for _, bad := range []string{"", "english please", "\x1b[0m", "toolongacode"} {
		if err := m.LabelLanguage(bad); !isKind(err, fault.InvalidID) {
			t.Errorf("LabelLanguage(%q): %v; want the invalid-id kind", bad, err)
		}
	}
}

// TestNoColourFromTheEnvironment is NFR-15: with no hint from the host, a
// non-empty NO_COLOR chooses no colour at all; a host that hints is not
// overruled by it.
func TestNoColourFromTheEnvironment(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := world(t, 40, 12)
	raw, _ := drawn(t, m, 40, 12)
	if colours.MatchString(raw) {
		t.Error("colour sequences with NO_COLOR set and no hint from the host")
	}
	m.ColourDepth(tuimaps.Truecolor)
	hinted, _ := drawn(t, m, 40, 12)
	if !colours.MatchString(hinted) {
		t.Error("a host that asked for truecolor was overruled by the environment")
	}
}

// frameAt is a frame drawn at a moment of the host's animation clock, as its
// own string.
func frameAt(t *testing.T, m *tuimaps.Map, cols, rows, ms int) string {
	t.Helper()
	f, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, start.Add(time.Duration(ms)*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\n")
}

// start is the moment the animation tests count from.
var start = time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)

// notANumber is a latitude that is no number at all.
func notANumber() float64 { return math.NaN() }

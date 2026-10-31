package tuimaps_test

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// hostType is the globe grid in a host's own type of 7 classes, MHz.
func hostType(colours []tuimaps.RGB) tuimaps.Overlay {
	g := globeGrid(1)
	g.Type = tuimaps.Type{Unit: "MHz", Breaks: hostBreaks, Colours: colours}
	return tuimaps.Overlay{ID: "host", Valid: noon, Keeps: 3600e9, Grid: &g}
}

// hostFrame is the whole world with one overlay set, at a depth, its frame
// and the map.
func hostFrame(t *testing.T, o tuimaps.Overlay, depth tuimaps.Depth, ground tuimaps.RGB) (string, *tuimaps.Map) {
	t.Helper()
	m := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, m.FitWorld())
	m.ColourDepth(depth)
	must(t, m.Ground(ground))
	mustSet(t, m, o)
	settle(t, m)
	f, err := m.Render(specimenSize, noon)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\n"), m
}

// hostLegend is the host overlay's legend entry.
func hostLegend(t *testing.T, m *tuimaps.Map) tuimaps.LegendEntry {
	t.Helper()
	for _, e := range m.Legend() {
		if e.ID == "host" {
			return e
		}
	}
	t.Fatal("the host type has no legend")
	return tuimaps.LegendEntry{}
}

// TestAHostTypeIsFilledFromLowMiddleHigh is L-2.2 (FR-15): a host's own type
// is filled across its classes from the low, middle and high colours - the
// first class low, the last high, the middle one middle - and with colours
// of the host's own, one a class, in those instead.
func TestAHostTypeIsFilledFromLowMiddleHigh(t *testing.T) {
	raw, m := hostFrame(t, hostType(nil), tuimaps.Truecolor, tuimaps.RGB{R: 16, G: 16, B: 16})
	classes := hostLegend(t, m).Classes
	if len(classes) != 7 {
		t.Fatalf("the legend keys %d classes; the type has 7", len(classes))
	}
	low, middle, high := tuimaps.RGB{R: 38, G: 1, B: 95}, tuimaps.RGB{R: 183, G: 129, B: 52}, tuimaps.RGB{R: 217, G: 253, B: 213}
	if classes[0].Colour != low || classes[3].Colour != middle || classes[6].Colour != high {
		t.Errorf("the classes run %v, %v, %v; want low %v, middle %v and high %v", classes[0].Colour, classes[3].Colour, classes[6].Colour, low, middle, high)
	}
	drawn := 0
	for _, c := range classes {
		if groundsOf([]tuimaps.Class{c}, raw) > 0 {
			drawn++
		}
	}
	if drawn < 5 {
		t.Errorf("%d of the 7 classes' colours are on the frame; the globe grid spans most of them", drawn)
	}
	own := []tuimaps.RGB{{R: 20, G: 30, B: 90}, {R: 40, G: 70, B: 140}, {R: 60, G: 110, B: 170}, {R: 90, G: 150, B: 190}, {R: 130, G: 190, B: 200}, {R: 180, G: 220, B: 210}, {R: 230, G: 245, B: 230}}
	raw, m = hostFrame(t, hostType(own), tuimaps.Truecolor, tuimaps.RGB{R: 16, G: 16, B: 16})
	for i, c := range hostLegend(t, m).Classes {
		if c.Colour != own[i] {
			t.Errorf("class %d is keyed %v; the host gave %v", i, c.Colour, own[i])
		}
	}
	if groundsOf([]tuimaps.Class{{Colour: own[3]}}, raw) == 0 {
		t.Error("the host's own colours are not on the frame")
	}
}

// TestTheHostRampDefaultsWhenUnset is L-2.2: with no colours set, the low,
// middle and high tokens take the library's defaults on either ground; a
// palette that sets them moves the spread.
func TestTheHostRampDefaultsWhenUnset(t *testing.T) {
	for _, ground := range []tuimaps.RGB{{R: 16, G: 16, B: 16}, {R: 250, G: 250, B: 245}} {
		_, m := hostFrame(t, hostType(nil), tuimaps.Truecolor, ground)
		if c := hostLegend(t, m).Classes; c[0].Colour != (tuimaps.RGB{R: 38, G: 1, B: 95}) || !c[0].Drawn {
			t.Errorf("on %v the first class is %v (drawn %v); want the default low", ground, c[0].Colour, c[0].Drawn)
		}
	}
	m := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, m.FitWorld())
	m.ColourDepth(tuimaps.Truecolor)
	if _, err := m.SetPalette(map[string]tuimaps.RGB{"low": {R: 0, G: 40, B: 0}, "middle": {R: 0, G: 140, B: 0}, "high": {R: 160, G: 255, B: 160}}); err != nil {
		t.Fatal(err)
	}
	mustSet(t, m, hostType(nil))
	if c := hostLegend(t, m).Classes; c[0].Colour != (tuimaps.RGB{R: 0, G: 40, B: 0}) || c[6].Colour != (tuimaps.RGB{R: 160, G: 255, B: 160}) {
		t.Errorf("with low and high set, the classes run %v to %v", c[0].Colour, c[6].Colour)
	}
}

// TestAHostRampIsHeldToCheckRamp is L-2.2 and D-53: the default spread
// passes the checker for up to 7 classes at truecolor and 256 colours on
// both grounds, so the type of 7 is handed in with no warning.
func TestAHostRampIsHeldToCheckRamp(t *testing.T) {
	for _, depth := range []tuimaps.Depth{tuimaps.Truecolor, tuimaps.Colours256} {
		for _, ground := range []tuimaps.RGB{{R: 16, G: 16, B: 16}, {R: 250, G: 250, B: 245}} {
			_, m := hostFrame(t, hostType(nil), depth, ground)
			var ramp []tuimaps.RGB
			for _, c := range hostLegend(t, m).Classes {
				ramp = append(ramp, c.Colour)
			}
			if f := tuimaps.CheckRamp(ramp, tuimaps.RampCheck{Ground: ground, Depth: depth, Midpoint: -1}); len(f) != 0 {
				t.Errorf("depth %v on %v: the spread of 7 breaks %v", depth, ground, f)
			}
			for _, w := range m.Warnings() {
				if w.Kind == tuimaps.RampRuleBroken {
					t.Errorf("depth %v on %v: a ramp warning for the default spread: %+v", depth, ground, w)
				}
			}
		}
	}
}

// TestABadHostRampIsWarned is L-2.3: a host type whose own colours break the
// checker is drawn all the same, and said in Warnings, never silently.
func TestABadHostRampIsWarned(t *testing.T) {
	same := tuimaps.RGB{R: 120, G: 120, B: 120}
	raw, m := hostFrame(t, hostType([]tuimaps.RGB{same, same, same, same, same, same, same}), tuimaps.Truecolor, tuimaps.RGB{R: 16, G: 16, B: 16})
	if groundsOf([]tuimaps.Class{{Colour: same}}, raw) == 0 {
		t.Error("the host's colours were not drawn")
	}
	warned := false
	for _, w := range m.Warnings() {
		warned = warned || (w.Kind == tuimaps.RampRuleBroken && strings.Contains(w.Subject.String(), "host"))
	}
	if !warned {
		t.Errorf("seven classes in one colour, and no ramp warning names the overlay: %v", m.Warnings())
	}
}

// TestAHostRampIsWarnedOnceAnOverlay is watchpost D-216 (C-26): a host
// type set again with the same broken colours is one warning, not one a
// Set; set with colours that pass, its warning goes.
func TestAHostRampIsWarnedOnceAnOverlay(t *testing.T) {
	same := tuimaps.RGB{R: 120, G: 120, B: 120}
	bad := hostType([]tuimaps.RGB{same, same, same, same, same, same, same})
	m := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, m.FitWorld())
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Ground(tuimaps.RGB{R: 16, G: 16, B: 16}))
	count := func() int {
		n := 0
		for _, w := range m.Warnings() {
			if w.Kind == tuimaps.RampRuleBroken && strings.Contains(w.Subject.String(), "host") {
				n++
			}
		}
		return n
	}
	for range 3 {
		mustSet(t, m, bad)
	}
	if n := count(); n != 1 {
		t.Errorf("a broken host ramp set three times is %d warnings; want one", n)
	}
	for range 2 {
		mustSet(t, m, bad)
	}
	mustSet(t, m, hostType(nil))
	if n := count(); n != 0 {
		t.Errorf("after the overlay was set with the default spread, %d ramp warnings remain", n)
	}
}

// TestAHostTypesLegendHasColoursAndUnit is L-2.5: a host type's legend says
// its unit and keys each class in the colour it is drawn in; where there is
// no ramp to draw it in, no class claims a colour.
func TestAHostTypesLegendHasColoursAndUnit(t *testing.T) {
	_, m := hostFrame(t, hostType(nil), tuimaps.Truecolor, tuimaps.RGB{R: 16, G: 16, B: 16})
	e := hostLegend(t, m)
	if e.Unit != "MHz" {
		t.Errorf("the legend's unit is %q; the type's is MHz", e.Unit)
	}
	for i, c := range e.Classes {
		if !c.Drawn || c.Colour == (tuimaps.RGB{}) || c.Label == "" {
			t.Errorf("class %d is keyed %+v; each carries its colour and its label", i, c)
		}
	}
	_, m = hostFrame(t, hostType(nil), tuimaps.NoColour, tuimaps.RGB{R: 16, G: 16, B: 16})
	for i, c := range hostLegend(t, m).Classes {
		if c.Drawn {
			t.Errorf("with no colour, class %d claims to be drawn in %v", i, c.Colour)
		}
	}
}

// TestNoGridDrawsNothingSilently is L-2.4 (P0.1's finding): a host-type grid
// draws something - its bands, or its lines - at every depth, on both
// grounds, with its lines asked for or not.
func TestNoGridDrawsNothingSilently(t *testing.T) {
	for _, d := range specimenDepths {
		for _, g := range specimenGrounds {
			bare := bareWorld(t, d.depth, g.behind)
			for _, lines := range []bool{false, true} {
				o := hostType(nil)
				o.Grid.Lines = lines
				if specimenFrame(t, o, d.depth, g.behind) == bare {
					t.Errorf("%s on %s, lines %v: the host type draws nothing", d.name, g.name, lines)
				}
			}
		}
	}
}

// bareWorld is the specimen view with nothing set on it.
func bareWorld(t *testing.T, depth tuimaps.Depth, ground tuimaps.RGB) string {
	t.Helper()
	m := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, m.FitWorld())
	m.ColourDepth(depth)
	must(t, m.Ground(ground))
	settle(t, m)
	f, err := m.Render(specimenSize, noon)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\n") + "\n"
}

// TestAHostRampIsCheckedAgainWhenTheLookChanges is L-2.3 (watchpost RC-15):
// a host type whose colours pass at one depth and break the checker at
// another is warned when the depth changes after it was set, and the
// warning goes when the depth that passes comes back - never drawn broken
// unsaid because the check ran only at Set.
func TestAHostRampIsCheckedAgainWhenTheLookChanges(t *testing.T) {
	m := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, m.FitWorld())
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Ground(tuimaps.RGB{R: 16, G: 16, B: 16}))
	var greens []tuimaps.RGB // passes the checker at truecolor; at 256 colours its steps fall together
	for i := range 7 {
		greens = append(greens, tuimaps.RGB{G: uint8(40 + 28*i)})
	}
	mustSet(t, m, hostType(greens))
	warned := func() bool {
		for _, w := range m.Warnings() {
			if w.Kind == tuimaps.RampRuleBroken && strings.Contains(w.Subject.String(), "host") {
				return true
			}
		}
		return false
	}
	if warned() {
		t.Fatalf("the greens break the checker at truecolor: %v", m.Warnings())
	}
	m.ColourDepth(tuimaps.Colours256)
	if !warned() {
		t.Fatal("the depth changed to 256 colours after the Set, and no ramp warning names the overlay")
	}
	m.ColourDepth(tuimaps.Truecolor)
	m.ColourDepth(tuimaps.Colours256) // warned again, untaken: warnings are taken once
	m.ColourDepth(tuimaps.Truecolor)
	if warned() {
		t.Error("back at truecolor, where the ramp passes, its warning stays")
	}
}

// TestEveryLookSetterChecksTheRampAgain is A-18 and watchpost R2C-6: each of
// the five setters that change the look - SetPalette, SafeRamps, Ground,
// PaintGround and ColourDepth - checks a host ramp again, so its warning
// follows the look it is drawn at. Warnings are taken once, so a clearing
// step first breaks the look and puts it back before reading: a setter that
// skipped the check would leave the breaking look's warning to be read.
func TestEveryLookSetterChecksTheRampAgain(t *testing.T) {
	step := func(m *tuimaps.Map, name string, want bool) {
		t.Helper()
		if got := rampWarned(m); got != want {
			t.Errorf("after %s the ramp warning is %v; want %v", name, got, want)
		}
	}
	palette := func(m *tuimaps.Map, tokens map[string]tuimaps.RGB) {
		t.Helper()
		if _, err := m.SetPalette(tokens); err != nil {
			t.Fatal(err)
		}
	}
	filled := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, filled.FitWorld())
	filled.ColourDepth(tuimaps.Truecolor)
	mustSet(t, filled, hostType(nil)) // filled from low, middle and high
	step(filled, "the Set, at the library's colours", false)
	grey := tuimaps.RGB{R: 128, G: 128, B: 128}
	greys := map[string]tuimaps.RGB{"low": grey, "middle": grey, "high": grey}
	palette(filled, greys)
	step(filled, "SetPalette with one grey for low, middle and high", true)
	palette(filled, greys)
	palette(filled, nil)
	step(filled, "SetPalette with nothing, the library's colours back", false)
	palette(filled, greys)
	filled.SafeRamps(true)
	step(filled, "SafeRamps on, the library's own ramp colours", false)
	filled.SafeRamps(false)
	step(filled, "SafeRamps off, the grey again", true)

	own := world(t, specimenSize.Cols, specimenSize.Rows)
	must(t, own.FitWorld())
	own.ColourDepth(tuimaps.Truecolor)
	var greens []tuimaps.RGB
	for i := range 7 {
		greens = append(greens, tuimaps.RGB{G: uint8(40 + 28*i)})
	}
	mustSet(t, own, hostType(greens))
	step(own, "the Set of the host's greens", false)
	must(t, own.Ground(greens[3]))
	step(own, "Ground at a class's own colour", true)
	must(t, own.Ground(greens[3]))
	own.PaintGround()
	step(own, "PaintGround, the library's ground", false)
	must(t, own.Ground(greens[3]))
	must(t, own.Ground(tuimaps.RGB{R: 16, G: 16, B: 16}))
	step(own, "Ground back to a dark one", false)
	own.ColourDepth(tuimaps.Colours256)
	own.ColourDepth(tuimaps.Truecolor)
	step(own, "ColourDepth at 256 colours and back at truecolor", false)
	own.ColourDepth(tuimaps.Colours256)
	step(own, "ColourDepth at 256 colours, where the greens fall together", true)
}

// rampWarned reports whether the warnings, taken, name the host overlay's
// ramp.
func rampWarned(m *tuimaps.Map) bool {
	for _, w := range m.Warnings() {
		if w.Kind == tuimaps.RampRuleBroken && strings.Contains(w.Subject.String(), "host") {
			return true
		}
	}
	return false
}

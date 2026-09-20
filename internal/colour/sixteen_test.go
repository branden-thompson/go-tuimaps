package colour

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestSixteenPalette is plan task 08.14 (D-59, D-79, specimen 23): the
// basemap has its own small palette chosen from the sixteen by hand, and
// anything that needs a ramp has none, so its no-colour form is drawn.
func TestSixteenPalette(t *testing.T) {
	seen := map[Token]uint8{}
	for _, tok := range Tokens() {
		for _, kind := range []GroundKind{Dark, Light} {
			index, ok := Sixteen(tok, kind)
			isRamp := tok >= AlertExtremeOutline && tok <= High
			if ok == isRamp {
				t.Errorf("%s on a %v ground: has a 16-colour entry = %v; ramps have none, everything else has one", tok.Name(), kind, ok)
			}
			if ok && index > 15 {
				t.Errorf("%s: entry %d", tok.Name(), index)
			}
			if kind == Dark {
				seen[tok] = index
			}
		}
	}
	// What specimen 23 chose and HUM LEAD found good: bright cyan rivers,
	// green parks, bright white borders and labels, a bright yellow marker,
	// on a black ground.
	for tok, want := range map[Token]uint8{Ground: 0, River: 14, Park: 2, BorderCountry: 15, LabelPlace: 15, Marker: 11} {
		if seen[tok] != want {
			t.Errorf("%s is entry %d on a dark ground, want %d", tok.Name(), seen[tok], want)
		}
	}
	var none Palette
	if c, ok := none.ResolveAt(River, Dark, Colours16); !ok || c != SixteenReference()[14] {
		t.Errorf("a river at 16 colours: %v, %v", c, ok)
	}
	if _, ok := none.ResolveAt(Radar1+2, Dark, Colours16); ok {
		t.Error("a ramp token has a colour at 16 colours; its no-colour form is what is drawn")
	}
	host, _ := NewPalette(map[string]RGB{"river": {250, 10, 10}})
	if c, _ := host.ResolveAt(River, Dark, Colours16); c != (RGB{250, 10, 10}) {
		t.Errorf("a host's river at 16 colours: %v; the renderer maps it to the nearest entry", c)
	}
	if i, shown := ToSixteen(RGB{250, 10, 10}); i != 9 || shown != SixteenReference()[9] {
		t.Errorf("a strong red is entry %d, %v", i, shown)
	}
}

// TestSixteenReferenceTable is plan task 08.19 (PL-AX-6): at 16 colours the
// terminal owns the colours, so checks use a stated reference table and are
// indicative only. Against it every line reads on its ground, and roads and
// borders differ by more than intensity.
func TestSixteenReferenceTable(t *testing.T) {
	table := SixteenReference()
	if table[0] != (RGB{}) || table[15] != (RGB{255, 255, 255}) || table[7] == table[15] {
		t.Fatalf("the reference table: %v", table)
	}
	lines := []Token{Coast, WaterLine, River, BorderCountry, BorderRegion, RoadMajor, RoadMinor, Rail, Park, Runway, LabelPlace, LabelWater, LabelRegion, Marker}
	for _, kind := range []GroundKind{Dark, Light} {
		g, _ := Sixteen(Ground, kind)
		for _, tok := range lines {
			i, _ := Sixteen(tok, kind)
			if got := Contrast(table[i], table[g]); got < LineContrast {
				t.Errorf("%v ground: %s (entry %d) is %.1f:1 on the reference table", kind, tok.Name(), i, got)
			}
		}
		road, _ := Sixteen(RoadMajor, kind)
		border, _ := Sixteen(BorderCountry, kind)
		r, b := table[road], table[border]
		grey := func(c RGB) bool { return c.R == c.G && c.G == c.B }
		if road == border || (grey(r) && grey(b)) {
			t.Errorf("%v ground: roads (entry %d) and borders (entry %d) differ only in intensity", kind, road, border)
		}
	}
}

// TestNoColourSelectedByEnv is plan task 08.15 (NFR-15).
func TestNoColourSelectedByEnv(t *testing.T) {
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "NO_COLOR" {
				return v
			}
			return ""
		}
	}
	if got := ChooseDepth(Truecolor, false, env("1")); got != NoColour {
		t.Errorf("NO_COLOR set and no hint: %v", got)
	}
	if got := ChooseDepth(Truecolor, false, env("")); got != Truecolor {
		t.Errorf("NO_COLOR empty and no hint: %v; an empty value is not a request", got)
	}
	if got := ChooseDepth(Colours256, true, env("1")); got != Colours256 {
		t.Errorf("a hint from the host and NO_COLOR set: %v; the host knows its own output", got)
	}
	if got := ChooseDepth(Depth(9), true, env("")); got != NoColour {
		t.Errorf("a hint that is no depth: %v; with nothing known, no colour is the safe form", got)
	}
	if got := ChooseDepth(Truecolor, false, nil); got != Truecolor {
		t.Errorf("no environment to ask: %v", got)
	}
}

// TestOverrideIsWarnedNotRefused is plan task 08.11 (D-69, D-53), and
// TestSafeRampsForcesPreset is 08.12 (D-63).
func TestOverrideIsWarnedNotRefused(t *testing.T) {
	var none Palette
	if w := none.Warnings(Dark, Truecolor); len(w) != 0 {
		t.Errorf("the library's own colours: %+v", w)
	}
	// A host makes the two heaviest radar classes the same red.
	host, _ := NewPalette(map[string]RGB{"radar.5": {200, 0, 0}, "radar.6": {200, 0, 0}})
	if c, _ := host.ResolveAt(Radar6, Dark, Truecolor); c != (RGB{200, 0, 0}) {
		t.Errorf("the host's colour was refused: %v", c)
	}
	w := host.Warnings(Dark, Truecolor)
	if len(w) != 1 || w[0].Kind != fault.RampRuleBroken || w[0].Count < 2 || w[0].Subject.String() != "radar" {
		t.Errorf("warnings %+v; want one for the radar ramp, counting its breaches", w)
	}
	basemap, _ := NewPalette(map[string]RGB{"road.major": {255, 0, 0}})
	if w := basemap.Warnings(Dark, Truecolor); len(w) != 0 {
		t.Errorf("a basemap colour is not a ramp: %+v", w)
	}
}

func TestSafeRampsForcesPreset(t *testing.T) {
	host, _ := NewPalette(map[string]RGB{"radar.6": {200, 0, 0}, "road.major": {255, 0, 0}, "alert.severe.tint": {1, 2, 3}})
	safe := host.SafeRamps(true)
	preset, _ := Ramp(Radar, Dark, Truecolor)
	if c, _ := safe.ResolveAt(Radar6, Dark, Truecolor); c != preset[5] {
		t.Errorf("radar.6 with safe ramps on: %v, want the preset's %v", c, preset[5])
	}
	if c, _ := safe.Resolve(AlertSevereTint, Dark); c != (RGB{112, 36, 28}) {
		t.Errorf("an alert tint with safe ramps on: %v", c)
	}
	if c, _ := safe.Resolve(RoadMajor, Dark); c != (RGB{255, 0, 0}) {
		t.Errorf("a basemap colour with safe ramps on: %v; safe ramps is about ramps", c)
	}
	if w := safe.Warnings(Dark, Truecolor); len(w) != 0 {
		t.Errorf("with safe ramps on nothing of the host's is in a ramp: %+v", w)
	}
	if c, _ := host.ResolveAt(Radar6, Dark, Truecolor); c != (RGB{200, 0, 0}) {
		t.Error("turning safe ramps on changed the palette it was made from")
	}
}

package tuimaps_test

// blend_test.go — v0.2.0 L3.6-L3.8 (L-11): inside an alert's area, the
// radar blends toward the tint as strongly as the search found safe, or on a
// ground where nothing is safe, is drawn over it.

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// blendScene is a severe alert over the middle of the map and heavy radar
// over all of it, labels and names included, in truecolour, over the embedded tiles; its frame's text.
func blendScene(t *testing.T, before func(m *tuimaps.Map)) string {
	t.Helper()
	m := world(t, 40, 12)
	m.ColourDepth(tuimaps.Truecolor)
	if before != nil {
		before(m)
	}
	must(t, m.Recentre(tuimaps.LonLat{Lon: -90, Lat: 35}))
	must(t, m.Zoom(3))
	mustSet(t, m, warningAt("alerts", -94, 32, -86, 38))
	mustSet(t, m, rainOver(t, -130, 10, -50, 60, color.NRGBA{R: 200, A: 255}))
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(f.Lines, "\\n")
}

func bgCode(c colour.RGB) string { return fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B) }

// radarColours are the radar ramp and the severe tint on a ground, as the
// library's own palette resolves them.
func radarColours(ground colour.GroundKind) ([]colour.RGB, colour.RGB) {
	var none colour.Palette
	var ramp []colour.RGB
	for i := range 6 {
		c, _ := none.ResolveAt(colour.Radar1+colour.Token(i), ground, colour.Truecolor)
		ramp = append(ramp, c)
	}
	tint, _ := none.ResolveAt(colour.AlertSevereTint, ground, colour.Truecolor)
	return ramp, tint
}

// TestTheTintBlendsOverTheRadar is L3.6 (L-11.1): on the dark ground, the
// cells where the alert and the radar meet are the radar's class colour
// shifted toward the tint by the strength the search chose; none is the bare
// tint, which would hide the radar.
func TestTheTintBlendsOverTheRadar(t *testing.T) {
	frame := blendScene(t, nil)
	ramp, tint := radarColours(colour.Dark)
	var none colour.Palette
	var groundColour colour.RGB
	groundColour, _ = none.ResolveAt(colour.Ground, colour.Dark, colour.Truecolor)
	s, blended := colour.SearchBlends(none, colour.Dark, groundColour, colour.Truecolor).Strength(colour.Radar, 1)
	if !blended {
		t.Fatal("the severe tint does not blend over radar on the dark ground, so this proves nothing")
	}
	blends, plain := 0, 0
	for _, c := range ramp {
		blends += strings.Count(frame, bgCode(colour.Blend(c, tint, s)))
		plain += strings.Count(frame, bgCode(c))
	}
	if blends == 0 {
		t.Errorf("no cell is radar blended toward the tint at %.2f", s)
	}
	if plain == 0 {
		t.Error("no cell outside the alert is plain radar, so the scene is not what it says")
	}
	if n := strings.Count(frame, bgCode(tint)); n != 0 {
		t.Errorf("%d cells are the bare tint over radar; the radar is hidden there", n)
	}
}

// TestOnTheLightGroundTheRadarIsDrawnOverTheTint is L3.8's fallback (L-11.4,
// D-27): on the light ground no strength separates, so inside the area the
// radar is drawn as it is outside, never blended and never hidden.
func TestOnTheLightGroundTheRadarIsDrawnOverTheTint(t *testing.T) {
	frame := blendScene(t, func(m *tuimaps.Map) { must(t, m.Ground(tuimaps.RGB{R: 250, G: 250, B: 245})) })
	ramp, tint := radarColours(colour.Light)
	plain := 0
	for _, c := range ramp {
		plain += strings.Count(frame, bgCode(c))
	}
	if plain == 0 {
		t.Error("no cell is plain radar on the light ground")
	}
	if n := strings.Count(frame, bgCode(tint)); n != 0 {
		t.Errorf("%d cells are the bare tint over radar on the light ground", n)
	}
	for s := 0.10; s <= 0.501; s += 0.05 {
		for _, c := range ramp {
			if n := strings.Count(frame, bgCode(colour.Blend(c, tint, s))); n != 0 && colour.Blend(c, tint, s) != c {
				t.Errorf("radar blended toward the tint at %.2f on the light ground, where no strength separates", s)
			}
		}
	}
}

// TestAHostPaletteThatBreaksTheBlendIsReported is L3.8's report (D-72): a
// palette set after New re-runs the search, and a tint it leaves unable to
// blend is a warning.
func TestAHostPaletteThatBreaksTheBlendIsReported(t *testing.T) {
	ramp, _ := radarColours(colour.Dark)
	twin := ramp[3]
	twin.R++
	var warnings []tuimaps.Warning
	blendScene(t, func(m *tuimaps.Map) {
		if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil { // the library's own colours first
			t.Fatal(err)
		}
		if w := m.Warnings(); len(w) != 0 {
			t.Errorf("the library's own palette warned: %v", w)
		}
		if _, err := m.SetPalette(map[string]tuimaps.RGB{"radar.3": twin}); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
			t.Fatal(err)
		}
		warnings = m.Warnings()
	})
	found := 0
	for _, w := range warnings {
		if w.Kind == fault.RampRuleBroken && strings.Contains(w.Subject.String(), "radar under the") {
			found++
		}
	}
	if found != 5 {
		t.Errorf("a radar ramp with two classes one colour: %d tints reported as drawn over, want all 5; warnings %v", found, warnings)
	}
}

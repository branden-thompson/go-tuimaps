package tuimaps_test

// airuv_test.go — v0.2.0 L-25 (watchpost D-137 to D-140): the UV index and
// the US AQI, two presets in their official scales' hues, set to pass the
// checker; and a monitor's AQI as a marker in its category's colour.

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// flatField is one value over a box.
func flatField(v float64) tuimaps.Grid {
	g := tuimaps.Grid{West: -125, South: 24, East: -66, North: 50, Cols: 8, Rows: 8}
	for range 64 {
		g.Values = append(g.Values, v)
	}
	return g
}

// TestUVAndAirQualityAreTheirScales is L-25.1: a UV grid keys five classes,
// Low to Extreme, broken at 3, 6, 8 and 11; an AQI grid six, Good to
// Hazardous, broken at 51, 101, 151, 201 and 301; each drawn.
func TestUVAndAirQualityAreTheirScales(t *testing.T) {
	for _, c := range []struct {
		o       tuimaps.Overlay
		preset  string
		classes int
		first   string
	}{{tuimaps.UVGrid("uv", flatField(7), noon), "uv", 5, "3"}, {tuimaps.AirQualityGrid("aqi", flatField(120), noon), "aqi", 6, "51"}} {
		m := reportMap(t)
		m.ColourDepth(tuimaps.Truecolor)
		mustSet(t, m, c.o)
		settle(t, m)
		found := false
		for _, e := range m.Legend() {
			if e.Preset == c.preset {
				found = true
				if len(e.Classes) != c.classes || !strings.Contains(e.Classes[0].Label, c.first) {
					t.Errorf("%s's legend is %+v; want %d classes, the first under %s", c.preset, e, c.classes, c.first)
				}
				if groundsOf(e.Classes, momentFrame(t, m, noon)) == 0 {
					t.Errorf("%s was not drawn", c.preset)
				}
			}
		}
		if !found {
			t.Errorf("the legend has no %s", c.preset)
		}
	}
}

// TestUVAndAirQualityAreChecked is L-25.2: each preset in its one unit, or
// the overlay is refused.
func TestUVAndAirQualityAreChecked(t *testing.T) {
	m := reportMap(t)
	for _, ty := range []tuimaps.Type{{Preset: "uv", Unit: "F"}, {Preset: "aqi", Unit: "ppm"}} {
		g := flatField(5)
		g.Type = ty
		if _, err := m.Set(tuimaps.Overlay{ID: "x", Valid: noon, Keeps: 3600e9, Grid: &g}); !isKind(err, fault.UnknownPreset) {
			t.Errorf("%+v: %v; want refused", ty, err)
		}
	}
}

// TestAMonitorIsDrawnInItsCategory is L-25.3: AirQualityRole is a reading's
// category - 42 Good, 51 Moderate, 150 Unhealthy for Sensitive Groups, 301
// Hazardous - and a marker in it keeps its words in a colour that reads.
func TestAMonitorIsDrawnInItsCategory(t *testing.T) {
	names := map[float64]string{42: "aqi.1", 51: "aqi.2", 150: "aqi.3", 200: "aqi.4", 250: "aqi.5", 301: "aqi.6"}
	for aqi, want := range names {
		if got := tuimaps.AirQualityRole(aqi).Name(); got != want {
			t.Errorf("an AQI of %v is %s; want %s", aqi, got, want)
		}
	}
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	if unknown, err := m.SetPalette(map[string]tuimaps.RGB{"marker.label": {R: 3, G: 250, B: 7}}); err != nil || len(unknown) != 0 {
		t.Fatal(unknown, err)
	}
	mustSet(t, m, tuimaps.Overlay{ID: "airnow", Valid: noon, Keeps: 3600e9, Features: []tuimaps.Feature{
		{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -90, Lat: 35.1}}}, Role: tuimaps.AirQualityRole(42), Label: "Memphis 42"}}})
	settle(t, m)
	raw := momentFrame(t, m, noon)
	if frame := colours.ReplaceAllString(raw, ""); !strings.Contains(frame, "Memphis 42") {
		t.Errorf("the monitor's words are not drawn:\n%s", frame)
	}
	if !strings.Contains(raw, "38;2;3;250;7") {
		t.Error("the monitor's words are not in the markers' ink: a pale class's colour does not read on a light ground")
	}
}

// A UV READING IS DRAWN IN ITS BAND (watchpost W18.4, its D-167): EPA's UV
// index for a city, as a marker, in the UV preset's own colour for its band -
// Low under 3, Moderate to 6, High to 8, Very High to 11, Extreme past it.
func TestAUVReadingIsDrawnInItsBand(t *testing.T) {
	names := map[float64]string{0: "uv.1", 2.9: "uv.1", 3: "uv.2", 6: "uv.3", 8: "uv.4", 11: "uv.5", 14: "uv.5"}
	for index, want := range names {
		if got := tuimaps.UVRole(index).Name(); got != want {
			t.Errorf("a UV index of %v is %s; want %s", index, got, want)
		}
	}
}

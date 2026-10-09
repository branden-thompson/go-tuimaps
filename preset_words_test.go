package tuimaps_test

// preset_words_test.go — v0.3.0 L-2.1: every preset's legend says its unit
// and puts each class in words, as a listener who cannot tell the colours
// apart reads it.

import (
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestEveryPresetHasLegendWordsAndAUnit is L-2.1's second half: each preset,
// set over the lower 48, keys every class in words and names its unit; the
// MUF and foF2 presets key the bands D-144 and D-145 rule.
func TestEveryPresetHasLegendWordsAndAUnit(t *testing.T) {
	for _, c := range []struct {
		o       tuimaps.Overlay
		preset  string
		classes int
		unit    string
	}{
		{tuimaps.TemperatureGrid("p", flatField(20), tuimaps.Celsius, noon), "temperature", 17, "C"},
		{tuimaps.UVGrid("p", flatField(5), noon), "uv", 5, "index"},
		{tuimaps.AirQualityGrid("p", flatField(60), noon), "aqi", 6, "AQI"},
		{tuimaps.QPFGrid("p", flatField(10), noon), "qpf", 8, "mm"}, // the trace below the first floor is keyed, not drawn
		{tuimaps.WaveGrid("p", flatField(3), tuimaps.Feet, noon), "waves", 6, "ft"},
		{tuimaps.MUFGrid("p", flatField(15), noon), "muf", 10, "MHz"},
		{tuimaps.FoF2Grid("p", flatField(6), noon), "fof2", 7, "MHz"},
	} {
		m := reportMap(t)
		m.ColourDepth(tuimaps.Truecolor)
		mustSet(t, m, c.o)
		settle(t, m)
		found := false
		for _, e := range m.Legend() {
			if e.Preset != c.preset {
				continue
			}
			found = true
			if e.Unit != c.unit || len(e.Classes) != c.classes {
				t.Errorf("%s: unit %q, %d classes; want %q and %d", c.preset, e.Unit, len(e.Classes), c.unit, c.classes)
			}
			for i, k := range e.Classes {
				if k.Label == "" {
					t.Errorf("%s: class %d has no words", c.preset, i)
				}
			}
		}
		if !found {
			t.Errorf("the legend has no %s", c.preset)
		}
	}
}

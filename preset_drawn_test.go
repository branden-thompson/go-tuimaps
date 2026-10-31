package tuimaps_test

// preset_drawn_test.go — every class of every preset is drawn, not only
// keyed. A legend can name a class whose colour the frame never paints: v0.2.0
// keyed the rain totals' two heaviest classes and drew neither (A-5).

import (
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestEveryClassOfEveryPresetIsDrawn sets a flat field inside each class of
// each filled preset and asks that the frame paint cells in that class's own
// colour.
func TestEveryClassOfEveryPresetIsDrawn(t *testing.T) {
	presets := []struct {
		preset string
		make   func(v float64) tuimaps.Overlay
		values []float64 // one inside each drawn class
	}{
		{"temperature", func(v float64) tuimaps.Overlay {
			return tuimaps.TemperatureGrid("p", flatField(v), tuimaps.Celsius, noon)
		},
			[]float64{-40, -27, -22, -17, -12, -7, -2, 3, 8, 13, 18, 23, 28, 33, 38, 43, 50}},
		{"uv", func(v float64) tuimaps.Overlay { return tuimaps.UVGrid("p", flatField(v), noon) }, []float64{1, 4, 7, 9, 12}},
		{"aqi", func(v float64) tuimaps.Overlay { return tuimaps.AirQualityGrid("p", flatField(v), noon) }, []float64{20, 70, 120, 170, 250, 400}},
		{"qpf", func(v float64) tuimaps.Overlay { return tuimaps.QPFGrid("p", flatField(v), noon) }, []float64{1, 4, 9, 18, 35, 70, 150}},
		{"muf", func(v float64) tuimaps.Overlay { return tuimaps.MUFGrid("p", flatField(v), noon) }, []float64{2, 4, 6, 8, 12, 16, 20, 23, 26, 30}},
		{"fof2", func(v float64) tuimaps.Overlay { return tuimaps.FoF2Grid("p", flatField(v), noon) }, []float64{2.5, 4, 6, 8, 12, 15}}, // below 1.8 MHz nothing is drawn (D-149)
	}
	for _, p := range presets {
		for _, v := range p.values {
			m := reportMap(t)
			m.ColourDepth(tuimaps.Truecolor)
			mustSet(t, m, p.make(v))
			settle(t, m)
			raw := momentFrame(t, m, noon)
			var class []tuimaps.Class
			for _, e := range m.Legend() {
				if e.Preset == p.preset {
					class = e.Classes
				}
			}
			if cellsIn(class, raw) == 0 {
				t.Errorf("%s at %v: no cell is painted in any of its classes' colours", p.preset, v)
			}
		}
	}
}

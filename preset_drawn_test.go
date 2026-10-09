package tuimaps_test

// preset_drawn_test.go — every class of every preset is drawn, not only
// keyed. A legend can name a class whose colour the frame never paints: v0.2.0
// keyed the rain totals' two heaviest classes and drew neither (v0.2.1).

import (
	"strconv"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// cellsIn counts the frame's cells whose background is one of the classes'
// colours: every character printed while such a background is in effect.
func cellsIn(classes []tuimaps.Class, raw string) int {
	want := map[[3]int]bool{}
	for _, c := range classes {
		want[[3]int{int(c.Colour.R), int(c.Colour.G), int(c.Colour.B)}] = true
	}
	n, bg, on := 0, [3]int{}, false
	for i := 0; i < len(raw); {
		if strings.HasPrefix(raw[i:], "\x1b[") {
			end := strings.IndexByte(raw[i:], 'm')
			if end < 0 {
				break
			}
			ps := strings.Split(raw[i+2:i+end], ";")
			for j := 0; j < len(ps); j++ {
				switch ps[j] {
				case "", "0", "49":
					on = false
				case "38":
					j += 4
				case "48":
					if j+4 < len(ps) {
						r, _ := strconv.Atoi(ps[j+2])
						g, _ := strconv.Atoi(ps[j+3])
						b, _ := strconv.Atoi(ps[j+4])
						bg, on = [3]int{r, g, b}, true
					}
					j += 4
				}
			}
			i += end + 1
			continue
		}
		r := []rune(raw[i:min(i+4, len(raw))])[0]
		if r != '\n' && on && want[bg] {
			n++
		}
		i += len(string(r))
	}
	return n
}

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

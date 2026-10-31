package tuimaps_test

// water_test.go — v0.3.0 L-1 (watchpost D-32): a host chooses, per overlay,
// whether a field continues over the sea and whether the sea masks an image.
// The defaults are v0.2.0's: a field stops at the shore, an image is never
// masked.

import (
	"strconv"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// seaField is a temperature field over the open Gulf of Mexico alone, rising
// eastwards, so its bands and their lines have somewhere to fall.
func seaField(overWater, lines bool) tuimaps.Overlay {
	g := tuimaps.Grid{West: -94, South: 23.5, East: -85, North: 27.5, Cols: 12, Rows: 6, OverWater: overWater, Lines: lines}
	for range g.Rows {
		for c := range g.Cols {
			g.Values = append(g.Values, 10+float64(c)*2)
		}
	}
	return tuimaps.TemperatureGrid("sea", g, tuimaps.Celsius, noon)
}

// maskedRadar is loopAt's single picture, over the Southeast and the sea
// beside it, with the host's water choice.
func maskedRadar(t *testing.T, masked bool) tuimaps.Overlay {
	o := loopAt(t, "radar", []int{0}, nil, nil)
	o.Image.MaskedByWater = masked
	return o
}

// openGulf is a truecolour map of the open Gulf of Mexico, momentFrame's
// size: seaField fills the sea between the coasts.
func openGulf(t *testing.T) *tuimaps.Map {
	t.Helper()
	m := world(t, 80, 24)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -90, Lat: 25.5}))
	must(t, m.Zoom(4))
	return m
}

// gulfFrame is openGulf's frame with one overlay set, or none.
func gulfFrame(t *testing.T, o *tuimaps.Overlay) (*tuimaps.Map, string) {
	t.Helper()
	m := openGulf(t)
	if o != nil {
		mustSet(t, m, *o)
	}
	settle(t, m)
	return m, momentFrame(t, m, noon)
}

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

// legendClasses is the classes of a preset's legend entry.
func legendClasses(m *tuimaps.Map, preset string) []tuimaps.Class {
	var out []tuimaps.Class
	for _, e := range m.Legend() {
		if e.Preset == preset {
			out = append(out, e.Classes...)
		}
	}
	return out
}

// TestAFieldStopsAtTheShoreByDefault is L-1.1's default (D-32): a field over
// the open sea draws nothing unless its host says it continues over water.
func TestAFieldStopsAtTheShoreByDefault(t *testing.T) {
	stops := seaField(false, false)
	m, raw := gulfFrame(t, &stops)
	if n := cellsIn(legendClasses(m, "temperature"), raw); n != 0 {
		t.Errorf("a field over the sea painted %d runs of cells; by default it stops at the shore", n)
	}
	over := seaField(true, false)
	m, raw = gulfFrame(t, &over)
	if cellsIn(legendClasses(m, "temperature"), raw) == 0 {
		t.Error("a field the host continues over water painted nothing over the sea")
	}
}

// TestHostCanFlipEither is L-1.1 and L-1.2: each default flips per overlay,
// and neither reaches the other kind - a field over water leaves an image's
// cells as they were, and beside an image masked by water a field's own
// choice still decides where it is drawn.
func TestHostCanFlipEither(t *testing.T) {
	m := rainMap(t)
	mustSet(t, m, maskedRadar(t, false))
	settle(t, m)
	unmasked := cellsIn(legendClasses(m, "radar"), momentFrame(t, m, noon))
	mustSet(t, m, maskedRadar(t, true))
	settle(t, m)
	masked := cellsIn(legendClasses(m, "radar"), momentFrame(t, m, noon))
	if unmasked == 0 || masked == 0 || masked >= unmasked {
		t.Errorf("rain coloured %d cells unmasked and %d masked; masking the sea takes the sea's cells and keeps the land's", unmasked, masked)
	}
	over := warmField()
	over.OverWater = true
	mustSet(t, m, maskedRadar(t, false))
	mustSet(t, m, tuimaps.TemperatureGrid("temp", over, tuimaps.Celsius, noon))
	settle(t, m)
	if n := cellsIn(legendClasses(m, "radar"), momentFrame(t, m, noon)); n != unmasked {
		t.Errorf("rain coloured %d cells beside a field over water; alone, %d: the field's choice reached the image", n, unmasked)
	}
	mustSet(t, m, maskedRadar(t, true))
	settle(t, m)
	withOver := momentFrame(t, m, noon)
	mustSet(t, m, tuimaps.TemperatureGrid("temp", warmField(), tuimaps.Celsius, noon))
	settle(t, m)
	if momentFrame(t, m, noon) == withOver {
		t.Error("beside an image masked by water, the field's own choice changed nothing")
	}
}

// TestContoursFollowTheirFieldOverWater is L-1.3: a field's lines go where
// its bands go - not over the sea by default, over it when the host says so.
func TestContoursFollowTheirFieldOverWater(t *testing.T) {
	_, bare := gulfFrame(t, nil)
	stops := seaField(false, true)
	if _, raw := gulfFrame(t, &stops); raw != bare {
		t.Error("a lined field over the sea drew there although it stops at the shore")
	}
	over := seaField(true, true)
	_, raw := gulfFrame(t, &over)
	if raw == bare {
		t.Fatal("a lined field the host continues over water drew nothing over the sea")
	}
	if colours.ReplaceAllString(raw, "") == colours.ReplaceAllString(bare, "") {
		t.Error("a lined field over water changed colours only: its lines and their labels are not drawn")
	}
}

// TestAChangedWaterChoiceRedraws is L-1.4: the same overlay set again with
// the other choice draws a new frame; frame reuse never keeps the old one.
func TestAChangedWaterChoiceRedraws(t *testing.T) {
	m := openGulf(t)
	mustSet(t, m, seaField(false, false))
	settle(t, m)
	before := momentFrame(t, m, noon)
	mustSet(t, m, seaField(true, false))
	settle(t, m)
	if after := momentFrame(t, m, noon); after == before || cellsIn(legendClasses(m, "temperature"), after) == 0 {
		t.Error("the field set again over water kept the frame drawn before")
	}
	mustSet(t, m, seaField(false, false))
	settle(t, m)
	if again := momentFrame(t, m, noon); again != before {
		t.Error("the field set back to stop at the shore did not draw as it first did")
	}
}

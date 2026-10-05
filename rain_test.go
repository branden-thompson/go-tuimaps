package tuimaps_test

// rain_test.go — v0.2.0 L-17 (watchpost D-115 to D-117): a grid in radar's
// scale is drawn as rain is - in its own colours, over the sea and over any
// field - and a grid's marks, the host's text a value, are written on the map
// at the arrows' spacing.

import (
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// rainField is 40 dBZ over the lower 48 and the Gulf, the sea south of the
// coast among it.
func rainField() tuimaps.Grid {
	const cols, rows = 24, 24
	g := tuimaps.Grid{West: -125, South: 20, East: -66, North: 50, Cols: cols, Rows: rows, Type: tuimaps.Type{Preset: "radar", Unit: "dBZ"}}
	for range cols * rows {
		g.Values = append(g.Values, 40)
	}
	return g
}

// rainGrounds counts the frame's cells painted in one of radar's class
// colours, as backgrounds.
func rainGrounds(m *tuimaps.Map, raw string) int {
	n := 0
	for _, e := range m.Legend() {
		if e.Preset == "radar" {
			n += groundsOf(e.Classes, raw)
		}
	}
	return n
}

// rainMap is a truecolour map of the lower 48 and the Gulf.
func rainMap(t *testing.T) *tuimaps.Map {
	t.Helper()
	m := world(t, 120, 40)
	m.ColourDepth(tuimaps.Truecolor)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -92, Lat: 32}))
	must(t, m.Zoom(4))
	return m
}

// TestARadarGridIsDrawnAsRain is L-17.1: a grid in radar's scale keeps its
// own colours at full strength - never the lined look, though it asks for it
// or shares the map - over the sea as over land, because rain falls on the
// sea; and a field beside it takes its lines, as beside an image, because
// both ramps run blue to red.
func TestARadarGridIsDrawnAsRain(t *testing.T) {
	m := rainMap(t)
	mustSet(t, m, tuimaps.TemperatureGrid("temp", warmField(), tuimaps.Celsius, noon))
	settle(t, m)
	if bandGrounds(m, momentFrame(t, m, noon)) == 0 {
		t.Fatal("the temperature alone painted no band: nothing to compare")
	}
	rain := rainField()
	rain.Lines = true // asked, and never taken: rain is never faint
	mustSet(t, m, tuimaps.Overlay{ID: "rain", Valid: noon, Keeps: 12 * time.Hour, Grid: &rain})
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 120, Rows: 40}, noon)
	if err != nil {
		t.Fatal(err)
	}
	raw := strings.Join(f.Lines, "\n")
	if rainGrounds(m, raw) == 0 {
		t.Error("the rain painted no cell in radar's own colours: it was drawn faint, or not at all")
	}
	if n := bandGrounds(m, raw); n != 0 {
		t.Errorf("beside the rain %d cells keep a temperature band's own colour; the field is its lines there", n)
	}

	sea := rainField() // the Gulf's open water alone
	sea.West, sea.South, sea.East, sea.North = -94, 24, -86, 28.5
	mustSet(t, m, tuimaps.Overlay{ID: "rain", Valid: noon, Keeps: 12 * time.Hour, Grid: &sea})
	settle(t, m)
	raw = momentFrame(t, m, noon)
	if rainGrounds(m, raw) == 0 {
		t.Error("rain over the sea was not drawn: it stopped at the shore, as temperature does")
	}
	if n := bandGrounds(m, raw); n != 0 {
		t.Errorf("with rain on the sea alone, %d cells of land keep a temperature band's own colour; the field is its lines while rain shares the map", n)
	}
}

// TestAGridsMarksAreWrittenOnTheMap is L-17.2: a grid's marks - the host's
// text, one a value - are written at the arrows' spacing, on the points an
// arrow leaves unlabelled, before any value a field writes: where a mark is,
// it takes the place of a wind's speed beside it; where none is, the wind
// keeps its speeds.
func TestAGridsMarksAreWrittenOnTheMap(t *testing.T) {
	m := rainMap(t)
	m.ShowStamp(false)
	rain := rainField()
	for i := range rain.Values {
		mark := ""
		if i%rain.Cols < rain.Cols/2 {
			mark = "7.5" // the west wet, the east dry
		}
		rain.Marks = append(rain.Marks, mark)
	}
	mustSet(t, m, tuimaps.Overlay{ID: "rain", Valid: noon, Keeps: 12 * time.Hour, Grid: &rain})
	g, from := westerly()
	mustSet(t, m, tuimaps.WindGrid("wind", g, from, tuimaps.MilesPerHour, noon))
	settle(t, m)
	body := colours.ReplaceAllString(momentFrame(t, m, noon), "")
	if strings.Count(body, "7.5") < 4 {
		t.Errorf("the grid's marks are not on the map:\n%s", body)
	}
	if !strings.Contains(body, "15") {
		t.Errorf("the marks took the arrows' speeds:\n%s", body)
	}
}

// TestAGridsMarksAreChecked is L-17.3: a mark a value, each at most eight
// cells wide, and marks only on a grid.
func TestAGridsMarksAreChecked(t *testing.T) {
	m := reportMap(t)
	for _, c := range []struct {
		name  string
		marks func(n int) []string
	}{
		{"one short", func(n int) []string { return make([]string, n-1) }},
		{"too wide", func(n int) []string { s := make([]string, n); s[0] = "123456789"; return s }},
	} {
		rain := rainField()
		rain.Marks = c.marks(len(rain.Values))
		_, err := m.Set(tuimaps.Overlay{ID: "rain", Valid: noon, Keeps: 12 * time.Hour, Grid: &rain})
		if !isKind(err, fault.SizeMismatch) {
			t.Errorf("%s: marks were not refused as a size mismatch: %v", c.name, err)
		}
	}
}

// TestTheRainLegendKeysTheColoursDrawn is L-17.1's defect, found building it:
// radar's class below its first floor is never drawn, and the legend keyed it
// in the first colour anyway - every class a colour off from the map, the
// heaviest in a temperature colour past the ramp's end. The class 40 dBZ
// falls in is keyed in the colour 40 dBZ is drawn in, and the lightest class
// is keyed as not drawn; the legend of a grid of rain is never faint.
func TestTheRainLegendKeysTheColoursDrawn(t *testing.T) {
	m := rainMap(t)
	rain := rainField()
	rain.Lines = true
	mustSet(t, m, tuimaps.Overlay{ID: "rain", Valid: noon, Keeps: 12 * time.Hour, Grid: &rain})
	settle(t, m)
	raw := momentFrame(t, m, noon)
	for _, e := range m.Legend() {
		if e.Preset != "radar" {
			continue
		}
		if e.Classes[0].Drawn {
			t.Errorf("the class under radar's first floor is keyed as drawn, in %v: nothing is drawn there", e.Classes[0].Colour)
		}
		for _, c := range e.Classes {
			if strings.HasPrefix(c.Label, "40 ") && groundsOf([]tuimaps.Class{c}, raw) == 0 {
				t.Errorf("40 dBZ is keyed %v, a colour the map does not draw it in", c.Colour)
			}
		}
		return
	}
	t.Fatal("the legend has no rain")
}

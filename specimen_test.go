package tuimaps_test

// specimen_test.go — v0.3.0 C-5: what the library draws for a whole-globe
// grid and for a host-type grid, at every colour depth on both grounds, kept
// as golden frames in testdata/specimens. Each change v0.3.0 makes to fields
// is seen against these: a frame that changes fails here until the specimens
// are written again, by intent, with -update-specimens.

import (
	"flag"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

var updateSpecimens = flag.Bool("update-specimens", false, "write testdata/specimens from what the library draws")

// specimenSize is a specimen frame's size: the whole globe at 2° in a
// terminal of ordinary width.
var specimenSize = tuimaps.Size{Cols: 120, Rows: 36}

// The grounds a specimen is drawn on: a dark terminal and a light one.
var specimenGrounds = []struct {
	name   string
	behind tuimaps.RGB
}{{"dark", tuimaps.RGB{R: 16, G: 16, B: 16}}, {"light", tuimaps.RGB{R: 250, G: 250, B: 245}}}

// The depths a specimen is drawn at.
var specimenDepths = []struct {
	name  string
	depth tuimaps.Depth
}{{"truecolor", tuimaps.Truecolor}, {"256", tuimaps.Colours256}, {"16", tuimaps.Colours16}}

// globeGrid is a whole-globe 2° grid, 180 x 91 points, west to east from
// 180°W and rows from the north: a smooth field, highest near the equator by
// day, as a MUF map is, at the values scale gives.
func globeGrid(scale float64) tuimaps.Grid {
	g := tuimaps.Grid{West: -180, South: -90, East: 180, North: 90, Cols: 180, Rows: 91}
	g.Values = make([]float64, 0, g.Cols*g.Rows)
	for r := range g.Rows {
		lat := 90 - 2*float64(r)
		for c := range g.Cols {
			lon := -180 + 2*float64(c)
			day := math.Cos((lon - 30) * math.Pi / 180)
			g.Values = append(g.Values, scale*(10+12*math.Cos(lat*math.Pi/180)*(1+0.6*day)+2*math.Sin(lon*math.Pi/45)))
		}
	}
	return g
}

// hostBreaks are a host's own classes for the globe grid: MHz, as watchpost
// gives them before the library has a MUF preset.
var hostBreaks = []float64{8, 12, 16, 20, 24, 28}

// The specimens: each an overlay over the whole world.
var specimens = []struct {
	name    string
	overlay func() tuimaps.Overlay
}{
	{"globe-temperature", func() tuimaps.Overlay {
		return tuimaps.TemperatureGrid("globe", globeGrid(1.4), tuimaps.Celsius, noon)
	}},
	{"globe-muf", func() tuimaps.Overlay {
		return tuimaps.MUFGrid("globe", globeGrid(1), noon)
	}},
	{"globe-fof2", func() tuimaps.Overlay {
		return tuimaps.FoF2Grid("globe", globeGrid(0.4), noon)
	}},
	{"host-type", func() tuimaps.Overlay {
		g := globeGrid(1)
		g.Type = tuimaps.Type{Unit: "MHz", Breaks: hostBreaks}
		return tuimaps.Overlay{ID: "host", Valid: noon, Keeps: 3600e9, Grid: &g}
	}},
	{"host-type-lines", func() tuimaps.Overlay {
		g := globeGrid(1)
		g.Type = tuimaps.Type{Unit: "MHz", Breaks: hostBreaks}
		g.Lines = true
		return tuimaps.Overlay{ID: "host", Valid: noon, Keeps: 3600e9, Grid: &g}
	}},
}

// specimenFrame is a specimen drawn: the whole world, the overlay set and
// settled, at a depth on a ground.
func specimenFrame(t *testing.T, o tuimaps.Overlay, depth tuimaps.Depth, ground tuimaps.RGB) string {
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
	return strings.Join(f.Lines, "\n") + "\n"
}

// TestTheSpecimensAreWhatTheLibraryDraws is C-5: every specimen, at every
// depth on both grounds, is the frame its golden file holds.
func TestTheSpecimensAreWhatTheLibraryDraws(t *testing.T) {
	dir := filepath.Join("testdata", "specimens")
	if *updateSpecimens {
		must(t, os.MkdirAll(dir, 0o755))
	}
	for _, s := range specimens {
		for _, d := range specimenDepths {
			for _, g := range specimenGrounds {
				name := s.name + "-" + d.name + "-" + g.name + ".ans"
				got := specimenFrame(t, s.overlay(), d.depth, g.behind)
				path := filepath.Join(dir, name)
				if *updateSpecimens {
					must(t, os.WriteFile(path, []byte(got), 0o644))
					continue
				}
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("%s: %v; write the specimens with -update-specimens", name, err)
				}
				if string(want) != got {
					t.Errorf("%s: the library draws a different frame; if the change is meant, write the specimens again with -update-specimens and say so in the changelog", name)
				}
			}
		}
	}
}

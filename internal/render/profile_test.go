package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// roads is a tile carrying one major road and one minor road, side by side.
func roads() *scene.Tile {
	major := scene.Feature{Kind: scene.GeomLine, Class: "motorway", FirstPart: 0, EndPart: 1}
	minor := scene.Feature{Kind: scene.GeomLine, Class: "secondary", FirstPart: 1, EndPart: 2}
	return &scene.Tile{Layers: []scene.Layer{{Name: "transportation", Extent: 4096,
		Features: []scene.Feature{major, minor},
		Coords:   []int16{0, 1024, 4095, 1024, 0, 3072, 4095, 3072},
		Parts:    []uint32{4, 8}}}}
}

// inked is how many dots of the frame's line work are drawn in a token.
func inked(p *Painter, v project.View, token colour.Token) int {
	n := 0
	for row := range v.Rows {
		for col := range v.Cols {
			glyph, ink := p.Lines().Cell(col, row)
			if ink != uint8(token) {
				continue
			}
			for mask := uint8(glyph - blank); mask != 0; mask &= mask - 1 {
				n++
			}
		}
	}
	return n
}

// painted draws one tile under a profile and answers with the painter.
func painted(t *testing.T, v project.View, tile *scene.Tile, profile style.Profile) *Painter {
	t.Helper()
	pt := painter(t, v)
	pt.SetProfile(profile)
	if err := pt.Tile(v, tile, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	return pt
}

// TestProfileThinsUnderOverlay is plan task 09.8 in the frame (FR-19): what
// the profile says the basemap gives up is what the canvas shows.
func TestProfileThinsUnderOverlay(t *testing.T) {
	v := worldView()
	full := painted(t, v, roads(), style.NewProfile(style.Bare, v.Cols, v.Rows, 0))
	under := painted(t, v, roads(), style.NewProfile(style.Covered, v.Cols, v.Rows, 0))
	if inked(full, v, colour.RoadMinor) == 0 {
		t.Fatal("no minor road was drawn with nothing over the basemap")
	}
	if got := inked(under, v, colour.RoadMinor); got != 0 {
		t.Errorf("%d dots of minor road under a field; minor roads are the first thing given up", got)
	}
	if inked(under, v, colour.RoadMajor) == 0 {
		t.Error("the major road went under a field; only its weight should have")
	}
	// The built-in style draws every line one dot thick, so thinning shows on
	// a style that asks for more: a road three dots thick is drawn as one.
	own, err := style.Parse([]byte(`{"layers":[{"id":"wide","type":"line","source-layer":"transportation",` +
		`"paint":{"line-color":"#ff0000","line-width":3}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	wide, thin := painter(t, v), painter(t, v)
	thin.SetProfile(style.NewProfile(style.Covered, v.Cols, v.Rows, 0))
	for _, pt := range []*Painter{wide, thin} {
		if err := pt.Tile(v, roads(), here, own); err != nil {
			t.Fatal(err)
		}
	}
	thickDots, thinDots := lit(wide), lit(thin)
	if thinDots == 0 || thinDots >= thickDots {
		t.Errorf("a road three dots thick draws %d dots under a field and %d with nothing over it", thinDots, thickDots)
	}
}

// TestLayerToggleInFrame is plan task 09.9 in the frame (FR-36).
func TestLayerToggleInFrame(t *testing.T) {
	v := worldView()
	on := painted(t, v, roads(), style.NewProfile(style.Bare, v.Cols, v.Rows, 0))
	off := painted(t, v, roads(), style.NewProfile(style.Bare, v.Cols, v.Rows, style.Off(style.RoadLayer)))
	if inked(on, v, colour.RoadMajor) == 0 || inked(on, v, colour.RoadMinor) == 0 {
		t.Fatal("the roads were not drawn with their layer on")
	}
	if inked(off, v, colour.RoadMajor) != 0 || inked(off, v, colour.RoadMinor) != 0 {
		t.Error("a road was drawn with the road layer switched off")
	}
}

// TestProfileChosenForTheFrame: the renderer chooses the profile itself, from
// what is drawn on top and the host's switches. Under a field it places fewer
// names; with the water layer switched off there is no shore at all.
func TestProfileChosenForTheFrame(t *testing.T) {
	v, err := project.FitWorld(149, 38)
	if err != nil {
		t.Fatal(err)
	}
	v.Zoom, v.Centre = 3, project.LonLat{Lon: -84, Lat: 26}
	bare := input(t, v)
	covered := input(t, v)
	covered.Fields = []scene.Field{{West: -100, South: 20, East: -70, North: 35, Cols: 4, Rows: 4,
		Classes: []int8{0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3}, Preset: uint8(colour.Temperature), ClassCount: 17}}
	covered.OverlaysVersion = 1
	plain, fewer := named(render(t, bare)), named(render(t, covered))
	if fewer >= plain {
		t.Errorf("%d names under a field and %d with nothing over it; a covered map places no more than %d", fewer, plain, style.NewProfile(style.Covered, v.Cols, v.Rows, 0).Labels())
	}
	noWater := input(t, v)
	noWater.Layers = style.Off(style.WaterLayer)
	if dots(render(t, noWater)) >= dots(render(t, bare)) {
		t.Error("switching off the water layer left the frame as it was")
	}
}

// lit is how many dots a painter's line work has drawn.
func lit(p *Painter) int {
	w, h := p.Lines().Dots()
	n := 0
	for y := range h {
		for x := range w {
			if p.Lines().Lit(x, y) {
				n++
			}
		}
	}
	return n
}

// dots is how many braille dots a frame carries: a rough measure of how much
// of the basemap it draws.
func dots(f Frame) int {
	n := 0
	for _, line := range f.Lines {
		for _, r := range plain(line) {
			if r >= blank && r <= blank+0xFF {
				for mask := uint8(r - blank); mask != 0; mask &= mask - 1 {
					n++
				}
			}
		}
	}
	return n
}

// named is how many names a frame carries, counted as runs of letters; the
// credit line is left out, being furniture and not a name.
func named(f Frame) int {
	n := 0
	for i, line := range f.Lines {
		if i == len(f.Lines)-1 {
			continue // the credit and the scale mark live on the last row
		}
		inWord := false
		for _, r := range plain(line) {
			letter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			if letter && !inWord {
				n++
			}
			inWord = letter
		}
	}
	return n
}

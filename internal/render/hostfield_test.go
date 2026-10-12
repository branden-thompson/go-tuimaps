package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// hostBands is a host type of n classes over the gulf, its class rising from
// west to east, over the sea as well.
func hostBands(n int) scene.Field {
	f := scene.Field{West: -110, South: 15, East: -70, North: 45, Cols: 40, Rows: 30, Classes: make([]int8, 40*30), ClassCount: n, OverWater: true}
	for row := range 30 {
		for col := range 40 {
			f.Classes[row*40+col] = int8(col * n / 40)
		}
	}
	return f
}

// reach is a two-class host field over the gulf: 1 inside a box, 0 around
// it, as a frequency's reach is (watchpost D-127).
func reach() scene.Field {
	f := scene.Field{West: -110, South: 15, East: -70, North: 45, Cols: 40, Rows: 30, Classes: make([]int8, 40*30), ClassCount: 2, OverWater: true}
	for row := 8; row < 22; row++ {
		for col := 12; col < 28; col++ {
			f.Classes[row*40+col] = 1
		}
	}
	return f
}

// TestAHostFieldIsFilledInItsSpread is L-2.2 at the cell: a host type's
// bands are drawn, each in its class's colour of the low, middle and high
// spread, or in the host's own colours where it gave them.
func TestAHostFieldIsFilledInItsSpread(t *testing.T) {
	v := gulfView()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Fields: []scene.Field{hostBands(3)}, OverlaysVersion: 1, Depth: colour.Truecolor}
	r, _ := drawn(t, in)
	want := colour.HostRamp(nil, 3, in.Palette, in.Ground.Kind(in.Palette), colour.Truecolor)
	seen := map[colour.RGB]bool{}
	for _, c := range r.grid.cells {
		if c.under == 0 {
			continue
		}
		rgb, ok := r.painter.Colour(c.under, in.Palette, in.Ground.Kind(in.Palette), colour.Truecolor)
		if !ok {
			t.Fatalf("a cell's ink %d has no colour", c.under)
		}
		seen[rgb] = true
	}
	for _, c := range want {
		if !seen[c] {
			t.Errorf("no cell is drawn in %v; the spread is %v and the cells %v", c, want, seen)
		}
	}
	own := hostBands(3)
	own.Colours = [][3]uint8{{10, 20, 30}, {100, 110, 120}, {200, 210, 220}}
	in.Fields, in.OverlaysVersion = []scene.Field{own}, 2
	r, _ = drawn(t, in)
	seen = map[colour.RGB]bool{}
	for _, c := range r.grid.cells {
		if rgb, ok := r.painter.Colour(c.under, in.Palette, in.Ground.Kind(in.Palette), colour.Truecolor); ok {
			seen[rgb] = true
		}
	}
	for _, c := range own.Colours {
		if !seen[colour.RGB{R: c[0], G: c[1], B: c[2]}] {
			t.Errorf("the host's colour %v is drawn in no cell: %v", c, seen)
		}
	}
}

// TestTheReachIsMarkedWithoutColour is L-2.6 (watchpost D-127): at every
// depth, the edge between a two-class host field's classes is a broken
// line - about half its dots - with no label, never a contour; alone, at a
// depth with a ramp, its classes are filled; over a filled field, its edge
// alone, the field beneath kept (A-10).
func TestTheReachIsMarkedWithoutColour(t *testing.T) {
	v := gulfView()
	for _, d := range []colour.Depth{colour.Truecolor, colour.Colours256, colour.Colours16, colour.NoColour} {
		for _, ground := range []colour.RGB{{R: 16, G: 16, B: 16}, {R: 250, G: 250, B: 245}} {
			choice, _ := colour.Declared(ground)
			in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), OverlaysVersion: 1, Depth: d, Ground: choice}
			bare, _ := drawn(t, in) // the basemap's own line work, which shares the canvas
			in.Fields, in.OverlaysVersion = []scene.Field{reach()}, 2
			r, _ := drawn(t, in)
			f := &in.Fields[0]
			edge, lit := 0, 0
			w, h := r.painter.lines.Dots()
			for y := range h - 1 {
				for x := range w - 1 {
					here := fieldClass(f, r.lons[x], r.lats[y])
					east, south := fieldClass(f, r.lons[x+1], r.lats[y]), fieldClass(f, r.lons[x], r.lats[y+1])
					onEdge := here >= 0 && ((east >= 0 && east != here) || (south >= 0 && south != here))
					on := r.painter.lines.Lit(x, y) && !bare.painter.lines.Lit(x, y)
					switch {
					case onEdge:
						edge++
						if on {
							lit++
						}
					case on && here >= 0:
						t.Fatalf("depth %v on %v: a dot off the reach's edge is lit at %d,%d", d, ground, x, y)
					}
				}
			}
			if edge == 0 || lit < edge/4 || lit > 3*edge/4 {
				t.Errorf("depth %v on %v: %d of the edge's %d dots lit; a broken line lights about half", d, ground, lit, edge)
			}
			if len(r.painter.bandLabels) != 0 {
				t.Errorf("depth %v on %v: the reach's edge carries %d labels; it is no contour", d, ground, len(r.painter.bandLabels))
			}
			filled := 0
			for _, c := range r.grid.cells {
				if c.under != 0 {
					filled++
				}
			}
			if rampless(d) == (filled > 0) {
				t.Errorf("depth %v on %v: %d cells filled; alone, a reach is filled where there is a ramp, and not where there is none", d, ground, filled)
			}
		}
	}
	under := bands(40, 30)
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Fields: []scene.Field{under, reach()}, OverlaysVersion: 1, Depth: colour.Truecolor}
	r, _ := drawn(t, in)
	for i, c := range r.grid.cells {
		if c.under != 0 && (c.under < uint8(colour.Temperature1) || c.under > uint8(colour.Temperature17)) {
			t.Fatalf("cell %d is drawn in ink %d over the temperature field; a reach over a filled field is its edge alone", i, c.under)
		}
	}
}

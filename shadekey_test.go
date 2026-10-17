package tuimaps_test

import (
	"context"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestEveryClassHasAShadeKey is L-7.1 (OW-24, D-35): at sixteen colours and
// none, each of an image's classes the map shades carries the glyph it is
// drawn in, heavier for heavier classes, so a reader without colour can map
// ░▒▓ to a range; where colour is drawn, and for a field drawn as contours,
// a class carries none.
func TestEveryClassHasAShadeKey(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	if _, err := m.Set(radarLoop(t, "radar", 1, 8, 8)); err != nil {
		t.Fatal(err)
	}
	grid := tuimaps.Grid{West: -100, South: 20, East: -70, North: 35, Cols: 2, Rows: 2, Values: []float64{4, 9, 15, 30}}
	if _, err := m.Set(tuimaps.MUFGrid("muf", grid, noon)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	rank := map[string]int{"": 0, "░": 1, "▒": 2, "▓": 3}
	for _, depth := range []tuimaps.Depth{tuimaps.Colours16, tuimaps.NoColour, tuimaps.Colours256, tuimaps.Truecolor} {
		m.ColourDepth(depth)
		frameAtTime(t, m, cols, rows, noon)
		rampless := depth == tuimaps.Colours16 || depth == tuimaps.NoColour
		for _, e := range m.Legend() {
			last, keyed := 0, 0
			for i, c := range e.Classes {
				r, ok := rank[c.Shade]
				if !ok {
					t.Fatalf("at %v %s's class %d carries %q, no shade", depth, e.ID, i, c.Shade)
				}
				if c.Shade == "" {
					continue
				}
				keyed++
				if r < last {
					t.Errorf("at %v %s's class %d is lighter (%q) than the one before", depth, e.ID, i, c.Shade)
				}
				last = r
			}
			switch want := e.ID == "radar" && rampless; {
			case want && keyed < len(e.Classes)-1: // radar's first class, below its first floor, is never drawn
				t.Errorf("at %v radar keys %d of its %d classes", depth, keyed, len(e.Classes))
			case !want && keyed > 0:
				t.Errorf("at %v %s keys %d shades it does not draw", depth, e.ID, keyed)
			}
		}
	}
}

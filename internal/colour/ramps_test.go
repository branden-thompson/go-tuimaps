package colour

import (
	"slices"
	"testing"
)

// TestTheKeptRampsAreTheBuiltOnes: every preset's ramp, for each ground and
// depth, is the one its function builds; a caller's copy is its own; and a
// ramp token resolves without allocating, as a frame resolves one a cell.
func TestTheKeptRampsAreTheBuiltOnes(t *testing.T) {
	for p := Temperature; p <= FoF2; p++ {
		for _, g := range []GroundKind{Dark, Light} {
			for _, d := range []Depth{Truecolor, Colours256, Colours16, NoColour} {
				got, ok := Ramp(p, g, d)
				want, wantOK := buildRamp(p, g, d)
				if d != Truecolor && d != Colours256 {
					want, wantOK = nil, false
				}
				if ok != wantOK || !slices.Equal(got, want) {
					t.Errorf("preset %d, %v, %v: %v (%v); built %v (%v)", p, g, d, got, ok, want, wantOK)
				}
			}
		}
	}
	mine, _ := Ramp(MUF, Dark, Truecolor)
	mine[0] = RGB{1, 2, 3}
	if again, _ := Ramp(MUF, Dark, Truecolor); again[0] == mine[0] {
		t.Error("a caller's change to its ramp reached the library's")
	}
	if _, ok := Ramp(FoF2+1, Dark, Truecolor); ok {
		t.Error("a value that is no preset has a ramp")
	}
	var pal Palette
	if n := testing.AllocsPerRun(100, func() { _, _ = pal.ResolveAt(MUF1+2, Dark, Truecolor) }); n != 0 {
		t.Errorf("a ramp token resolves with %v allocations; want none", n)
	}
}

package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// TestParityP31_Simplify is plan task 09.30: upstream's line simplification,
// at upstream's tolerance, and off by default as upstream leaves it.
func TestParityP31_Simplify(t *testing.T) {
	if SimplifyTolerance != 0.5 {
		t.Errorf("the tolerance is %v, want upstream's 0.5", SimplifyTolerance)
	}
	// A line of many points that never leaves the straight: two points left.
	straight := make([]Point, 0, 100)
	for i := range 100 {
		straight = append(straight, Point{X: i, Y: 40})
	}
	if got := simplify(straight, nil, SimplifyTolerance); len(got) != 2 {
		t.Errorf("a straight line of 100 points simplified to %d, want its two ends", len(got))
	}
	// A point a whole dot off the straight is kept; one a quarter off is not.
	for _, c := range []struct {
		off  int
		want int
	}{{0, 2}, {1, 3}} {
		line := []Point{{X: 0, Y: 40}, {X: 20, Y: 40 + c.off}, {X: 40, Y: 40}}
		if got := simplify(line, nil, SimplifyTolerance); len(got) != c.want {
			t.Errorf("a bend of %d dots simplified to %d points, want %d", c.off, len(got), c.want)
		}
	}
	// Ends are never dropped, and a ring stays closed.
	ring := []Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}, {X: 0, Y: 20}, {X: 0, Y: 0}}
	got := simplify(ring, nil, SimplifyTolerance)
	if got[0] != ring[0] || got[len(got)-1] != ring[len(ring)-1] {
		t.Errorf("simplifying moved a ring's ends: %v", got)
	}
	if len(got) != 5 {
		t.Errorf("a square of six points simplified to %d, want five: the point in the middle of a side goes", len(got))
	}
}

// TestSimplifyOffByDefault: simplification is a choice the host makes, and
// it is off unless the host makes it (P-31). At half a dot it is a saving in
// work, not a change to the picture: every point it drops lies on the line
// its neighbours already draw.
func TestSimplifyOffByDefault(t *testing.T) {
	v := worldView()
	// A line of 200 points: four long straights, and a step between each.
	coords := make([]int16, 0, 400)
	for i := range 200 {
		coords = append(coords, int16(i*20), int16(2048+(i/50)*64))
	}
	tile := tileWith("waterway", scene.Feature{Kind: scene.GeomLine, Class: "river"}, coords...)
	asItCame, simplified := painter(t, v), painter(t, v)
	simplified.SetSimplify(true)
	for _, p := range []*Painter{asItCame, simplified} {
		if err := p.Tile(v, tile, here, style.BuiltIn()); err != nil {
			t.Fatal(err)
		}
	}
	if lit(asItCame) == 0 {
		t.Fatal("the line was not drawn at all")
	}
	if len(simplified.thinner) == 0 || len(simplified.thinner) >= 200 {
		t.Errorf("simplified, the line is %d points of 200 drawn; the ones on the straight go", len(simplified.thinner))
	}
	if len(asItCame.thinner) != 0 {
		t.Error("a painter nobody asked simplified the line anyway")
	}
	// A frame draws the same picture either way: only the work changes.
	in := Input{View: v, Style: style.BuiltIn(), Tiles: []Drawn{{Tile: tile, At: here, Exact: true}}}
	off := render(t, in)
	in.Simplify = true
	on := render(t, in)
	if dots(off) != dots(on) {
		t.Errorf("%d dots with simplification off and %d with it on; at half a dot the picture does not change", dots(off), dots(on))
	}
	for i := range off.Lines {
		if plain(off.Lines[i]) != plain(on.Lines[i]) {
			t.Fatalf("row %d differs:\n%s\n%s", i, plain(off.Lines[i]), plain(on.Lines[i]))
		}
	}
}

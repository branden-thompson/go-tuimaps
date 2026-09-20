package render

import (
	"math"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// indexOf builds the run index of a ring the way the overlay store does: one
// box for each run of 64 vertices, a run ending on the vertex the next
// begins with, from the extremes of longitude and latitude in the run.
func indexOf(t *testing.T, rings [][]project.LonLat) []scene.Run {
	t.Helper()
	var index []scene.Run
	for _, ring := range rings {
		for start := 0; start < len(ring); start += RunLength {
			end := min(start+RunLength, len(ring)-1)
			west, east, south, north := 180.0, -180.0, 90.0, -90.0
			for _, p := range ring[start : end+1] {
				west, east = math.Min(west, p.Lon), math.Max(east, p.Lon)
				south, north = math.Min(south, p.Lat), math.Max(north, p.Lat)
			}
			x0, y0, err := project.ToTile(project.LonLat{Lon: west, Lat: north}, 0)
			if err != nil {
				t.Fatal(err)
			}
			x1, y1, err := project.ToTile(project.LonLat{Lon: east, Lat: south}, 0)
			if err != nil {
				t.Fatal(err)
			}
			index = append(index, scene.Run{MinX: whole(x0), MinY: whole(y0), MaxX: whole(x1), MaxY: whole(y1)})
		}
	}
	return index
}

// coastline is a long line of places running west to east across the world,
// at a latitude that wanders a little.
func coastline(n int) []project.LonLat {
	ring := make([]project.LonLat, 0, n)
	for i := range n {
		lon := -179 + 358*float64(i)/float64(n-1)
		ring = append(ring, project.LonLat{Lon: lon, Lat: 20 + 8*math.Sin(float64(i)/40)})
	}
	return ring
}

// TestDrawFromRunIndex is plan task 09.31 (FR-11, D-92): an overlay whose
// geometry is the host's own is drawn by testing one box for each run and
// reading only the runs whose box meets the view. The work it costs is the
// bound the constants state, and no more.
func TestDrawFromRunIndex(t *testing.T) {
	ring := coastline(20_000)
	rings := [][]project.LonLat{ring}
	borrowed := Borrowed{Kind: scene.ShapeLine, Rings: rings, Index: indexOf(t, rings), Role: uint8(colour.Track)}
	runs := (len(ring) + RunLength - 1) / RunLength
	if len(borrowed.Index) != runs {
		t.Fatalf("%d boxes for %d vertices; one for each run of %d", len(borrowed.Index), len(ring), RunLength)
	}
	// A view of one small part of the world.
	v, err := project.WholeWorld(149, 38)
	if err != nil {
		t.Fatal(err)
	}
	v.Zoom, v.Centre = 6, project.LonLat{Lon: -84, Lat: 26}
	p := painter(t, v)
	if err := p.Borrow(v, borrowed); err != nil {
		t.Fatal(err)
	}
	got := p.Reads()
	if got.Boxes != runs {
		t.Errorf("%d box tests, want one for each of the %d runs", got.Boxes, runs)
	}
	if got.Vertices == 0 {
		t.Fatal("nothing of the line was read; the view crosses it")
	}
	if got.Vertices > len(ring)/8 {
		t.Errorf("%d of %d vertices read for a view of a small part of the world; only the runs the view meets are read", got.Vertices, len(ring))
	}
	if lit(p) == 0 {
		t.Error("the line was read and nothing was drawn")
	}
	// What is drawn is what drawing the whole of it draws: the index culls,
	// it does not change the picture.
	whole := painter(t, v)
	if err := whole.Borrow(v, Borrowed{Kind: scene.ShapeLine, Rings: rings, Role: uint8(colour.Track)}); err != nil {
		t.Fatal(err)
	}
	// With no index every run is read, each ending on the vertex the next
	// begins with, so a vertex a run is read twice and no box is tested.
	if got := whole.Reads(); got.Boxes != 0 || got.Vertices < len(ring) || got.Vertices > len(ring)+runs {
		t.Errorf("with no index: %+v; the whole of it is read, %d vertices in %d runs", got, len(ring), runs)
	}
	if lit(p) != lit(whole) {
		t.Errorf("%d dots drawn through the index and %d without it", lit(p), lit(whole))
	}
}

// TestBorrowedSplitAtAntimeridian: a line that steps across the world's edge
// is drawn as two lines, not as one straight across the map (L-17 f).
func TestBorrowedSplitAtAntimeridian(t *testing.T) {
	v, err := project.WholeWorld(149, 38)
	if err != nil {
		t.Fatal(err)
	}
	v.Centre = project.LonLat{Lon: 0, Lat: 0} // the whole world, as FitWorld gives it
	rings := [][]project.LonLat{{{Lon: 170, Lat: 10}, {Lon: 179, Lat: 12}, {Lon: -179, Lat: 12}, {Lon: -170, Lat: 10}}}
	p := painter(t, v)
	if err := p.Borrow(v, Borrowed{Kind: scene.ShapeLine, Rings: rings, Role: uint8(colour.Track)}); err != nil {
		t.Fatal(err)
	}
	middle := 0
	w, h := p.Lines().Dots()
	for x := w/2 - 20; x < w/2+20; x++ {
		for y := range h {
			if p.Lines().Lit(x, y) {
				middle++
			}
		}
	}
	if middle != 0 {
		t.Errorf("%d dots drawn across the middle of the map; a step over the world's edge is not a line across it", middle)
	}
	if lit(p) == 0 {
		t.Error("neither end of the line was drawn")
	}
}

// TestBorrowedDrawnInTheFrame: the renderer draws what it is lent along with
// everything else, and a frame with no index draws the same picture.
func TestBorrowedDrawnInTheFrame(t *testing.T) {
	v, in := gulf(t)
	rings := [][]project.LonLat{coastline(4_000)}
	in.Borrowed = []Borrowed{{Kind: scene.ShapeLine, Rings: rings, Index: indexOf(t, rings), Role: uint8(colour.Track)}}
	in.OverlaysVersion = 1
	bare := input(t, v)
	if dots(render(t, in)) <= dots(render(t, bare)) {
		t.Error("the frame with a borrowed overlay draws no more than the frame without it")
	}
}

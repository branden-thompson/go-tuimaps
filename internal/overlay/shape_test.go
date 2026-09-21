package overlay

import (
	"math"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// circle is a ring of n vertices round a centre, with a ripple so that no
// three are in line; it repeats its first vertex to close.
func circle(centre project.LonLat, radius float64, n int) []project.LonLat {
	ring := make([]project.LonLat, 0, n+1)
	for i := range n {
		a := 2 * math.Pi * float64(i) / float64(n)
		r := radius * (1 + 0.02*math.Sin(40*a))
		ring = append(ring, project.LonLat{Lon: centre.Lon + r*math.Cos(a), Lat: centre.Lat + r*math.Sin(a)})
	}
	return append(ring, ring[0])
}

// farthest is the greatest distance from any vertex of a ring to the nearest
// edge of another, in the world's own units.
func farthest(from, to []Vertex) float64 {
	worst := 0.0
	for _, p := range from {
		best := math.Inf(1)
		for i := 0; i+1 < len(to); i++ {
			best = math.Min(best, distance(p, to[i], to[i+1]))
		}
		worst = math.Max(worst, best)
	}
	return worst
}

// TestZoomBuckets is plan task 10.8: a fractional zoom maps to one bucket.
func TestZoomBuckets(t *testing.T) {
	for zoom, want := range map[float64]int{0: 0, 0.99: 0, 1: 1, 6.4: 6, 11.999: 11, 12: 12, -0.5: -1, -8: -8, 18: 18, 40: 18, -40: -8} {
		if got := Bucket(zoom); got != want {
			t.Errorf("zoom %v is bucket %d, want %d", zoom, got, want)
		}
	}
	if Bucket(math.NaN()) != 0 {
		t.Error("a zoom that is no number has a bucket other than 0")
	}
	// Half a braille dot at the finest zoom the bucket serves: zoom b+1.
	if got, want := Tolerance(6), 0.5/(256*math.Exp2(7)); math.Abs(got-want) > 1e-18 {
		t.Errorf("bucket 6's tolerance is %v of the world's side, want %v", got, want)
	}
	if Tolerance(7) >= Tolerance(6) {
		t.Error("a deeper bucket has a coarser tolerance")
	}
	// On a miss the nearest prepared bucket is drawn.
	for _, c := range []struct {
		want     int
		prepared []int
		nearest  int
		ok       bool
	}{{6, []int{6}, 6, true}, {6, []int{3, 9}, 9, true}, {6, []int{2, 9}, 9, true}, {6, []int{4, 9}, 4, true}, {6, []int{4, 7}, 7, true}, {6, []int{5, 7}, 7, true}, {6, nil, 0, false}} {
		got, ok := Nearest(c.want, c.prepared)
		if got != c.nearest || ok != c.ok {
			t.Errorf("want bucket %d of %v: %d, %v; a tie goes to the finer", c.want, c.prepared, got, ok)
		}
	}
}

// TestSimplifyIterative is plan task 10.6: a 14,001-vertex ring, simplified
// without a function calling itself, with every vertex it dropped within the
// tolerance of what is left, and the ring still closed.
func TestSimplifyIterative(t *testing.T) {
	ring, err := Project(circle(project.LonLat{Lon: -95, Lat: 38}, 3, 14000))
	if err != nil || len(ring) != 14001 {
		t.Fatal(len(ring), err)
	}
	for _, bucket := range []int{4, 8, 12} {
		tol := Tolerance(bucket)
		kept := Simplify(ring, tol)
		if len(kept) < 4 || len(kept) >= len(ring) || kept[0] != kept[len(kept)-1] {
			t.Fatalf("bucket %d: %d vertices kept of %d, closed=%v", bucket, len(kept), len(ring), kept[0] == kept[len(kept)-1])
		}
		if worst := farthest(ring, kept); worst > tol*1.0000001 {
			t.Errorf("bucket %d: a dropped vertex is %v from the simplified ring, over the tolerance %v", bucket, worst, tol)
		}
		t.Logf("bucket %d: %d of %d vertices kept", bucket, len(kept), len(ring))
	}
	// A shape that would take a recursive version as deep as it is long: a
	// comb whose teeth grow, so that every split is at the far end.
	comb := make([]Vertex, 0, 200001)
	for i := range 200000 {
		comb = append(comb, Vertex{X: uint32(1000 + i), Y: uint32(1000 + (i%2)*(i+2))})
	}
	comb = append(comb, comb[0])
	// It would cost a plain version the square of its length; the work is
	// bounded, and what the budget does not reach is kept as it is, which is
	// never coarser than the tolerance.
	if kept := Simplify(comb, 0.25/math.Exp2(32)); len(kept) < 100000 {
		t.Errorf("the comb kept %d of 200,001 vertices; its teeth are all larger than the tolerance", len(kept))
	}
	if Simplify(nil, 1) != nil || len(Simplify(ring[:2], 1)) != 2 {
		t.Error("a ring of fewer than three vertices is returned as it is")
	}
}

// TestSubDotRingsDropped is plan task 10.7: three hand-made rings.
func TestSubDotRingsDropped(t *testing.T) {
	tol := Tolerance(6)
	dot := 2 * tol // a dot at zoom 7, in the world's units
	ring := func(size float64) []Vertex {
		at := func(x, y float64) Vertex { return Vertex{X: fixed(0.5 + x*size), Y: fixed(0.5 + y*size)} }
		return []Vertex{at(0, 0), at(1, 0), at(1, 1), at(0, 1), at(0, 0)}
	}
	for _, c := range []struct {
		name string
		size float64
		kept bool
	}{{"a square a tenth of a dot across", dot / 10, false}, {"a square a third of a dot across", dot / 3, false}, {"a square five dots across", 5 * dot, true}} {
		_, kept := Keep(Simplify(ring(c.size), tol))
		if kept != c.kept {
			t.Errorf("%s: kept=%v", c.name, kept)
		}
	}
	sliver := []Vertex{{X: fixed(0.5), Y: fixed(0.5)}, {X: fixed(0.5 + 40*dot), Y: fixed(0.5)}, {X: fixed(0.5 + 20*dot), Y: fixed(0.5 + dot/8)}, {X: fixed(0.5), Y: fixed(0.5)}}
	if _, kept := Keep(Simplify(sliver, tol)); kept {
		t.Error("a sliver forty dots long and an eighth of a dot wide simplifies to a line, and a line is not a ring")
	}
}

// TestSplitAtAntimeridian is plan task 10.9.
func TestSplitAtAntimeridian(t *testing.T) {
	// A box from 170 E to 170 W, across the line.
	across := []project.LonLat{{Lon: 170, Lat: -10}, {Lon: -170, Lat: -10}, {Lon: -170, Lat: 10}, {Lon: 170, Lat: 10}, {Lon: 170, Lat: -10}}
	parts, err := Split(across)
	if err != nil || len(parts) != 2 {
		t.Fatalf("%d parts, %v; a ring across the line becomes one either side", len(parts), err)
	}
	for _, part := range parts {
		west, east := 181.0, -181.0
		for _, p := range part {
			if p.Lon < -180 || p.Lon > 180 {
				t.Errorf("a vertex at longitude %v", p.Lon)
			}
			west, east = math.Min(west, p.Lon), math.Max(east, p.Lon)
		}
		if east-west > 10.0001 || (east != 180 && west != -180) {
			t.Errorf("a part runs from %v to %v; each is ten degrees wide and ends at the line", west, east)
		}
		if part[0] != part[len(part)-1] {
			t.Error("a part is not closed")
		}
	}
	// A ring that does not cross is returned as it is.
	inside := []project.LonLat{{Lon: 10, Lat: 0}, {Lon: 20, Lat: 0}, {Lon: 20, Lat: 10}, {Lon: 10, Lat: 0}}
	if parts, err := Split(inside); err != nil || len(parts) != 1 || len(parts[0]) != 4 {
		t.Errorf("%v, %v", parts, err)
	}
	// A ring that reaches the line and does not cross it is one part.
	touching := []project.LonLat{{Lon: 170, Lat: 0}, {Lon: 180, Lat: 0}, {Lon: 180, Lat: 10}, {Lon: 170, Lat: 0}}
	if parts, _ := Split(touching); len(parts) != 1 {
		t.Errorf("%d parts for a ring that only touches the line", len(parts))
	}
	if _, err := Split([]project.LonLat{{Lon: math.NaN(), Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 2, Lat: 0}}); err == nil {
		t.Error("a coordinate that is no number must be an error")
	}
}

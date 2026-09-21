package describe

import (
	"math"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// TestMain holds every test in this package to the loopback rule: nothing
// here reaches a network, and the rule is enforced rather than promised.
func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// box is a square ring, anticlockwise from its south-west corner.
func box(west, south, east, north float64) []project.LonLat {
	return []project.LonLat{{Lon: west, Lat: south}, {Lon: east, Lat: south}, {Lon: east, Lat: north}, {Lon: west, Lat: north}}
}

// TestInsideOutsideUnsimplified is plan task 11.3 (FR-11): the answer comes
// from the host's own rings, by the even-odd rule the picture is filled
// with, so a point in a hole is outside.
func TestInsideOutsideUnsimplified(t *testing.T) {
	outer := box(-10, -10, 10, 10)
	hole := box(-2, -2, 2, 2)
	cases := []struct {
		name string
		at   project.LonLat
		want Where
	}{
		{"the middle of a plain square", project.LonLat{Lon: 0, Lat: 0}, Inside},
		{"well outside it", project.LonLat{Lon: 20, Lat: 0}, Outside},
		{"just inside the west edge", project.LonLat{Lon: -9.999, Lat: 0}, Inside},
		{"just outside the west edge", project.LonLat{Lon: -10.001, Lat: 0}, Outside},
		{"north of it", project.LonLat{Lon: 0, Lat: 10.001}, Outside},
	}
	for _, c := range cases {
		if got := InArea(c.at, [][]project.LonLat{outer}); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	// With a hole, the middle is outside and the ring between them inside.
	rings := [][]project.LonLat{outer, hole}
	if got := InArea(project.LonLat{Lon: 0, Lat: 0}, rings); got != Outside {
		t.Errorf("the middle of a square with a hole in it is %v", got)
	}
	if got := InArea(project.LonLat{Lon: 5, Lat: 5}, rings); got != Inside {
		t.Errorf("between the hole and the edge is %v", got)
	}
	// A long thin shape, which a simplified copy would round off: the answer
	// is the host's geometry, not the picture's.
	sliver := []project.LonLat{{Lon: 0, Lat: 0}, {Lon: 0.0009, Lat: 0}, {Lon: 0.0009, Lat: 4}, {Lon: 0, Lat: 4}}
	if got := InArea(project.LonLat{Lon: 0.0004, Lat: 2}, [][]project.LonLat{sliver}); got != Inside {
		t.Errorf("a place inside a sliver a hundred metres wide is %v", got)
	}
}

// TestNearestEdgeDistanceAndBearing is plan task 11.4: how far the nearest
// edge is and which way, in kilometres and degrees from north.
func TestNearestEdgeDistanceAndBearing(t *testing.T) {
	area := [][]project.LonLat{box(-10, -10, 10, 10)}
	// A place a little east of the eastern edge: the nearest edge is west.
	edge, ok := NearestEdge(project.LonLat{Lon: 11, Lat: 0}, area)
	if !ok {
		t.Fatal("a square has no nearest edge")
	}
	if math.Abs(edge.Km-111.3) > 3 {
		t.Errorf("one degree of longitude at the equator is %v km, want about 111", edge.Km)
	}
	if math.Abs(edge.Bearing-270) > 1 {
		t.Errorf("the edge is at bearing %v, want due west", edge.Bearing)
	}
	// A place inside has a nearest edge too: the answer says how far it is
	// from getting out, which is what a warning area's edge means.
	inside, ok := NearestEdge(project.LonLat{Lon: 9, Lat: 0}, area)
	if !ok || math.Abs(inside.Bearing-90) > 1 {
		t.Errorf("from inside, the nearest edge is at %v, want due east", inside.Bearing)
	}
	// A corner is nearest when the place is diagonally out.
	corner, _ := NearestEdge(project.LonLat{Lon: 12, Lat: 12}, area)
	if math.Abs(corner.At.Lon-10) > 0.001 || math.Abs(corner.At.Lat-10) > 0.001 {
		t.Errorf("the nearest point is %+v, want the corner at 10,10", corner.At)
	}
	if math.Abs(corner.Bearing-225) > 2 {
		t.Errorf("the corner is at bearing %v, want about south-west", corner.Bearing)
	}
	// Nothing to measure against is said plainly rather than guessed.
	if _, ok := NearestEdge(project.LonLat{}, nil); ok {
		t.Error("an area with no rings has a nearest edge")
	}
}

// TestSeamIsNeverAnEdge is the other half of 11.4: longitude is circular.
// A place just west of the seam and an area just east of it are close
// neighbours, not a world apart, and a ring that crosses the seam is
// followed round it.
func TestSeamIsNeverAnEdge(t *testing.T) {
	// An area from 179 east to -179, which crosses the seam.
	over := [][]project.LonLat{{{Lon: 179, Lat: -1}, {Lon: -179, Lat: -1}, {Lon: -179, Lat: 1}, {Lon: 179, Lat: 1}}}
	if got := InArea(project.LonLat{Lon: 180, Lat: 0}, over); got != Inside {
		t.Errorf("the middle of an area over the seam is %v", got)
	}
	if got := InArea(project.LonLat{Lon: -179.5, Lat: 0}, over); got != Inside {
		t.Errorf("just east of the seam, inside the area, is %v", got)
	}
	if got := InArea(project.LonLat{Lon: 0, Lat: 0}, over); got != Outside {
		t.Errorf("the other side of the world is %v", got)
	}
	// And the distance across the seam is short, not most of the way round.
	edge, ok := NearestEdge(project.LonLat{Lon: 178, Lat: 0}, over)
	if !ok {
		t.Fatal("no edge found")
	}
	if edge.Km > 200 {
		t.Errorf("a degree across the seam measured %v km; longitude is circular", edge.Km)
	}
	if math.Abs(edge.Bearing-90) > 2 {
		t.Errorf("the edge across the seam is at bearing %v, want due east", edge.Bearing)
	}
}

// TestEastwardIsTheShortWay: the one rule the rest of the package leans on.
func TestEastwardIsTheShortWay(t *testing.T) {
	cases := []struct {
		from, to, want float64
	}{
		{0, 10, 10},
		{10, 0, -10},
		{179, -179, 2},
		{-179, 179, -2},
		{-180, 180, 0},
	}
	for _, c := range cases {
		if got := eastward(c.from, c.to); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("east from %v to %v is %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

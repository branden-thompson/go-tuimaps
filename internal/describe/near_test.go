package describe

import (
	"math"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// TestNearestPointAndLine is plan task 11.5: the nearest point and the
// nearest line to a place, each with its own label, measured on the sphere.
func TestNearestPointAndLine(t *testing.T) {
	home := project.LonLat{Lon: -84.39, Lat: 33.75} // the place everything is measured from
	points := []Labelled{
		{Run: []project.LonLat{{Lon: -84.0, Lat: 33.75}}, Label: "lightning, 12:04"},
		{Run: []project.LonLat{{Lon: -90.0, Lat: 33.75}}, Label: "lightning, 12:06"},
		{Run: []project.LonLat{{Lon: -84.39, Lat: 34.75}, {Lon: -84.39, Lat: 35.75}}, Label: "two strikes north"},
	}
	near, ok := NearestPoint(home, points)
	if !ok {
		t.Fatal("no nearest point at all")
	}
	if near.Label != "lightning, 12:04" {
		t.Errorf("the nearest point is %q at %v km", near.Label, near.Km)
	}
	if math.Abs(near.Km-36) > 3 {
		t.Errorf("it is %v km away; 0.39 degrees of longitude at 33.75 north is about 36 km", near.Km)
	}
	if math.Abs(near.Bearing-90) > 2 {
		t.Errorf("it is at bearing %v, want due east", near.Bearing)
	}
	// A run of several positions is searched through, not just its first.
	onlyNorth, ok := NearestPoint(project.LonLat{Lon: -84.39, Lat: 40}, []Labelled{points[2]})
	if !ok || math.Abs(onlyNorth.At.Lat-35.75) > 0.001 {
		t.Errorf("the nearest of a run of two is %+v, want the northern one", onlyNorth.At)
	}

	// A line is nearest along its length, not only at its ends: a track that
	// passes the place has its nearest point in the middle of a stretch.
	track := []Labelled{{Run: []project.LonLat{
		{Lon: -86, Lat: 33.75}, {Lon: -82, Lat: 33.75},
	}, Label: "storm track"}}
	line, ok := NearestLine(home, track)
	if !ok {
		t.Fatal("no nearest line")
	}
	if line.Label != "storm track" {
		t.Errorf("the nearest line is %q", line.Label)
	}
	if line.Km > 1 {
		t.Errorf("a track straight through the place is %v km away", line.Km)
	}
	if math.Abs(line.At.Lon-home.Lon) > 0.01 {
		t.Errorf("the nearest point of the track is at %v, want beside the place", line.At.Lon)
	}
	// A line of one position is a point, and is still answered for.
	single := []Labelled{{Run: []project.LonLat{{Lon: -84.0, Lat: 33.75}}, Label: "one position"}}
	if got, ok := NearestLine(home, single); !ok || got.Label != "one position" {
		t.Errorf("a line of one position: %+v, %v", got, ok)
	}
	// Nothing to measure against is said plainly.
	if _, ok := NearestPoint(home, nil); ok {
		t.Error("an empty set of points has a nearest")
	}
	if _, ok := NearestLine(home, []Labelled{{Run: nil, Label: "empty"}}); ok {
		t.Error("a line with no positions has a nearest")
	}
}

// TestNearestAcrossTheSeam: a line and a point on the other side of the
// seam are close neighbours, as an area's edge is.
func TestNearestAcrossTheSeam(t *testing.T) {
	at := project.LonLat{Lon: 179.5, Lat: 0}
	point, ok := NearestPoint(at, []Labelled{{Run: []project.LonLat{{Lon: -179.5, Lat: 0}}, Label: "over the seam"}})
	if !ok {
		t.Fatal("nothing found")
	}
	if point.Km > 200 {
		t.Errorf("a degree across the seam is %v km", point.Km)
	}
	if math.Abs(point.Bearing-90) > 2 {
		t.Errorf("it is at bearing %v, want due east", point.Bearing)
	}
	line, ok := NearestLine(at, []Labelled{{Run: []project.LonLat{{Lon: 179, Lat: -5}, {Lon: -179, Lat: 5}}, Label: "over the seam"}})
	if !ok || line.Km > 600 {
		t.Errorf("a line across the seam is %v km away", line.Km)
	}
}

// TestLineAndAreaEdgeAgree: the nearest point of a line and the nearest
// point of the same geometry read as an area's edge are the same place, so
// a host never hears two answers for one shape.
func TestLineAndAreaEdgeAgree(t *testing.T) {
	at := project.LonLat{Lon: 0, Lat: 0}
	run := []project.LonLat{{Lon: -5, Lat: 3}, {Lon: 5, Lat: 3}, {Lon: 5, Lat: 9}}
	line, ok := NearestLine(at, []Labelled{{Run: run, Label: "as a line"}})
	if !ok {
		t.Fatal("no line")
	}
	// The same run as an area's ring, where the closing edge is further away.
	edge, ok := NearestEdge(at, [][]project.LonLat{run})
	if !ok {
		t.Fatal("no edge")
	}
	if math.Abs(line.Km-edge.Km) > 0.001 || math.Abs(line.Bearing-edge.Bearing) > 0.001 {
		t.Errorf("as a line %v km at %v; as an edge %v km at %v", line.Km, line.Bearing, edge.Km, edge.Bearing)
	}
}

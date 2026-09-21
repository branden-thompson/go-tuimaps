package overlay

import (
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestPrepare: what a Work call makes of an overlay for one zoom bucket - the
// library's own simplified copy, which is what is drawn (FR-11).
func TestPrepare(t *testing.T) {
	big := circle(project.LonLat{Lon: -95, Lat: 38}, 2, 5000)
	speck := square(-90, 30, 0.00001)
	o := Overlay{ID: "warnings", Valid: noon, Keeps: time.Hour, Features: []Feature{
		{Kind: Polygon, Rings: [][]project.LonLat{big, speck}, Role: colour.AlertSevereOutline, Label: "Tornado Warning"},
		{Kind: Line, Rings: [][]project.LonLat{{{Lon: -100, Lat: 30}, {Lon: -90, Lat: 40}}}, Role: colour.Track, Label: "Track"},
		{Kind: Point, Rings: [][]project.LonLat{{{Lon: -87.6, Lat: 41.9}}}, Role: colour.Marker, Label: "Chicago"},
		{Kind: Circle, Centre: project.LonLat{Lon: -80, Lat: 25}, RadiusKm: 100, Role: colour.AlertMinorOutline},
	}}
	shapes, err := Prepare(o, 5)
	if err != nil || len(shapes) != 4 {
		t.Fatalf("%d shapes, %v", len(shapes), err)
	}
	area := shapes[0]
	if area.Kind != scene.ShapeArea || area.Role != uint8(colour.AlertSevereOutline) || area.Label != "Tornado Warning" {
		t.Errorf("%+v", area)
	}
	if len(area.Rings) != 1 {
		t.Fatalf("%d rings; the speck is smaller than a dot at this bucket and is dropped", len(area.Rings))
	}
	if n := len(area.Rings[0]); n >= 5001 || n < 8 || area.Rings[0][0] != area.Rings[0][n-1] {
		t.Errorf("the ring kept %d of 5,001 vertices, closed=%v", n, area.Rings[0][0] == area.Rings[0][n-1])
	}
	deeper, _ := Prepare(o, 12)
	if len(deeper[0].Rings[0]) <= len(area.Rings[0]) {
		t.Error("a deeper bucket keeps no more of the shape than a shallow one")
	}
	if shapes[1].Kind != scene.ShapeLine || len(shapes[1].Rings[0]) != 2 || shapes[2].Kind != scene.ShapePoint || len(shapes[2].Rings[0]) != 1 {
		t.Errorf("line %+v, point %+v", shapes[1], shapes[2])
	}
	// A circle becomes a ring, every vertex the radius from the centre.
	ring := shapes[3]
	if ring.Kind != scene.ShapeArea || len(ring.Rings) != 1 || len(ring.Rings[0]) < 16 {
		t.Fatalf("%+v", ring)
	}
	for _, v := range ring.Rings[0] {
		p, err := project.FromTile(float64(v.X)/(1<<32), float64(v.Y)/(1<<32), 0)
		if err != nil {
			t.Fatal(err)
		}
		if km, _ := project.GreatCircleKm(o.Features[3].Centre, p); math.Abs(km-100) > 1.5 {
			t.Errorf("a vertex of the circle is %.1f km from its centre", km)
		}
	}
	// Across the antimeridian: a ring either side.
	across := Overlay{ID: "x", Valid: noon, Keeps: time.Hour, Features: []Feature{{Kind: Polygon, Role: colour.AlertMinorOutline,
		Rings: [][]project.LonLat{{{Lon: 170, Lat: -10}, {Lon: -170, Lat: -10}, {Lon: -170, Lat: 10}, {Lon: 170, Lat: 10}, {Lon: 170, Lat: -10}}}}}}
	parts, err := Prepare(across, 3)
	if err != nil || len(parts) != 1 || len(parts[0].Rings) != 2 {
		t.Errorf("%+v, %v", parts, err)
	}
	if _, err := Prepare(Overlay{}, 3); err == nil {
		t.Error("an overlay that was never accepted cannot be prepared")
	}
}

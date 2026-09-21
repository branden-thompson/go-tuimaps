package tuimaps_test

import (
	"context"
	"encoding/json"
	"math"
	"path/filepath"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// coastalZone is the largest zone shape the fixture carries: a Gulf-coast
// county of 11,135 vertices, which is the hard case risk RS-19 names.
//
// **PLAN's task says 14,001 vertices.** That figure came from a coastal
// sample measured in DISCOVER which was never committed; this uses the
// largest shape that *is* committed, so the case can be rebuilt from this
// repository alone - the same choice scenario 2 made, and said so.
const coastalZone = "FLC015"

// rings reads a zone's geometry, whatever shape the feed wrote it in: one
// polygon, several, or a collection of them.
type rings struct {
	Geometry geometry `json:"geometry"`
}

type geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
	Geometries  []geometry      `json:"geometries"`
}

// runs is every ring of a geometry, in the order it was written.
func (g geometry) runs(t *testing.T) [][]tuimaps.LonLat {
	t.Helper()
	if len(g.Geometries) > 0 {
		var all [][]tuimaps.LonLat
		for _, one := range g.Geometries {
			all = append(all, one.runs(t)...)
		}
		return all
	}
	switch g.Type {
	case "Polygon":
		var raw [][][2]float64
		if err := json.Unmarshal(g.Coordinates, &raw); err != nil {
			t.Fatal(err)
		}
		return ringsOf(raw)
	case "MultiPolygon":
		var raw [][][][2]float64
		if err := json.Unmarshal(g.Coordinates, &raw); err != nil {
			t.Fatal(err)
		}
		var all [][]tuimaps.LonLat
		for _, polygon := range raw {
			all = append(all, ringsOf(polygon)...)
		}
		return all
	}
	t.Fatalf("the zone's geometry is a %s, which this test does not read", g.Type)
	return nil
}

// ringsOf turns written pairs into positions.
func ringsOf(raw [][][2]float64) [][]tuimaps.LonLat {
	out := make([][]tuimaps.LonLat, 0, len(raw))
	for _, ring := range raw {
		run := make([]tuimaps.LonLat, 0, len(ring)+1)
		for _, at := range ring {
			run = append(run, tuimaps.LonLat{Lon: at[0], Lat: at[1]})
		}
		if len(run) > 0 && run[0] != run[len(run)-1] {
			run = append(run, run[0])
		}
		out = append(out, run)
	}
	return out
}

// TestCoastalZone14001 is plan task 14.11 and risk RS-19: **the hard case,
// from real data.** A Gulf-coast county's zone shape has thousands of
// vertices threading in and out of the water, and the question a warning
// asks of it - am I in this zone? - must have one answer, whatever the map
// is doing.
//
// The answer is held to two things: a rule written out here, which is not
// the library's (a winding number where the library counts crossings), and
// itself, as the view moves and zooms across the whole range a person can
// look at it in.
func TestCoastalZone14001(t *testing.T) {
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	body, err := testkit.LoadFixture(root, filepath.ToSlash(filepath.Join("zones", coastalZone+".json")))
	if err != nil {
		t.Fatal(err)
	}
	var zone rings
	if err := json.Unmarshal(body, &zone); err != nil {
		t.Fatal(err)
	}
	shape := zone.Geometry.runs(t)
	vertices := 0
	for _, run := range shape {
		vertices += len(run)
	}
	if vertices < 10000 {
		t.Fatalf("%s has %d vertices; this test is about the hard case", coastalZone, vertices)
	}
	t.Logf("%s: %d rings, %d vertices", coastalZone, len(shape), vertices)

	// A spread of places across the zone's own box: some in it, some out,
	// some threading its coastline.
	west, south, east, north := box(shape)
	var places []tuimaps.Place
	for row := range 12 {
		for col := range 12 {
			at := tuimaps.LonLat{
				Lon: west + (east-west)*(float64(col)+0.5)/12,
				Lat: south + (north-south)*(float64(row)+0.5)/12,
			}
			places = append(places, tuimaps.Place{Name: "p" + itoa(row*12+col), At: at})
		}
	}

	m := world(t, 149, 38)
	if _, err := m.SetPlaces(places); err != nil {
		t.Fatal(err)
	}
	feature := tuimaps.Feature{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, Label: "Coastal Flood Warning", Rings: shape}
	if _, err := m.Set(tuimaps.Overlay{ID: "zone", Valid: noon, Keeps: 3600000000000, Credit: "National Weather Service",
		Features: []tuimaps.Feature{feature}}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}

	// The library's answer, against a rule written out here from the
	// definition - a winding number, where the library counts crossings.
	said, err := m.Describe(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(said) != len(places) {
		t.Fatalf("%d places described, %d set", len(said), len(places))
	}
	first := map[string]bool{}
	disagreed := 0
	for i, one := range said {
		if len(one.Answers) != 1 {
			t.Fatalf("%s has %d answers", one.Place, len(one.Answers))
		}
		mine := one.Answers[0].Relation.String() == "inside"
		first[one.Place] = mine
		theirs := wound(shape, places[i].At)
		if mine != theirs && one.Answers[0].Distance > 0.2 {
			// Within 200 m of the boundary the two rules may differ on
			// which side a point falls, and neither is wrong: that is the
			// knife edge D-67 names. Further out they must agree.
			disagreed++
			t.Errorf("%s at %+v: the library says %v, the winding number says %v, %.2f km from the edge",
				one.Place, places[i].At, mine, theirs, one.Answers[0].Distance)
		}
	}
	if disagreed > 0 {
		t.Fatalf("%d of %d places disagree with a rule written from the definition", disagreed, len(places))
	}
	// Agreement is evidence only if both answers were actually given: a
	// grid that fell entirely outside the zone would agree with anything.
	in := 0
	for _, inside := range first {
		if inside {
			in++
		}
	}
	if in == 0 || in == len(first) {
		t.Fatalf("%d of %d places are inside the zone; a grid that is all one answer proves nothing", in, len(first))
	}
	t.Logf("%d of %d places inside, %d outside, every one agreeing with the winding number", in, len(first), len(first)-in)

	// **It never flips.** The answer is the data's, so moving and zooming
	// the map cannot change it: every zoom a person can look at this in,
	// and a pan across the whole zone at each.
	for _, zoom := range []float64{6, 9, 12, 14} {
		for _, at := range []tuimaps.LonLat{{Lon: west, Lat: south}, {Lon: east, Lat: north}, {Lon: (west + east) / 2, Lat: (south + north) / 2}} {
			if err := m.Recentre(at); err != nil {
				t.Fatal(err)
			}
			if err := m.Zoom(zoom); err != nil {
				t.Fatal(err)
			}
			if _, err := m.Settle(context.Background()); err != nil {
				t.Fatal(err)
			}
			if _, err := m.Render(tuimaps.Size{Cols: 149, Rows: 38}, noon); err != nil {
				t.Fatal(err)
			}
			again, err := m.Describe(nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, one := range again {
				if now := one.Answers[0].Relation.String() == "inside"; now != first[one.Place] {
					t.Fatalf("%s changed from %v to %v at zoom %v centred on %+v; the answer is the data's, not the view's",
						one.Place, first[one.Place], now, zoom, at)
				}
			}
		}
	}
}

// box is the smallest rectangle holding every ring.
func box(shape [][]tuimaps.LonLat) (west, south, east, north float64) {
	west, south, east, north = 180, 90, -180, -90
	for _, run := range shape {
		for _, at := range run {
			west, east = math.Min(west, at.Lon), math.Max(east, at.Lon)
			south, north = math.Min(south, at.Lat), math.Max(north, at.Lat)
		}
	}
	return west, south, east, north
}

// wound reports whether a place is inside a shape by the **winding
// number**, written out here from its definition. It is deliberately a
// different rule from the one the library uses, so that agreement between
// them is evidence and not a reflection (D-43).
func wound(shape [][]tuimaps.LonLat, at tuimaps.LonLat) bool {
	turns := 0.0
	for _, run := range shape {
		for i := 0; i+1 < len(run); i++ {
			turns += angleBetween(at, run[i], run[i+1])
		}
	}
	return math.Abs(turns) > math.Pi
}

// angleBetween is the angle one edge subtends at a place, signed.
func angleBetween(at, a, b tuimaps.LonLat) float64 {
	ax, ay := a.Lon-at.Lon, a.Lat-at.Lat
	bx, by := b.Lon-at.Lon, b.Lat-at.Lat
	return math.Atan2(ax*by-ay*bx, ax*bx+ay*by)
}

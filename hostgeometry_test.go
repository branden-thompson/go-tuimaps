package tuimaps_test

import (
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// hostPoint is a host's own point type, written the way a host really writes
// one: **with struct tags**, because a host that keeps its data also
// serialises it. A tagged struct is a different type from an untagged one, so
// anything that matched on layout alone would refuse this - which is how the
// first host would have been turned away.
//
// Note what it does NOT do: it names no library type. A host's data model
// should not have to import a renderer to describe a position.
type hostPoint struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

func (p hostPoint) LonLat() (float64, float64) { return p.Lon, p.Lat }

// backwards keeps its fields the other way round and its own names, to show
// the helper is about what a point can SAY, not how it is laid out.
type backwards struct {
	Northing float64
	Easting  float64
}

func (b backwards) LonLat() (float64, float64) { return b.Easting, b.Northing }

// TestAHostBuildsAFeatureFromItsOwnPoints is the ergonomic case the first
// integration turned up: a host holds its geometry in its own type, and
// handing it over should be one call, not a nested loop in every host.
func TestAHostBuildsAFeatureFromItsOwnPoints(t *testing.T) {
	// One area: an outline and a hole in it, in the host's own type.
	area := [][]hostPoint{
		{{Lon: -85, Lat: 41}, {Lon: -83, Lat: 41}, {Lon: -83, Lat: 43}, {Lon: -85, Lat: 43}, {Lon: -85, Lat: 41}},
		{{Lon: -84.5, Lat: 41.5}, {Lon: -83.5, Lat: 41.5}, {Lon: -83.5, Lat: 42.5}, {Lon: -84.5, Lat: 42.5}, {Lon: -84.5, Lat: 41.5}},
	}
	rings := tuimaps.Rings(area)
	if len(rings) != 2 || len(rings[0]) != 5 || len(rings[1]) != 5 {
		t.Fatalf("an outline and a hole became %d rings", len(rings))
	}
	if rings[0][0] != (tuimaps.LonLat{Lon: -85, Lat: 41}) {
		t.Errorf("the first position came over as %v", rings[0][0])
	}

	// It is accepted by the library, and the answers are the real ones.
	m, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(tuimaps.Overlay{
		ID: "alerts", Valid: time.Now(), Keeps: time.Hour,
		Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, Rings: rings}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		what string
		at   tuimaps.LonLat
		want tuimaps.Where
	}{
		{"on the land", tuimaps.LonLat{Lon: -83.2, Lat: 42.8}, tuimaps.Inside},
		{"in the hole", tuimaps.LonLat{Lon: -84, Lat: 42}, tuimaps.Outside},
	} {
		got, err := m.Report([]tuimaps.Place{{ID: c.what, Name: c.what, At: c.at}})
		if err != nil {
			t.Fatal(err)
		}
		if a := got.Places[0].Alerts[0]; a.Where != c.want {
			t.Errorf("a place %s is %v; it is %v", c.what, a.Where, c.want)
		}
	}
}

// TestAnyPointTypeThatSaysWhereItIsWorks: the helper asks a point to say its
// position, so field order, field names and struct tags are all the host's own
// business.
func TestAnyPointTypeThatSaysWhereItIsWorks(t *testing.T) {
	got := tuimaps.Ring([]backwards{{Northing: 41, Easting: -85}, {Northing: 42, Easting: -84}})
	want := []tuimaps.LonLat{{Lon: -85, Lat: 41}, {Lon: -84, Lat: 42}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("%v; want %v", got, want)
	}
}

// TestTheLibrarysOwnPointNeedsNoConversion: a host that already holds LonLat
// should not be made to wrap it to use the same call.
func TestTheLibrarysOwnPointNeedsNoConversion(t *testing.T) {
	in := []tuimaps.LonLat{{Lon: -85, Lat: 41}, {Lon: -84, Lat: 42}}
	got := tuimaps.Ring(in)
	if len(got) != 2 || got[0] != in[0] || got[1] != in[1] {
		t.Errorf("%v; want %v", got, in)
	}
}

// TestNothingComesOfNothing: an empty area is empty, not a ring of no
// positions, which the library refuses elsewhere.
func TestNothingComesOfNothing(t *testing.T) {
	if got := tuimaps.Rings([][]hostPoint(nil)); got != nil {
		t.Errorf("no areas became %v", got)
	}
	if got := tuimaps.Ring([]hostPoint(nil)); got != nil {
		t.Errorf("no positions became %v", got)
	}
}

// hostRing and hostArea are what a host that keeps geometry really declares:
// NAMED types, not anonymous slices. The first version of Rings took `[][]P`
// and so refused them - inference cannot match a `[]hostRing` against `[][]P`
// - and every test here passed because every test here used an unnamed
// `[][]hostPoint`. The real integration found it in one compile.
type hostRing []hostPoint
type hostArea []hostRing

// TestAHostThatNamesItsRingTypeIsAccepted.
func TestAHostThatNamesItsRingTypeIsAccepted(t *testing.T) {
	area := hostArea{
		hostRing{{Lon: -85, Lat: 41}, {Lon: -83, Lat: 41}, {Lon: -83, Lat: 43}, {Lon: -85, Lat: 41}},
		hostRing{{Lon: -84.5, Lat: 41.5}, {Lon: -83.5, Lat: 41.5}, {Lon: -83.5, Lat: 42.5}, {Lon: -84.5, Lat: 41.5}},
	}
	rings := tuimaps.Rings(area)
	if len(rings) != 2 || len(rings[0]) != 4 {
		t.Fatalf("a named area of two named rings became %d rings", len(rings))
	}
	if got := tuimaps.Ring(area[0]); len(got) != 4 || got[0] != (tuimaps.LonLat{Lon: -85, Lat: 41}) {
		t.Errorf("a named ring converted to %v", got)
	}
}

package describe

import (
	"math"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// ringFrom makes a ring out of arbitrary bytes: every longitude and
// latitude a place can have, and some that no place can.
func ringFrom(data []byte) []project.LonLat {
	var ring []project.LonLat
	for len(data) >= 4 {
		lon := float64(int16(uint16(data[0])<<8|uint16(data[1]))) / 182.0 // about -180 to 180
		lat := float64(int16(uint16(data[2])<<8|uint16(data[3]))) / 364.0 // about -90 to 90
		ring = append(ring, project.LonLat{Lon: lon, Lat: lat})
		data = data[4:]
	}
	return ring
}

// FuzzArea is the property test of this package's entry points: whatever
// geometry a host hands in, the answers hold together. Nothing panics; the
// answer does not depend on where the seam happens to fall, because
// longitude is circular; and the nearest edge is never further than the
// nearest vertex.
func FuzzArea(f *testing.F) {
	f.Add([]byte("a ring of nothing much"), int16(0), int16(0))
	f.Add([]byte{0, 0, 0, 0, 255, 255, 0, 0, 255, 255, 255, 255}, int16(100), int16(-50))
	f.Add([]byte{0x7f, 0xff, 0, 0, 0x80, 0x00, 0, 0}, int16(32767), int16(0))
	f.Fuzz(func(t *testing.T, data []byte, lon, lat int16) {
		ring := ringFrom(data)
		if len(ring) < 3 {
			t.Skip()
		}
		at := project.LonLat{Lon: float64(lon) / 182.0, Lat: float64(lat) / 364.0}
		if !onGlobe(at) {
			t.Skip()
		}
		// An edge that spans exactly half the world has no short way round:
		// nothing in the data says which way it goes, so turning the world
		// may honestly change the answer. Such a ring is not a fair test of
		// what follows.
		for i := range ring {
			b := ring[(i+1)%len(ring)]
			if math.Abs(math.Abs(eastward(ring[i].Lon, b.Lon))-180) < 1e-9 {
				t.Skip()
			}
		}
		// A place whose own meridian passes exactly through a vertex is on a
		// knife edge: which side of the vertex the ray north passes is
		// decided by the last bit of a floating-point number, and turning the
		// world moves that bit. Nothing can promise an answer there, so the
		// test does not ask for one.
		for _, v := range ring {
			east := math.Abs(eastward(v.Lon, at.Lon))
			if east < 1e-6 || math.Abs(east-180) < 1e-6 {
				t.Skip() // on the place's own meridian, or exactly opposite it
			}
		}
		rings := [][]project.LonLat{ring}
		// A place all but on the boundary is a knife edge: which side of it
		// the place falls is decided by the last bits of a floating-point
		// number, and turning the world moves those bits. A kilometre is far
		// smaller than a cell at any zoom the library draws, so skipping
		// these leaves the property worth something.
		if near, ok := NearestEdge(at, rings); ok && near.Km < 1 {
			t.Skip()
		}
		where := InArea(at, rings)
		if where != Inside && where != Outside {
			t.Fatalf("a place is neither inside nor outside: %v", where)
		}
		// Turning the whole world, place and ring together, cannot change
		// the answer: longitude is circular and the seam is not a place.
		for _, turn := range []float64{90, 180, -90, 37.5} {
			turned := make([]project.LonLat, len(ring))
			for i, v := range ring {
				turned[i] = project.LonLat{Lon: wrapped(v.Lon + turn), Lat: v.Lat}
			}
			moved := project.LonLat{Lon: wrapped(at.Lon + turn), Lat: at.Lat}
			if got := InArea(moved, [][]project.LonLat{turned}); got != where {
				t.Fatalf("turning the world by %v made %v into %v", turn, where, got)
			}
		}
		edge, ok := NearestEdge(at, rings)
		if !ok {
			return
		}
		if math.IsNaN(edge.Km) || math.IsInf(edge.Km, 0) || edge.Km < 0 {
			t.Fatalf("the nearest edge is %v km away", edge.Km)
		}
		if edge.Bearing < 0 || edge.Bearing >= 360 {
			t.Fatalf("the nearest edge is at bearing %v", edge.Bearing)
		}
		if !onGlobe(edge.At) {
			t.Fatalf("the nearest point is not on the globe: %+v", edge.At)
		}
		// No vertex is nearer than the nearest point of the edges.
		for _, v := range ring {
			km, err := project.GreatCircleKm(at, v)
			if err != nil {
				continue
			}
			if km < edge.Km-0.001 {
				t.Fatalf("a vertex is %v km away and the nearest edge %v", km, edge.Km)
			}
		}
	})
}

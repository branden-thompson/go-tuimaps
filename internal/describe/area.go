// Package describe answers, in data a host can speak, the question the
// project exists for: where is this, relative to my place. It is computed
// from the host's own geometry, unsimplified, never from the drawn cells: a
// picture cannot show a distance smaller than one cell, and a screen reader
// reads braille as noise (D-52, FR-29).
package describe

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Where a place is, relative to an area.
type Where uint8

// The answers. A place on an edge is inside or outside by the fill rule
// like any other; the picture may not be able to show which, which is what
// the under-one-cell flag is for (D-67).
const (
	Outside Where = iota + 1
	Inside
)

// String names the answer, in words a speech engine says as they are.
func (w Where) String() string {
	if w == Inside {
		return "inside"
	}
	return "outside"
}

// Edge is the nearest point of an area's edge to a place: how far, which
// way, and on which ring it lies.
type Edge struct {
	Km      float64
	Bearing float64 // degrees clockwise from north
	At      project.LonLat
}

// InArea reports whether a place is inside a feature's rings, by the
// even-odd rule the renderer fills with (FR-11): a point in a hole is
// outside. Longitude is circular, so a ring that crosses the seam at 180
// degrees is followed round it rather than back across the world.
func InArea(at project.LonLat, rings [][]project.LonLat) Where {
	if !onGlobe(at) {
		return Outside // a place that is nowhere is inside nothing
	}
	at = onWorld(at)
	crossings := 0
	for _, ring := range rings {
		crossings += crossingsOf(at, ring)
	}
	if crossings%2 == 1 {
		return Inside
	}
	return Outside
}

// crossingsOf counts the times the ray north from the place crosses one
// ring. **A ray east would not do**: longitude is circular, so a ray east
// has no end - it comes back round and crosses everything an even number of
// times. A ray north ends at the pole, and each edge is tested in its own
// span of longitude, so a ring that crosses the seam is followed over it
// rather than back across the world.
func crossingsOf(at project.LonLat, ring []project.LonLat) int {
	crossings := 0
	for i := range ring {
		a, b := ring[i], ring[(i+1)%len(ring)]
		// Measured from the place itself, which sits at zero: turning the
		// whole world moves every vertex by the same amount and leaves these
		// two numbers where they were, so the answer does not depend on
		// where the seam happens to fall.
		from, to := eastward(at.Lon, a.Lon), eastward(at.Lon, b.Lon)
		span := to - from
		if math.Abs(span) >= 180 || span == 0 {
			continue // it goes the other way round the world, or along a meridian
		}
		if (from > 0) == (to > 0) {
			continue // both ends the same side: the place's meridian is not cut
		}
		along := -from / span
		if a.Lat+along*(b.Lat-a.Lat) > at.Lat {
			crossings++
		}
	}
	return crossings
}

// onGlobe reports whether a place is one. **Longitude is not checked for
// range, it is wrapped**: it is circular, so 180.04 east is 179.96 west and
// nothing is in doubt. Latitude has ends, and a place beyond them is
// nowhere.
func onGlobe(p project.LonLat) bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lon >= -720 && p.Lon <= 720
}

// onWorld is a place with its longitude brought back into -180 to 180.
func onWorld(p project.LonLat) project.LonLat {
	return project.LonLat{Lon: wrapped(p.Lon), Lat: p.Lat}
}

// eastward is how far east one longitude is from another, taken the short
// way round: from -180 up to 180. It is what makes longitude circular
// everywhere in this package.
//
// **An edge that spans exactly half the world has no short way round**, and
// nothing in the data says which way it goes. This takes it eastward, so
// that the answer is at least the same every time; a host whose geometry
// has such an edge has said less than it meant to.
func eastward(from, to float64) float64 {
	d := math.Mod(to-from+540, 360) - 180
	return d
}

// NearestEdge is the nearest point of an area's edge to a place, and how
// far and which way it is. **A stretch of edge that lies along the seam at
// 180 degrees is still an edge**: it is the host's geometry, not a cut the
// library made, and the library never invents one (FR-11).
func NearestEdge(at project.LonLat, rings [][]project.LonLat) (Edge, bool) {
	if !onGlobe(at) || len(rings) == 0 {
		return Edge{}, false
	}
	at = onWorld(at)
	best := Edge{Km: math.Inf(1)}
	found := false
	for _, ring := range rings {
		if len(ring) < 2 {
			continue
		}
		for i := range ring {
			a, b := ring[i], ring[(i+1)%len(ring)]
			// The point along the edge is worked out flat, which is close
			// enough at the sizes a map shows; the two ends are exact. Taking
			// the nearest of the three keeps the answer from ever being worse
			// than a vertex, which a very long edge could otherwise make it.
			// A vertex that is nowhere is passed over on its own, so **one bad
			// vertex does not take its good neighbours with it**.
			candidates := [3]project.LonLat{a, b, {}}
			if onGlobe(a) && onGlobe(b) {
				candidates[2] = nearestOn(at, a, b)
			} else {
				candidates[2] = a
			}
			for _, candidate := range candidates {
				if !onGlobe(candidate) {
					continue
				}
				edge, ok := measured(at, onWorld(candidate))
				if ok && edge.Km < best.Km {
					best, found = edge, true
				}
			}
		}
	}
	return best, found
}

// measured is how far a point is from a place and which way, or false for
// a pair the sphere cannot answer for.
func measured(at, to project.LonLat) (Edge, bool) {
	km, err := project.GreatCircleKm(at, to)
	if err != nil {
		return Edge{}, false
	}
	bearing, err := project.Bearing(at, to)
	if err != nil {
		return Edge{}, false
	}
	return Edge{Km: km, Bearing: bearing, At: to}, true
}

// nearestOn is the point of one edge nearest a place. The edge is treated
// as straight in longitude and latitude, with longitude taken the short way
// round and narrowed by the latitude, which is close enough at the sizes a
// map shows and never wrong about which end is nearer.
func nearestOn(at, a, b project.LonLat) project.LonLat {
	narrow := math.Cos(at.Lat * math.Pi / 180)
	dx, dy := eastward(a.Lon, b.Lon)*narrow, b.Lat-a.Lat
	if dx == 0 && dy == 0 {
		return a
	}
	px, py := eastward(a.Lon, at.Lon)*narrow, at.Lat-a.Lat
	along := (px*dx + py*dy) / (dx*dx + dy*dy)
	along = math.Max(0, math.Min(1, along))
	return project.LonLat{
		Lon: wrapped(a.Lon + along*eastward(a.Lon, b.Lon)),
		Lat: a.Lat + along*dy,
	}
}

// wrapped brings a longitude back into -180 to 180.
func wrapped(lon float64) float64 {
	return math.Mod(lon+540, 360) - 180
}

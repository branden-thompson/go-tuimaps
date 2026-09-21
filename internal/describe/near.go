package describe

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Near is the nearest thing of a kind to a place: how far, which way, and
// what it is called.
type Near struct {
	Km      float64
	Bearing float64 // degrees clockwise from north
	At      project.LonLat
	Label   string // the feature's own label, cleaned by whoever hands it out
}

// Labelled is one piece of geometry with a name: a point is a run of one, a
// line a run of two or more, in the order the host gave them.
type Labelled struct {
	Run   []project.LonLat
	Label string
}

// NearestPoint is the nearest of a set of places to a place, with its own
// label. A point overlay is a set of these: lightning, a fire detection, a
// station.
func NearestPoint(at project.LonLat, points []Labelled) (Near, bool) {
	if !onGlobe(at) {
		return Near{}, false
	}
	at = onWorld(at)
	best, found := Near{Km: math.Inf(1)}, false
	for _, one := range points {
		for _, p := range one.Run {
			if !onGlobe(p) {
				continue
			}
			edge, ok := measured(at, onWorld(p))
			if !ok || edge.Km >= best.Km {
				continue
			}
			best, found = Near{Km: edge.Km, Bearing: edge.Bearing, At: edge.At, Label: one.Label}, true
		}
	}
	return best, found
}

// NearestLine is the nearest point of a set of lines to a place, with the
// line's own label: a storm track, a river, a road. A line is never closed,
// so its last position is an end and not an edge back to its first.
func NearestLine(at project.LonLat, lines []Labelled) (Near, bool) {
	if !onGlobe(at) {
		return Near{}, false
	}
	at = onWorld(at)
	best, found := Near{Km: math.Inf(1)}, false
	for _, one := range lines {
		for i := 0; i+1 < len(one.Run); i++ {
			edge, ok := nearestAlong(at, one.Run[i], one.Run[i+1])
			if !ok || edge.Km >= best.Km {
				continue
			}
			best, found = Near{Km: edge.Km, Bearing: edge.Bearing, At: edge.At, Label: one.Label}, true
		}
		if len(one.Run) == 1 { // a line of one position is a point
			if edge, ok := measured(at, onWorld(one.Run[0])); ok && onGlobe(one.Run[0]) && edge.Km < best.Km {
				best, found = Near{Km: edge.Km, Bearing: edge.Bearing, At: edge.At, Label: one.Label}, true
			}
		}
	}
	return best, found
}

// nearestAlong is the nearest point of one stretch of line to a place: the
// same three candidates the nearest edge of an area is chosen from, so that
// a line and the edge of an area are never measured by different rules.
func nearestAlong(at, a, b project.LonLat) (Edge, bool) {
	best, found := Edge{Km: math.Inf(1)}, false
	candidates := [3]project.LonLat{a, b, a}
	if onGlobe(a) && onGlobe(b) {
		candidates[2] = nearestOn(at, a, b)
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
	return best, found
}

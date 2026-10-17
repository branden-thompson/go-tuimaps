package tuimaps

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/sun"
)

// Terminator is the line between day and night at a moment, its latitude at
// every degree of longitude from 180°W to 180°E, within 0.5° of NOAA's
// solar position (L-4.2): a line a host draws, as an overlay feature in the
// terminator's role.
func Terminator(at time.Time) []LonLat {
	line := sun.Line(at)
	out := make([]LonLat, len(line))
	for i, p := range line {
		out[i] = LonLat{Lon: p[0], Lat: p[1]}
	}
	return out
}

// NightSide is the night at a moment as one ring - the terminator closed
// through the pole the sun does not reach (L-4.2): an area a host draws, as
// an overlay feature in the night's role.
func NightSide(at time.Time) [][]LonLat {
	ring := sun.Night(at)
	out := make([]LonLat, len(ring))
	for i, p := range ring {
		out[i] = LonLat{Lon: p[0], Lat: p[1]}
	}
	return [][]LonLat{out}
}

// NightOverlay is the night at a moment as one overlay, in one call: the
// night side shading the ground and the sea, and the terminator dotted at
// its edge, each in its own role (L-4.2, L-4.3; watchpost D-216). It keeps
// an hour; a host hands in another as its moment moves.
func NightOverlay(id string, at time.Time) Overlay {
	return Overlay{ID: id, Valid: at, Keeps: time.Hour, Features: []Feature{
		{Kind: Polygon, Rings: NightSide(at), Role: NightTint},
		{Kind: Line, Rings: [][]LonLat{Terminator(at)}, Role: TerminatorLine},
	}}
}

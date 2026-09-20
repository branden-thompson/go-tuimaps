// Package overlay is what a host lays over the map: shapes, grids, images
// and their freshness, validated on hand-in and prepared, off the drawing
// path, into what the renderer draws. It draws nothing itself.
package overlay

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Vertex is a position on the world as two 32-bit fractions of its side: 0
// is the west or north edge, 2^32 the east or south. It is 8 bytes, as the
// shape cache's arithmetic assumes, and fine enough for the deepest zoom: a
// braille dot at zoom 18 is 64 of these units across.
type Vertex struct{ X, Y uint32 }

const unit = 1 << 32

// fixed is a fraction of the world's side as a Vertex coordinate.
func fixed(f float64) uint32 {
	if !(f > 0) {
		return 0
	}
	if f >= 1 {
		return math.MaxUint32
	}
	return uint32(f * unit)
}

// Bucket is the zoom bucket a view's zoom falls in: one for each whole zoom
// level, serving every zoom from b up to, not including, b + 1.
func Bucket(zoom float64) int {
	if !(zoom >= project.MinViewZoom) {
		if zoom < project.MinViewZoom {
			return project.MinViewZoom
		}
		return 0 // not a number
	}
	if zoom > project.MaxViewZoom {
		return project.MaxViewZoom
	}
	return int(math.Floor(zoom))
}

// Tolerance is a bucket's simplification tolerance, as a fraction of the
// world's side: half a braille dot at zoom b + 1, the finest zoom the bucket
// serves, so that a simplified shape is never coarser than the view needs.
func Tolerance(bucket int) float64 {
	return 0.5 / (project.TileSize * math.Exp2(float64(bucket+1)))
}

// Nearest is the prepared bucket nearest the one wanted, which is what is
// drawn while the right one is prepared. A tie goes to the finer.
func Nearest(want int, prepared []int) (int, bool) {
	if len(prepared) == 0 {
		return 0, false
	}
	best := prepared[0]
	for _, b := range prepared[1:] {
		gap, bestGap := abs(b-want), abs(best-want)
		if gap < bestGap || (gap == bestGap && b > best) {
			best = b
		}
	}
	return best, true
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func invalidCoordinates() error {
	return fault.Make(fault.InvalidCoordinates,
		textsafe.Const("the overlay was refused"),
		textsafe.Const("one of its coordinates is not a number, or is off the globe"),
		textsafe.Const("give longitude from -180 to 180 and latitude from -90 to 90, in that order"))
}

// Project takes a ring's positions to the world's fractions.
func Project(ring []project.LonLat) ([]Vertex, error) {
	if len(ring) == 0 {
		return nil, nil
	}
	out := make([]Vertex, 0, len(ring))
	for _, p := range ring {
		x, y, err := project.ToTile(p, 0)
		if err != nil {
			return nil, invalidCoordinates()
		}
		out = append(out, Vertex{X: fixed(x), Y: fixed(y)})
	}
	return out, nil
}

// distance is how far p lies from the segment a to b, as a fraction of the
// world's side.
func distance(p, a, b Vertex) float64 {
	px, py := float64(p.X)/unit, float64(p.Y)/unit
	ax, ay := float64(a.X)/unit, float64(a.Y)/unit
	bx, by := float64(b.X)/unit, float64(b.Y)/unit
	dx, dy := bx-ax, by-ay
	length := dx*dx + dy*dy
	if length == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := math.Max(0, math.Min(1, ((px-ax)*dx+(py-ay)*dy)/length))
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

type span struct{ from, to int }

// Simplify is Ramer-Douglas-Peucker on one ring, which stays closed: the
// first and last vertices are always kept. It follows the ring with a list
// of spans, not by calling itself, and its work is bounded: a shape built to
// make every split lopsided costs a plain version the square of its length.
// Spans the budget does not reach are kept whole, which is never coarser
// than the tolerance.
func Simplify(ring []Vertex, tolerance float64) []Vertex {
	if len(ring) < 3 || !(tolerance > 0) {
		return ring
	}
	keep := make([]bool, len(ring))
	keep[0], keep[len(ring)-1] = true, true
	budget := 48 * len(ring) * (bitLength(len(ring)) + 1)
	open := []span{{0, len(ring) - 1}}
	for range 2 * len(ring) {
		if len(open) == 0 {
			break
		}
		s := open[len(open)-1]
		open = open[:len(open)-1]
		if s.to-s.from < 2 {
			continue
		}
		if budget < s.to-s.from {
			for i := s.from; i <= s.to; i++ {
				keep[i] = true
			}
			continue
		}
		budget -= s.to - s.from
		far, farthest := -1, tolerance
		for i := s.from + 1; i < s.to; i++ {
			if d := distance(ring[i], ring[s.from], ring[s.to]); d > farthest {
				far, farthest = i, d
			}
		}
		if far < 0 {
			continue
		}
		keep[far] = true
		open = append(open, span{s.from, far}, span{far, s.to})
	}
	out := make([]Vertex, 0, len(ring)/4+4)
	for i, k := range keep {
		if k {
			out = append(out, ring[i])
		}
	}
	return out
}

func bitLength(n int) int {
	bits := 0
	for range 64 {
		if n == 0 {
			break
		}
		bits++
		n >>= 1
	}
	return bits
}

// Keep reports whether a simplified ring is still a ring: three distinct
// points or more. One that is not is smaller than a dot and is dropped.
func Keep(ring []Vertex) ([]Vertex, bool) {
	if len(ring) < 4 {
		return nil, false
	}
	distinct := 1
	for i := 1; i < len(ring); i++ {
		if ring[i] != ring[i-1] && ring[i] != ring[0] {
			distinct++
		}
	}
	if distinct < 3 {
		return nil, false
	}
	return ring, true
}

// unwrap follows a ring across the antimeridian: where longitude jumps by more
// than half the world, the ring went the short way round, and a whole turn is
// added or taken away. It returns the ring's extent as unwrapped.
func unwrap(ring []project.LonLat) (out []project.LonLat, west, east float64, err error) {
	if len(ring) == 0 {
		return nil, 0, 0, nil
	}
	out = make([]project.LonLat, len(ring))
	shift := 0.0
	west, east = math.Inf(1), math.Inf(-1)
	for i, p := range ring {
		if !(p.Lon >= -180 && p.Lon <= 180 && p.Lat >= -90 && p.Lat <= 90) {
			return nil, 0, 0, invalidCoordinates()
		}
		if i > 0 && p.Lon+shift-out[i-1].Lon > 180 {
			shift -= 360
		}
		if i > 0 && p.Lon+shift-out[i-1].Lon < -180 {
			shift += 360
		}
		out[i] = project.LonLat{Lon: p.Lon + shift, Lat: p.Lat}
		west, east = math.Min(west, out[i].Lon), math.Max(east, out[i].Lon)
	}
	return out, west, east, nil
}

// Split cuts a ring that crosses the antimeridian into a ring either side,
// each within -180 to 180. A ring that does not cross is returned as it is.
func Split(ring []project.LonLat) ([][]project.LonLat, error) {
	if len(ring) == 0 {
		return nil, nil
	}
	unwrapped, west, east, err := unwrap(ring)
	if err != nil {
		return nil, err
	}
	if west >= -180 && east <= 180 {
		return [][]project.LonLat{ring}, nil
	}
	var parts [][]project.LonLat
	for _, band := range []float64{-360, 0, 360} {
		part := clip(clip(unwrapped, band-180, true), band+180, false)
		if len(part) < 3 {
			continue
		}
		for i := range part {
			part[i].Lon -= band
		}
		parts = append(parts, append(part, part[0]))
	}
	return parts, nil
}

// clip keeps the part of a ring on one side of a meridian: east of it, or
// west. The ring comes back open; its cut edge lies along the meridian.
func clip(ring []project.LonLat, meridian float64, keepEast bool) []project.LonLat {
	if len(ring) < 3 {
		return nil
	}
	inside := func(p project.LonLat) bool {
		if keepEast {
			return p.Lon >= meridian
		}
		return p.Lon <= meridian
	}
	var out []project.LonLat
	prev := ring[len(ring)-1]
	for _, p := range ring {
		if inside(p) != inside(prev) {
			t := (meridian - prev.Lon) / (p.Lon - prev.Lon)
			out = append(out, project.LonLat{Lon: meridian, Lat: prev.Lat + t*(p.Lat-prev.Lat)})
		}
		if inside(p) {
			out = append(out, p)
		}
		prev = p
	}
	return dedupe(out)
}

// dedupe drops a vertex that repeats the one before it.
func dedupe(ring []project.LonLat) []project.LonLat {
	if len(ring) == 0 {
		return nil
	}
	out := ring[:1]
	for _, p := range ring[1:] {
		if p != out[len(out)-1] {
			out = append(out, p)
		}
	}
	if len(out) > 1 && out[0] == out[len(out)-1] {
		out = out[:len(out)-1]
	}
	return out
}

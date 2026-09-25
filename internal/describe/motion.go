package describe

import (
	"math"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Motion's rules (L-1.12, D-72): the heavier rain in a frame is each
// connected area of pixels at or above one threshold class, held for the
// whole loop; a cell is placed at its area's centroid. Following a cell from
// one frame to the frame before, the nearest area there is the same cell if
// it lies within what rain can travel in the time between them.
const (
	// cellSpeedKmh is the fastest a storm cell is taken to move: faster than
	// any but a very few, so a cell is not lost between two frames.
	cellSpeedKmh = 150.0
	// leastStepKm is the least a cell may have moved between two frames and
	// still be held, whatever their times: a centroid shifts as a cell grows.
	leastStepKm = 10.0
	// heldKm is how much nearer or farther a cell must end up to have come
	// closer or moved away; less than it, the cell held.
	heldKm = 2.0
)

// Trend is how a cell's distance from a place went over the loop.
type Trend uint8

// The trends.
const (
	Held Trend = iota + 1
	Closer
	Away
)

// String names the trend, in words a speech engine says as they are.
func (t Trend) String() string {
	switch t {
	case Closer:
		return "closer"
	case Away:
		return "away"
	}
	return "held"
}

// Frame is one observed frame of a loop, as classes.
type Frame struct {
	Valid time.Time
	Image Image
}

// Cells are a frame's areas of heavier rain: each 4-connected area of pixels
// at or above the threshold class, at its centroid.
func (i Image) Cells(threshold int) []project.LonLat {
	seen := make([]bool, len(i.Classes))
	var out []project.LonLat
	var stack []int
	for start, class := range i.Classes {
		if seen[start] || int(class) < threshold {
			continue
		}
		var lon, lat float64
		n := 0
		seen[start] = true
		stack = append(stack[:0], start)
		for len(stack) > 0 {
			at := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := at%i.Width, at/i.Width
			c := i.centre(x, y)
			lon, lat, n = lon+c.Lon, lat+c.Lat, n+1
			for _, next := range [4][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
				if next[0] < 0 || next[0] >= i.Width || next[1] < 0 || next[1] >= i.Height {
					continue
				}
				k := next[1]*i.Width + next[0]
				if !seen[k] && int(i.Classes[k]) >= threshold {
					seen[k] = true
					stack = append(stack, k)
				}
			}
		}
		out = append(out, project.LonLat{Lon: lon / float64(n), Lat: lat / float64(n)})
	}
	return out
}

// Sighting is where a cell was, and when.
type Sighting struct {
	Valid time.Time
	At    project.LonLat
}

// Track follows, through the frames given oldest first, the cell nearest a
// place in the newest frame back to the oldest frame it can be held in. It
// is false when the newest frame has no cell, or there are fewer than two
// frames.
func Track(frames []Frame, threshold int, near project.LonLat) (from, to Sighting, ok bool) {
	if len(frames) < 2 {
		return Sighting{}, Sighting{}, false
	}
	newest := frames[len(frames)-1]
	at, found := nearestCell(newest.Image.Cells(threshold), near, math.Inf(1))
	if !found {
		return Sighting{}, Sighting{}, false
	}
	to = Sighting{Valid: newest.Valid, At: at}
	from = to
	for k := len(frames) - 2; k >= 0; k-- {
		step := math.Max(leastStepKm, cellSpeedKmh*from.Valid.Sub(frames[k].Valid).Hours())
		prev, held := nearestCell(frames[k].Image.Cells(threshold), from.At, step)
		if !held {
			break
		}
		from = Sighting{Valid: frames[k].Valid, At: prev}
	}
	return from, to, from.Valid.Before(to.Valid)
}

// nearestCell is the cell nearest a place, within a distance.
func nearestCell(cells []project.LonLat, to project.LonLat, withinKm float64) (project.LonLat, bool) {
	best, bestKm := project.LonLat{}, math.Inf(1)
	for _, c := range cells {
		if km, err := project.GreatCircleKm(to, c); err == nil && km < bestKm && km <= withinKm {
			best, bestKm = c, km
		}
	}
	return best, !math.IsInf(bestKm, 1)
}

// TrendOf is how the distance from a place went, from the first sighting to
// the second.
func TrendOf(fromKm, toKm float64) Trend {
	switch {
	case toKm < fromKm-heldKm:
		return Closer
	case toKm > fromKm+heldKm:
		return Away
	}
	return Held
}

// Measure is how far a place is from another, and which way.
func Measure(from, to project.LonLat) (km, bearing float64, compass string, ok bool) {
	edge, ok := measured(from, to)
	if !ok {
		return 0, 0, "", false
	}
	return edge.Km, edge.Bearing, compassOf(edge.Bearing), true
}

package describe

import (
	"math"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Motion's rules (L-1.12, D-72, D-122): the heavier rain in a frame is every
// pixel at or above one threshold class, held for the whole loop. Its motion
// near a place is measured over every pair of frames in turn: the heavier
// rain within flowRadiusKm of the nearest heavier rain at the newest frame is
// shifted until it best meets the heavier rain of the next frame, and the
// shifts are added up over the loop.
const (
	// cellSpeedKmh is the fastest a storm cell is taken to move: faster than
	// any but a very few. It bounds how far a frame is shifted to meet the next.
	cellSpeedKmh = 150.0
	// flowRadiusKm is how far around the nearest heavier rain the rain is
	// taken to move together: a storm and its neighbours, not the next system.
	flowRadiusKm = 100.0
	// leastHot is the fewest pixels of heavier rain a pair of frames is
	// measured by; fewer say nothing about where rain moves.
	leastHot = 8
	// widestShift bounds, in pixels, how far a frame is shifted either way.
	widestShift = 64
	// stillKmh is the slowest the rain is said to move at all; below it, its
	// heading is a pixel's noise and is not given.
	stillKmh = 5.0
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
	case Held:
		return "held"
	}
	return ""
}

// Frame is one observed frame of a loop, as classes.
type Frame struct {
	Valid time.Time
	Image Image
	// steps is how the rain moved from the frame before to this one, for
	// each window measured, so places whose nearest rain is the same area
	// share one measure.
	steps map[window]step
	// counted is the frame's heavier rain in blocks, counted once a frame.
	counted *blocks
}

// window is where a pair of frames is measured: around a point, at a
// threshold.
type window struct {
	at        project.LonLat
	threshold int
}

// step is how far the heavier rain moved between two frames, in kilometres
// east and north, and whether it could be measured.
type step struct {
	east, north float64
	ok          bool
}

// Sighting is where the heavier rain was, and when.
type Sighting struct {
	Valid time.Time
	At    project.LonLat
}

// Motion is the heavier rain's observed motion near a place: the nearest
// heavier rain at the newest frame (To), where it stood the time measured
// before by that motion (From, inferred, not seen), and the heading and
// speed it moved at, over the pairs of frames that could be measured.
// Heading is the compass bearing the rain moves towards, not from; it means
// nothing when Moving is false, below stillKmh.
type Motion struct {
	From, To Sighting
	Heading  float64
	SpeedKmh float64
	Moving   bool
}

// Missing is why a loop has no motion to tell.
type Missing uint8

// The reasons.
const (
	NoHeavierRain Missing = iota + 1 // no heavier rain at the newest frame
	TooFewFrames                     // fewer than two frames could be measured
	Decoding                         // frames are still being read
)

// String names the reason, in words a speech engine says as they are.
func (m Missing) String() string {
	switch m {
	case NoHeavierRain:
		return "no heavier rain"
	case TooFewFrames:
		return "too few frames"
	case Decoding:
		return "frames still being read"
	}
	return ""
}

// Track measures, through the frames given oldest first, the motion of the
// heavier rain near a place. The frames keep each measure, so a place whose
// nearest heavier rain is the same pixel costs no second one.
func Track(frames []Frame, threshold int, near project.LonLat) (Motion, Missing) {
	if len(frames) < 2 {
		return Motion{}, TooFewFrames
	}
	newest := &frames[len(frames)-1]
	at, found := newest.Image.nearestHot(threshold, near)
	if !found {
		return Motion{}, NoHeavierRain
	}
	var east, north float64
	var covered time.Duration
	for k := 1; k < len(frames); k++ {
		moved := frames[k].stepFrom(&frames[k-1], window{at: at, threshold: threshold})
		if !moved.ok {
			continue
		}
		east, north = east+moved.east, north+moved.north
		covered += frames[k].Valid.Sub(frames[k-1].Valid)
	}
	if covered <= 0 {
		return Motion{}, TooFewFrames
	}
	speed := math.Hypot(east, north) / covered.Hours()
	heading := math.Mod(math.Atan2(east, north)*180/math.Pi+360, 360)
	// From is inferred, not seen: where the rain stood the measured time
	// before the newest frame, by the motion measured.
	from := Sighting{Valid: newest.Valid.Add(-covered), At: offset(at, -east, -north)}
	return Motion{From: from, To: Sighting{Valid: newest.Valid, At: at}, Heading: heading, SpeedKmh: speed, Moving: speed >= stillKmh}, 0
}

// offset is a point moved a distance east and north, in kilometres: flat,
// which is close enough at the distances a loop moves rain.
func offset(at project.LonLat, east, north float64) project.LonLat {
	lat := at.Lat + north/kmPerDegree
	narrow := math.Max(math.Cos(at.Lat*math.Pi/180), 1e-6)
	return project.LonLat{Lon: wrapped(at.Lon + east/(kmPerDegree*narrow)), Lat: math.Max(-90, math.Min(90, lat))}
}

// kmPerDegree is a degree of latitude, in kilometres.
const kmPerDegree = 111.32

// stepFrom is how the heavier rain around a window moved from the frame
// before to this one, measured once a window and kept.
func (f *Frame) stepFrom(before *Frame, w window) step {
	if got, ok := f.steps[w]; ok {
		return got
	}
	moved := measureStep(before, f, w)
	if f.steps == nil {
		f.steps = map[window]step{}
	}
	f.steps[w] = moved
	return moved
}

// blocks is a frame's heavier rain counted in square blocks of pixels about
// blockKm a side, so that a step is measured over hundreds of blocks rather
// than tens of thousands of pixels.
type blocks struct {
	side, threshold int
	cols, rows      int
	count           []int32
}

// blockKm is about how wide a block is: finer than a storm cell, coarse
// enough that a step between frames is a few blocks.
const blockKm = 2.0

// blocksOf is the frame's blocks, counted once a frame and kept.
func (f *Frame) blocksOf(side, threshold int) *blocks {
	if f.counted != nil && f.counted.side == side && f.counted.threshold == threshold {
		return f.counted
	}
	i := f.Image
	b := &blocks{side: side, threshold: threshold, cols: (i.Width + side - 1) / side, rows: (i.Height + side - 1) / side}
	b.count = make([]int32, b.cols*b.rows)
	for y := range i.Height {
		for x := range i.Width {
			if int(i.Classes[y*i.Width+x]) >= threshold {
				b.count[(y/side)*b.cols+x/side]++
			}
		}
	}
	f.counted = b
	return b
}

// measureStep shifts the heavier rain of one frame, within flowRadiusKm of
// the window's point, to where it best meets the heavier rain of the next,
// block by block, within what rain can travel between them; then places the
// best shift between blocks by the parabola through its neighbours' fits.
// Of equal fits the smaller shift wins, so rain that held is not moved.
func measureStep(a, b *Frame, w window) step {
	ia, ib := a.Image, b.Image
	if ia.Width != ib.Width || ia.Height != ib.Height || ia.West != ib.West || ia.East != ib.East ||
		ia.South != ib.South || ia.North != ib.North || ia.Mercator != ib.Mercator ||
		len(ia.Classes) != ia.Width*ia.Height || len(ib.Classes) != len(ia.Classes) {
		return step{}
	}
	hours := b.Valid.Sub(a.Valid).Hours()
	px, py, ok := ia.pixel(w.at)
	if !ok || hours <= 0 {
		return step{}
	}
	kmX, kmY := ia.pixelKm(px, py)
	if !(kmX > 0) || !(kmY > 0) {
		return step{}
	}
	side := max(1, int(math.Round(blockKm/math.Min(kmX, kmY))))
	ba, bb := a.blocksOf(side, w.threshold), b.blocksOf(side, w.threshold)
	cx, cy := px/side, py/side
	rx, ry := int(flowRadiusKm/(kmX*float64(side))), int(flowRadiusKm/(kmY*float64(side)))
	var hot []int
	total := int32(0)
	for y := max(0, cy-ry); y <= min(ba.rows-1, cy+ry); y++ {
		for x := max(0, cx-rx); x <= min(ba.cols-1, cx+rx); x++ {
			if n := ba.count[y*ba.cols+x]; n > 0 {
				hot, total = append(hot, y*ba.cols+x), total+n
			}
		}
	}
	if total < leastHot {
		return step{}
	}
	reach := min(int(math.Ceil(cellSpeedKmh*hours/(math.Min(kmX, kmY)*float64(side)))), widestShift)
	fit := func(dx, dy int) float64 {
		n := int64(0)
		for _, at := range hot {
			x, y := at%ba.cols+dx, at/ba.cols+dy
			if x >= 0 && y >= 0 && x < ba.cols && y < ba.rows {
				n += int64(ba.count[at]) * int64(bb.count[y*ba.cols+x])
			}
		}
		return float64(n)
	}
	// The fit is not normalised: of two shifts that meet the same rain, the
	// one landing on denser rain in the next frame wins. The window stays
	// where the newest frame's rain is, for every pair.
	best, bx, by := -1.0, 0, 0
	for dy := -reach; dy <= reach; dy++ {
		for dx := -reach; dx <= reach; dx++ {
			n := fit(dx, dy)
			if n > best || (n == best && dx*dx+dy*dy < bx*bx+by*by) {
				best, bx, by = n, dx, dy
			}
		}
	}
	if best <= 0 {
		return step{} // nothing met at any shift: the pair says nothing about motion
	}
	east := (float64(bx) + vertex(fit(bx-1, by), best, fit(bx+1, by))) * kmX * float64(side)
	south := (float64(by) + vertex(fit(bx, by-1), best, fit(bx, by+1))) * kmY * float64(side)
	return step{east: east, north: -south, ok: true} // rows run from the north
}

// vertex is where, between -0.5 and 0.5 of a block, the parabola through
// three fits a block apart peaks; zero where they make no peak.
func vertex(before, at, after float64) float64 {
	curve := before - 2*at + after
	if curve >= 0 {
		return 0
	}
	return math.Max(-0.5, math.Min(0.5, (before-after)/(2*curve)))
}

// nearestHot is the middle of the pixel at or above the threshold nearest a
// place, which may lie outside the image. Pixels are measured flat at the
// place's own scale, close enough to choose the nearest at a loop's size.
func (i Image) nearestHot(threshold int, near project.LonLat) (project.LonLat, bool) {
	if len(i.Classes) != i.Width*i.Height || i.Width < 2 || i.Height < 2 {
		return project.LonLat{}, false
	}
	kmX, kmY := i.pixelKm(i.Width/2, i.Height/2)
	px, py := i.pixelFloat(near)
	best, bx, by := math.Inf(1), -1, -1
	for y := range i.Height {
		for x := range i.Width {
			if int(i.Classes[y*i.Width+x]) < threshold {
				continue
			}
			dx, dy := (float64(x)+0.5-px)*kmX, (float64(y)+0.5-py)*kmY
			if d := dx*dx + dy*dy; d < best {
				best, bx, by = d, x, y
			}
		}
	}
	if bx < 0 {
		return project.LonLat{}, false
	}
	return i.centre(bx, by), true
}

// pixelFloat is where a place falls in pixels, fractional and unbounded: a
// place off the image has a position past its edge.
func (i Image) pixelFloat(at project.LonLat) (float64, float64) {
	span := eastward(i.West, i.East)
	if span <= 0 {
		span += 360
	}
	// From the image's middle, so a place on an image wider than half the
	// world is measured the short way to the side it lies on.
	x := (span/2 + eastward(i.West+span/2, at.Lon)) / span * float64(i.Width)
	down := (i.North - at.Lat) / (i.North - i.South)
	if i.Mercator {
		down = (webHeight(i.North) - webHeight(at.Lat)) / (webHeight(i.North) - webHeight(i.South))
	}
	return x, down * float64(i.Height)
}

// pixelKm is how wide and how tall one pixel is, in kilometres, at a pixel.
func (i Image) pixelKm(x, y int) (float64, float64) {
	nx, ny := x+1, y+1
	if nx >= i.Width {
		nx = x - 1
	}
	if ny >= i.Height {
		ny = y - 1
	}
	if nx < 0 || ny < 0 {
		return 0, 0
	}
	wide, err := project.GreatCircleKm(i.centre(x, y), i.centre(nx, y))
	if err != nil {
		return 0, 0
	}
	tall, err := project.GreatCircleKm(i.centre(x, y), i.centre(x, ny))
	if err != nil {
		return 0, 0
	}
	return wide, tall
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

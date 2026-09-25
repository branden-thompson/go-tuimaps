package tuimaps

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// The zooms a map may be at. The closest is upstream's (P-55). The furthest
// out is past the world in one tile: a rectangle taller than the world
// needs a zoom below zero to hold it, which is what fitting the world to a
// wide, short map gives (P-54, P-56).
const (
	MinZoom = project.MinViewZoom
	MaxZoom = project.MaxViewZoom
)

func badView(why, todo textsafe.Text) error {
	return fault.Make(fault.InvalidCoordinates, textsafe.Const("the view was not moved"), why, todo)
}

// Centre is where the map is centred, and at what zoom.
func (m *Map) Centre() (LonLat, float64) {
	defer m.guardQuiet("Centre")
	m.plant("Centre")

	if m == nil {
		return LonLat{}, 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.view.Centre, m.view.Zoom
}

// Recentre puts a place at the middle of the map (FR-24).
func (m *Map) Recentre(at LonLat) (err error) {
	defer guard("Recentre", &err)
	m.plant("Recentre")

	if math.IsNaN(at.Lat) || math.IsNaN(at.Lon) || at.Lat < -90 || at.Lat > 90 || at.Lon < -180 || at.Lon > 180 {
		return badView(textsafe.Const("that place is not on the world"),
			textsafe.Const("latitude runs from -90 to 90 and longitude from -180 to 180"))
	}
	return m.move(func(v *project.View) { v.Centre = at })
}

// DeepestZoom is the deepest zoom this map can draw real detail at: the
// source's own limit, or the embedded tiles' when there is no source. Zoom
// accepts anything up to MaxZoom, and beyond this the map is drawn from
// ancestors standing in - which is a picture, but not a sharper one.
//
// **A host cannot work this out for itself.** It is a property of the tiles
// the map was given, not of the assets package a host imported (task 14.19).
func (m *Map) DeepestZoom() float64 {
	defer m.guardQuiet("DeepestZoom")
	m.plant("DeepestZoom")

	if m == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut || m.pipe == nil {
		return 0
	}
	return float64(m.pipe.Deepest())
}

// Zoom sets the zoom, holding the centre where it is.
func (m *Map) Zoom(to float64) (err error) {
	defer guard("Zoom", &err)
	m.plant("Zoom")

	if math.IsNaN(to) || to < MinZoom || to > MaxZoom {
		return badView(textsafe.Const("that is not a zoom the map has"),
			textsafe.Const("zoom runs from -8, further out than the whole world on a wide map, to 18, the closest"))
	}
	return m.move(func(v *project.View) { v.Zoom = to })
}

// ZoomBy zooms in or out by a number of levels, stopping at the ends.
func (m *Map) ZoomBy(levels float64) (err error) {
	defer guard("ZoomBy", &err)
	m.plant("ZoomBy")

	if math.IsNaN(levels) {
		return badView(textsafe.Const("that is not a number of levels"), textsafe.Const("pass how many levels to zoom, positive to zoom in"))
	}
	return m.move(func(v *project.View) { v.Zoom = min(max(v.Zoom+levels, MinZoom), MaxZoom) })
}

// PanCells moves the map by whole cells, which is what a key press means.
func (m *Map) PanCells(cols, rows int) (err error) {
	defer guard("PanCells", &err)
	m.plant("PanCells")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if !m.sized {
		return noSize()
	}
	at, err := m.view.FromDot(float64(m.view.Cols+cols*2), float64(m.view.Rows*2+rows*4))
	if err != nil {
		return badView(textsafe.Const("the map cannot be moved that far"), textsafe.Const("pan by fewer cells, or zoom out first"))
	}
	return m.moveLocked(func(v *project.View) { v.Centre = at })
}

// FitWorld puts the whole world in the map (P-56).
func (m *Map) FitWorld() (err error) {
	defer guard("FitWorld", &err)
	m.plant("FitWorld")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if !m.sized {
		return noSize()
	}
	whole, err := project.WholeWorld(m.view.Cols, m.view.Rows)
	if err != nil {
		return noSize()
	}
	return m.moveLocked(func(v *project.View) { v.Centre, v.Zoom = whole.Centre, whole.Zoom })
}

// FitTo frames the places and overlays named, with a margin in cells on
// every side, at the largest zoom that holds them all (FR-24). The places
// are the host's own and any given here; the overlays are named by id, and
// **it needs no Work to have run**, because where each overlay is was
// recorded when it was handed in (D-76).
func (m *Map) FitTo(places []LonLat, overlays []string, margin int) (err error) {
	defer guard("FitTo", &err)
	m.plant("FitTo")

	if m == nil {
		return closed()
	}
	if margin < 0 {
		return badView(textsafe.Const("the margin is below zero"), textsafe.Const("give a margin of cells to leave on each side, or none"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if !m.sized {
		return noSize()
	}
	points := make([]LonLat, 0, len(places)+len(m.places))
	points = append(points, places...)
	for _, p := range m.places {
		points = append(points, p.At)
	}
	boxes := make([]project.Box, 0, len(overlays))
	for _, id := range overlays {
		box, ok := m.store.Box(id)
		if !ok {
			return fault.Make(fault.InvalidID, textsafe.Const("the view was not moved"),
				textsafe.Const("one of the overlays named is not on the map, or has nowhere to fit to"),
				textsafe.Const("name overlays the map holds; Overlays() lists them"))
		}
		boxes = append(boxes, box)
	}
	if len(points) == 0 && len(boxes) == 0 {
		return badView(textsafe.Const("nothing was named to fit to, and the map has no places of its own"),
			textsafe.Const("name places or overlays, or set the map's places first"))
	}
	fitted, err := project.Frame(m.view, points, boxes, margin)
	if err != nil {
		return badView(textsafe.Const("what was named cannot be framed"), textsafe.Const("check the places and overlays named are on the world"))
	}
	return m.moveLocked(func(v *project.View) { v.Centre, v.Zoom = fitted.Centre, fitted.Zoom })
}

// move applies one change to the view under the map's lock.
func (m *Map) move(change func(*project.View)) error {
	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if !m.sized {
		return noSize()
	}
	return m.moveLocked(change)
}

// moveLocked applies one change with the lock already held, and refuses a
// change that would leave the view invalid, keeping the view it had.
func (m *Map) moveLocked(change func(*project.View)) error {
	was := m.view
	change(&m.view)
	m.view = m.boundedLocked(m.view) // the host's bound, on every move (L-3.1)
	if err := m.view.Validate(); err != nil {
		m.view = was
		return badView(textsafe.Const("the map would not be a map of anywhere"), textsafe.Const("move it a shorter way, or zoom out first"))
	}
	// The host has said where to look, so a later change of size keeps the
	// place it chose rather than refitting the world (task 14.19).
	m.placed = true
	if m.view != was {
		m.changed++
	}
	return nil
}

// Bound is where a host keeps its map (L-3.1): a least zoom, and a box it
// never looks outside of. West may be east of East: the box then crosses the
// antimeridian, as Alaska's and the Pacific's do. The zero Bound is none.
type Bound struct {
	MinZoom    float64
	W, S, E, N float64
}

// SetBound sets the map's bound, and moves the view into it at once. It is
// held on every path that moves the view - a pan, a zoom, a fit, a change of
// size and the fall-back to the whole world - and is called again whenever
// the host's region changes. The zero Bound takes it away.
func (m *Map) SetBound(b Bound) (err error) {
	defer guard("SetBound", &err)
	m.plant("SetBound")

	if m == nil {
		return closed()
	}
	if b != (Bound{}) {
		ok := b.MinZoom >= 0 && b.MinZoom <= project.MaxViewZoom &&
			b.W >= -180 && b.W <= 180 && b.E >= -180 && b.E <= 180 && b.W != b.E &&
			b.S >= -project.MaxLatitude && b.N <= project.MaxLatitude && b.S < b.N
		if !ok {
			return badView(textsafe.Const("the bound is not a box on the map, or its least zoom is not a zoom"),
				textsafe.Const("give west and east from -180 to 180 and not equal, south below north, and a least zoom the map can draw"))
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	if b == (Bound{}) {
		m.bound = nil
	} else {
		m.bound = &b
	}
	if m.sized {
		if held := m.boundedLocked(m.view); held != m.view {
			m.view = held
		}
	}
	m.changed++
	return nil
}

// boundedLocked is a view held inside the map's bound: its zoom raised to the
// least, and each axis moved into the box - or, where the box is narrower
// than the view, centred on it. It works in the world's projected fractions,
// where a box across the antimeridian is one unbroken span.
func (m *Map) boundedLocked(v project.View) project.View {
	b := m.bound
	if b == nil {
		return v
	}
	v.Zoom = math.Max(v.Zoom, b.MinZoom)
	x0, y0, err0 := project.ToTile(project.LonLat{Lon: b.W, Lat: b.N}, 0)
	x1, y1, err1 := project.ToTile(project.LonLat{Lon: b.E, Lat: b.S}, 0)
	cx, cy, err2 := project.ToTile(v.Centre, 0)
	if err0 != nil || err1 != nil || err2 != nil {
		return v
	}
	if b.W > b.E {
		x1++ // across the antimeridian: the box runs on past the world's edge
	}
	// Of the centre's copies a world apart, the one nearest the box: a
	// centre just west of a box that crosses the antimeridian goes to its
	// west end, not round the world to its east end.
	best := cx
	for _, c := range []float64{cx - 1, cx + 1} {
		if gap(c, x0, x1) < gap(best, x0, x1) {
			best = c
		}
	}
	cx = best
	world := project.TileSize * math.Exp2(v.Zoom)
	hw := float64(v.Cols*project.DotsPerCol) / 2 / world
	hh := float64(v.Rows*project.DotsPerRow) / 2 / world
	cx, cy = within(cx, x0, x1, hw), within(cy, y0, y1, hh)
	at, err := project.FromTile(cx-math.Floor(cx), cy, 0)
	if err != nil {
		return v
	}
	v.Centre = at
	return v
}

// gap is how far a value lies outside a span, zero inside it.
func gap(c, lo, hi float64) float64 {
	return math.Max(0, math.Max(lo-c, c-hi))
}

// within holds a centre so that a half-width either side of it stays in a
// span, or centres it on a span narrower than that.
func within(c, lo, hi, half float64) float64 {
	if hi-lo <= 2*half {
		return (lo + hi) / 2
	}
	return math.Max(lo+half, math.Min(hi-half, c))
}

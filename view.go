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

// Zoom sets the zoom, holding the centre where it is.
func (m *Map) Zoom(to float64) (err error) {
	defer guard("Zoom", &err)
	m.plant("Zoom")

	if math.IsNaN(to) || to < MinZoom || to > MaxZoom {
		return badView(textsafe.Const("that is not a zoom the map has"),
			textsafe.Const("zoom runs from 0, the whole world, to 20"))
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
	if err := m.view.Validate(); err != nil {
		m.view = was
		return badView(textsafe.Const("the map would not be a map of anywhere"), textsafe.Const("move it a shorter way, or zoom out first"))
	}
	if m.view != was {
		m.changed++
	}
	return nil
}

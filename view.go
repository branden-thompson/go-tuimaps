package tuimaps

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// The zooms a map may be at: the world in one tile, to the deepest tile any
// source has (P-54, P-55).
const (
	MinZoom = 0.0
	MaxZoom = 20.0
)

func badView(why, todo textsafe.Text) error {
	return fault.Make(fault.InvalidCoordinates, textsafe.Const("the view was not moved"), why, todo)
}

// Centre is where the map is centred, and at what zoom.
func (m *Map) Centre() (LonLat, float64) {
	if m == nil {
		return LonLat{}, 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.view.Centre, m.view.Zoom
}

// Recentre puts a place at the middle of the map (FR-24).
func (m *Map) Recentre(at LonLat) error {
	if math.IsNaN(at.Lat) || math.IsNaN(at.Lon) || at.Lat < -90 || at.Lat > 90 || at.Lon < -180 || at.Lon > 180 {
		return badView(textsafe.Const("that place is not on the world"),
			textsafe.Const("latitude runs from -90 to 90 and longitude from -180 to 180"))
	}
	return m.move(func(v *project.View) { v.Centre = at })
}

// Zoom sets the zoom, holding the centre where it is.
func (m *Map) Zoom(to float64) error {
	if math.IsNaN(to) || to < MinZoom || to > MaxZoom {
		return badView(textsafe.Const("that is not a zoom the map has"),
			textsafe.Const("zoom runs from 0, the whole world, to 20"))
	}
	return m.move(func(v *project.View) { v.Zoom = to })
}

// ZoomBy zooms in or out by a number of levels, stopping at the ends.
func (m *Map) ZoomBy(levels float64) error {
	if math.IsNaN(levels) {
		return badView(textsafe.Const("that is not a number of levels"), textsafe.Const("pass how many levels to zoom, positive to zoom in"))
	}
	return m.move(func(v *project.View) { v.Zoom = min(max(v.Zoom+levels, MinZoom), MaxZoom) })
}

// PanCells moves the map by whole cells, which is what a key press means.
func (m *Map) PanCells(cols, rows int) error {
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
func (m *Map) FitWorld() error {
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

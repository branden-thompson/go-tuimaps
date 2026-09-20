package tuimaps

import (
	"slices"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/render"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// Overlay is what a host hands in: an id, when its data was valid and how
// long it stays current, a credit, and its shapes - features, or a grid, or
// an image, and one only (FR-6).
type Overlay = overlay.Overlay

// Feature is one shape of a feature overlay. Its rings are the host's own
// memory, borrowed and never copied (FR-11).
type Feature = overlay.Feature

// FeatureKind is what a feature is.
type FeatureKind = overlay.FeatureKind

// The kinds of feature.
const (
	Point   = overlay.Point
	Line    = overlay.Line
	Polygon = overlay.Polygon
	Circle  = overlay.Circle
)

// Grid is a scalar field on the host's own regular grid of longitude and
// latitude, its values rows from the north, each row west to east.
type Grid = overlay.Grid

// Type is how a host says what a grid's values are: a preset by name, or
// breaks of its own (D-69).
type Type = overlay.Type

// Image is a georeferenced image and the table that says what its colours
// mean (D-45).
type Image = overlay.Image

// TableEntry is one row of that table: a colour of the provider's, the
// value it stands for, and whether it means "no data".
type TableEntry = overlay.TableEntry

// Projection is how an image's pixels are laid on the world.
type Projection = overlay.Projection

// The projections an image may be handed in.
const (
	PlateCarree = overlay.PlateCarree // equal steps of longitude and latitude
	WebMercator = overlay.WebMercator // the projection the tiles are in
)

// Unit is a unit of temperature.
type Unit = colour.Unit

// The units a temperature grid may be handed in.
const (
	Celsius    = colour.Celsius
	Fahrenheit = colour.Fahrenheit
)

// Token is the role a feature is drawn in: its colour, and for an alert its
// severity.
type Token = colour.Token

// The roles a feature overlay may carry. An alert's area is filled with the
// tint that belongs to its outline.
const (
	AlertExtreme  = colour.AlertExtremeOutline
	AlertSevere   = colour.AlertSevereOutline
	AlertModerate = colour.AlertModerateOutline
	AlertMinor    = colour.AlertMinorOutline
	AlertUnknown  = colour.AlertUnknownOutline
	Track         = colour.Track
	Low           = colour.Low
	Middle        = colour.Middle
	High          = colour.High
)

// SetResult is what Set says at once (D-74, D-86).
type SetResult = overlay.SetResult

// RemoveResult is what Remove says at once.
type RemoveResult = overlay.RemoveResult

// Warning is something the library noticed and did not refuse.
type Warning = fault.Warning

// TemperatureGrid is a temperature field in one call: the preset supplies
// the breaks and the colours, and the host says only which unit its values
// are in (D-69). Everything it sets can be set by hand instead.
func TemperatureGrid(id string, grid Grid, unit Unit, validAt time.Time) Overlay {
	grid.Type = Type{Preset: "temperature", Unit: unitName(unit)}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &grid}
}

// RadarImage is a radar image in one call: the host gives the picture, the
// table that says what its colours mean, and when it was valid (D-45, D-69).
func RadarImage(id string, image Image, validAt time.Time) Overlay {
	image.Type = Type{Preset: "radar", Unit: "dBZ"}
	return Overlay{ID: id, Valid: validAt, Keeps: 15 * time.Minute, Image: &image}
}

// unitName is the name the overlay store knows a unit by.
func unitName(u Unit) string {
	if u == Fahrenheit {
		return "F"
	}
	return "C"
}

// Set hands an overlay to the map, or replaces the one of that id. It never
// waits: what it answers says whether the geometry the overlay replaced is
// free for the host to write over again (D-86).
func (m *Map) Set(o Overlay) (res SetResult, err error) {
	defer guard("Set", &err)
	m.plant("Set")

	if m == nil {
		return SetResult{}, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return SetResult{}, closed()
	}
	res, err = m.store.HandIn(o)
	if err != nil {
		return res, err
	}
	m.overlays++
	m.described = nil
	m.changed++
	return res, nil
}

// Remove takes an overlay away, and says whether it was there and whether
// its geometry is free at once.
func (m *Map) Remove(id string) (res RemoveResult, err error) {
	defer guard("Remove", &err)
	m.plant("Remove")

	if m == nil {
		return RemoveResult{}, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return RemoveResult{}, closed()
	}
	res, err = m.store.Drop(id)
	if err != nil || !res.Found {
		return res, err
	}
	m.overlays++
	m.described = nil
	m.changed++
	return res, nil
}

// InUse reports whether anything is still reading an overlay's old geometry.
// While it is true the host must leave that memory alone (FR-11, D-86).
func (m *Map) InUse(id string) bool {
	defer m.guardQuiet("InUse")
	m.plant("InUse")

	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.shut && m.store.Reading(id)
}

// Overlays are the ids the map holds, sorted.
func (m *Map) Overlays() []string {
	defer m.guardQuiet("Overlays")
	m.plant("Overlays")

	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil
	}
	return m.store.IDs()
}

// Warnings are what the library noticed and did not refuse, taken once: a
// host that never asks is never blocked by them (NFR-20).
func (m *Map) Warnings() []Warning {
	defer m.guardQuiet("Warnings")
	m.plant("Warnings")

	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil
	}
	mine := m.own
	m.own = nil
	return slices.Concat(mine, m.store.TakeWarnings(), m.pipe.TakeWarnings())
}

// bucket is the zoom bucket the view draws overlays at.
func (m *Map) bucket() int {
	return overlay.Bucket(m.view.Zoom)
}

// overlayWork adds the jobs that prepare what the view wants to draw and has
// not got: one job an overlay, at the bucket in view.
func (m *Map) overlayWork() error {
	bucket := m.bucket()
	for _, id := range m.store.IDs() {
		_, _, path := m.store.Drawn(id, bucket)
		if path == overlay.Cached {
			continue
		}
		if _, ok := m.store.Field(id); ok {
			continue
		}
		if _, _, ok := m.store.Raster(id); ok {
			continue
		}
		err := m.member.Add(m.store.PrepareJob(id, bucket))
		if err != nil {
			return err
		}
	}
	return nil
}

// draw fills in what the overlays put on this frame: prepared shapes, fields
// and images, and the ones read straight from the host's memory (D-92).
func (m *Map) draw(in *render.Input) {
	bucket := m.bucket()
	m.shapes, m.fields, m.rasters, m.borrowed = m.shapes[:0], m.fields[:0], m.rasters[:0], m.borrowed[:0]
	for _, id := range m.store.IDs() {
		if field, ok := m.store.Field(id); ok {
			m.fields = append(m.fields, field)
			continue
		}
		if raster, _, ok := m.store.Raster(id); ok {
			m.rasters = append(m.rasters, raster)
			continue
		}
		shapes, _, path := m.store.Drawn(id, bucket)
		switch path {
		case overlay.Cached:
			m.shapes = append(m.shapes, shapes...)
		case overlay.FromMemory:
			m.borrow(id)
		}
	}
	in.Shapes, in.Fields, in.Rasters, in.Borrowed = m.shapes, m.fields, m.rasters, m.borrowed
	in.OverlaysVersion = m.overlays
}

// borrow reads one overlay where it lies, through its run index: the frame
// draws it without the library ever holding a copy (D-92).
func (m *Map) borrow(id string) {
	reader, ok := m.store.Read(id)
	if !ok {
		return
	}
	defer reader.Done()
	index := reader.Index()
	at := 0
	for _, f := range reader.Overlay().Features {
		runs := 0
		for _, ring := range f.Rings {
			runs += (len(ring) + render.RunLength - 1) / render.RunLength
		}
		lent := render.Borrowed{Kind: shapeKind(f.Kind), Rings: f.Rings, Role: uint8(f.Role), Label: f.Label}
		if at+runs <= len(index) {
			lent.Index = index[at : at+runs]
		}
		m.borrowed = append(m.borrowed, lent)
		at += runs
	}
}

// shapeKind is how a feature is drawn.
func shapeKind(k FeatureKind) scene.ShapeKind {
	switch k {
	case Point:
		return scene.ShapePoint
	case Polygon, Circle:
		return scene.ShapeArea
	}
	return scene.ShapeLine
}

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

// Feature is ONE AREA of a feature overlay: an outline first, then any holes
// in it. Ground that is separate is a separate feature, because every ring
// after the first is read as a hole. Its rings are read where they are and not
// copied again (FR-11); a host holding its own point type converts once with
// Rings.
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

// LoopFrame is one frame of an image's loop: when its picture was valid, and
// the picture. A gap is a missing frame, stated as missing, with no bytes; a
// forecast frame is one the provider forecast rather than observed.
type LoopFrame = overlay.LoopFrame

// Span is when an overlay is drawn: while the map's moment meets it - the
// loop's frame, or the host's moment (ShowMoment), or the frame's clock. A
// zero end is open (L-15.1).
type Span = overlay.Span

// MaxFrames is the most frames a loop may have, gaps included.
const MaxFrames = overlay.MaxFrames

// Severity is how severe an alert is; the zero value means the one its role
// implies.
type Severity = overlay.Severity

// The severities, least first.
const (
	SeverityUnknown  = overlay.SeverityUnknown
	SeverityMinor    = overlay.SeverityMinor
	SeverityModerate = overlay.SeverityModerate
	SeveritySevere   = overlay.SeveritySevere
	SeverityExtreme  = overlay.SeverityExtreme
)

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
	Fire          = colour.Fire       // a fire's perimeter, its incident, a strong hotspot (L-18)
	FireFaint     = colour.FireFaint  // a weaker hotspot
	QuakeHour     = colour.QuakeHour  // a quake of the past hour (L-19)
	QuakeDay      = colour.QuakeDay   // of the past day
	QuakeOlder    = colour.QuakeOlder // older
	Buoy          = colour.Buoy       // a buoy's marker and words (L-21)
	Tide          = colour.Tide       // a tide station's
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

// SpeedUnit is a unit of wind speed (FR-8).
type SpeedUnit = colour.SpeedUnit

// The units of wind speed.
const (
	MilesPerHour      = colour.MilesPerHour
	KilometresPerHour = colour.KilometresPerHour
	MetresPerSecond   = colour.MetresPerSecond
	Knots             = colour.Knots
)

// WindGrid is a wind field in one call (FR-8, L-16): the speeds in a unit,
// and the direction each blows FROM, meteorological degrees clockwise from
// north, one a speed, NaN where there is none. It is drawn as arrows, never
// bands, and the preset supplies the classes and their colours.
func WindGrid(id string, speed Grid, from []float64, unit SpeedUnit, validAt time.Time) Overlay {
	speed.From = from
	speed.Type = Type{Preset: "wind", Unit: speedUnitName(unit)}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &speed}
}

// speedUnitName is the name the overlay store knows a speed unit by.
func speedUnitName(u SpeedUnit) string {
	switch u {
	case KilometresPerHour:
		return "km/h"
	case MetresPerSecond:
		return "m/s"
	case Knots:
		return "kt"
	}
	return "mph"
}

// WaveUnit is a unit of wave height (L-20).
type WaveUnit = colour.WaveUnit

// The units of wave height.
const (
	Feet   = colour.Feet
	Metres = colour.Metres
)

// WaveGrid is a wave-height field in one call (L-20): the preset supplies
// the classes and their colours, and the field is drawn over the sea alone.
func WaveGrid(id string, grid Grid, unit WaveUnit, validAt time.Time) Overlay {
	name := "ft"
	if unit == Metres {
		name = "m"
	}
	grid.Type = Type{Preset: "waves", Unit: name}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &grid}
}

// UVGrid is a UV index field in one call (L-25): the preset supplies its
// five categories and their colours.
func UVGrid(id string, grid Grid, validAt time.Time) Overlay {
	grid.Type = Type{Preset: "uv", Unit: "index"}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &grid}
}

// AirQualityGrid is a US AQI field in one call (L-25): the preset supplies
// its six categories and their colours.
func AirQualityGrid(id string, grid Grid, validAt time.Time) Overlay {
	grid.Type = Type{Preset: "aqi", Unit: "AQI"}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &grid}
}

// QPFGrid is a field of rain and snow totals, liquid-equivalent, in mm, in
// one call (L-26): the preset supplies WPC's seven classes and their
// colours, and draws nothing under a trace.
func QPFGrid(id string, grid Grid, validAt time.Time) Overlay {
	grid.Type = Type{Preset: "qpf", Unit: "mm"}
	return Overlay{ID: id, Valid: validAt, Keeps: time.Hour, Grid: &grid}
}

// AirQualityRole is the role a feature is drawn in to show an AQI - a
// monitor's reading - in its category's colour, as the AQI preset draws it
// (L-25). Its words are drawn in the markers' ink, which reads on the ground.
func AirQualityRole(aqi float64) Token {
	c := 0
	for _, b := range colour.AirQualityBreaks() {
		if aqi >= b {
			c++
		}
	}
	return colour.AQI1 + Token(c)
}

// UVRole is the role a feature is drawn in to show a UV index - a city's
// reading - in its band's colour, as the UV preset draws it (watchpost W18.4,
// its D-167: EPA's UV index as markers). Its words are drawn in the markers'
// ink, as AirQualityRole's.
func UVRole(index float64) Token {
	c := 0
	for _, b := range colour.UVBreaks() {
		if index >= b {
			c++
		}
	}
	return colour.UV1 + Token(c)
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
	if o.Image != nil {
		if err := overlay.CheckFrameTimes(o.Image, m.wallClock); err != nil {
			return SetResult{}, err
		}
	}
	res, err = m.store.HandIn(o)
	if err != nil {
		return res, err
	}
	m.overlays++
	m.reported = nil
	m.play.seen = m.shownLocked() // a refresh that moves the moment is an input, not an advance
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
	m.reported = nil
	m.play.seen = m.shownLocked() // a refresh that moves the moment is an input, not an advance
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

// stillWanted reports whether a render's plan keeps the work under key: a
// tile the view needs, or the preparation, at the bucket in view, of an
// overlay the map holds (watchpost UAT-2 U2-59). A preparation withdrawn here
// is cancelled where it runs, so a loop that takes longer to prepare than the
// gap between two renders would never be drawn. It allocates nothing: a
// render pays it for every job waiting or running.
func (m *Map) stillWanted(key string) bool {
	if m.pipe.StillWanted(key) {
		return true
	}
	id, bucket, ok := overlay.PreparesOf(key)
	return ok && bucket == m.bucket() && m.store.Holds(id)
}

// draw fills in what the overlays put on this frame: prepared shapes, fields
// and images, and the ones read straight from the host's memory (D-92).
func (m *Map) draw(in *render.Input) {
	bucket := m.bucket()
	shown := m.shownLocked()
	if !shown.Equal(m.play.drawn) {
		m.play.moved++ // another frame of a loop is drawn: the last frame cannot be reused
		m.play.drawn = shown
	}
	m.shapes, m.reserved, m.fields, m.rasters, m.borrowed = m.shapes[:0], m.reserved[:0], m.fields[:0], m.rasters[:0], m.borrowed[:0]
	from, to := m.momentLocked(shown)
	hidden := uint64(0)
	for _, id := range m.store.IDs() {
		// AN OVERLAY OUTSIDE THE MOMENT IS NOT DRAWN (L-15.1), but it stays
		// prepared: the frame that meets it draws it at once, never a frame late.
		if span, ok := m.store.During(id); ok && !span.Meets(from, to) {
			hidden = hidden*1099511628211 ^ idHash(id)
			m.reserve(in, id, bucket)
			continue
		}
		if field, ok := m.store.Field(id); ok {
			m.fields = append(m.fields, field)
			continue
		}
		if raster, _, ok := m.store.RasterAt(id, shown); ok {
			m.rasters = append(m.rasters, raster)
			in.ImageHeld = true
			continue
		}
		if _, _, ok := m.store.Raster(id); ok {
			in.ImageHeld = true // a loop on a gap still shares the map: the field's look holds across it (L-15.3)
			continue
		}
		shapes, _, path := m.store.Drawn(id, bucket)
		switch path {
		case overlay.Cached, overlay.StandIn: // a stand-in is drawn while its replacement is prepared (L11.5)
			m.shapes = append(m.shapes, shapes...)
		case overlay.FromMemory:
			m.borrow(id)
		}
	}
	if hidden != m.play.hidden {
		m.play.moved++ // another set of overlays is drawn: the last frame cannot be reused
		m.play.hidden = hidden
	}
	in.Shapes, in.Reserved, in.Fields, in.Rasters, in.Borrowed = m.shapes, m.reserved, m.fields, m.rasters, m.borrowed
	in.OverlaysVersion = m.overlays + m.play.moved // the description's key reads only the first (L-1.10e)
}

// reserve keeps what an overlay outside the moment would take from the names
// (L-28, watchpost D-200): its prepared shapes, whose alert words and digits
// hold their room, and the cover a field or image puts on the name budget.
// The names are placed as if every overlay in view were present, so they
// stand still as a loop plays past each one's hours. Reading a prepared form
// marks it used, which keeps it from eviction until its hour comes round. An
// overlay read from the host's memory holds nothing: it is read only to draw.
func (m *Map) reserve(in *render.Input, id string, bucket int) {
	if _, ok := m.store.Field(id); ok {
		in.Covered = true
		return
	}
	if _, _, ok := m.store.Raster(id); ok {
		in.Covered = true
		return
	}
	if shapes, _, path := m.store.Drawn(id, bucket); path == overlay.Cached || path == overlay.StandIn {
		m.reserved = append(m.reserved, shapes...)
	}
}

// idHash is an overlay id's FNV-1a hash, for the set of those left out.
func idHash(id string) uint64 {
	h := uint64(14695981039346656037)
	for i := range len(id) {
		h = (h ^ uint64(id[i])) * 1099511628211
	}
	return h
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

// Provider names a source of radar images whose colour table the library
// carries: an image that names one, with no table of its own, is read with
// the provider's (L-2.5).
type Provider = overlay.Provider

// The providers the library carries tables for.
const (
	ProviderIEM  = overlay.ProviderIEM  // the Iowa Environmental Mesonet's N0Q composite, its published table
	ProviderMRMS = overlay.ProviderMRMS // NOAA's MRMS reflectivity, its observed palette valued from its legend: approximate
)

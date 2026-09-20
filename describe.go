package tuimaps

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/describe"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Answer is what the map says about one place and one overlay: data to
// render or speak, never a sentence of the library's own (D-52).
type Answer = describe.Answer

// Description is everything the map says about one place.
type Description struct {
	Place   string
	Answers []Answer
	// Pending is true for a part that is not worked out yet. **It is always
	// false**: the description is computed from the data the host handed in,
	// not from anything the library has to prepare first, so a call answers
	// at once and never waits (FR-29). It is here because the contract says
	// each part is marked, and a part that is ready says so.
	Pending bool
}

// Units sets the units the descriptions come back in: miles rather than
// kilometres, Fahrenheit rather than Celsius. Whichever is in use is always
// stated in the answer itself.
func (m *Map) Units(miles, fahrenheit bool) {
	defer m.guardQuiet("Units")
	m.plant("Units")

	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return
	}
	want := describe.Units{Miles: miles, Fahrenheit: fahrenheit}
	if m.units != want {
		m.units, m.described = want, nil
		m.changed++
	}
}

// Describe answers, for each place, where every overlay is relative to it:
// inside or outside an area and how far its edge is, the nearest point or
// line with its label, a field's value and which way it rises, an image's
// class and where it gets heavier (FR-29, D-52).
//
// It is computed from the host's own geometry, unsimplified - never from
// the drawn cells, which cannot show a distance smaller than one cell - and
// it needs no Work to have run. With no places given it describes the map's
// own.
func (m *Map) Describe(places []Place) (out []Description, err error) {
	defer guard("Describe", &err)
	m.plant("Describe")

	if m == nil {
		return nil, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return nil, closed()
	}
	asked := places
	if len(asked) == 0 {
		asked = m.places
	}
	if len(asked) == 0 {
		return nil, nil
	}
	if kept, ok := m.rememberedDescription(asked); ok {
		return kept, nil
	}
	described := make([]Description, 0, len(asked))
	for _, place := range asked {
		one, err := checkPlace(place)
		if err != nil {
			return nil, err
		}
		described = append(described, Description{Place: nameOf(one), Answers: m.answersFor(one)})
	}
	m.described, m.describedKey = described, m.describeKey(asked)
	return described, nil
}

// nameOf is what a place is called in a description: its name, or its id
// where it has no name of its own.
func nameOf(p Place) string {
	if p.Name != "" {
		return p.Name
	}
	return p.ID
}

// describeKey is what a description was computed from: it changes whenever
// the answer would (FR-29).
type describeKey struct {
	overlays uint64
	places   uint64
	units    describe.Units
	stale    bool
	asked    int
}

// describeKey is the key for this call.
func (m *Map) describeKey(asked []Place) describeKey {
	return describeKey{overlays: m.overlays, places: m.placesVersion, units: m.units, stale: m.staleNow, asked: len(asked)}
}

// rememberedDescription is the description already worked out, when nothing
// it was worked out from has changed. Repeating a call then costs nothing,
// which is what lets a host ask on every frame (FR-29).
func (m *Map) rememberedDescription(asked []Place) ([]Description, bool) {
	if m.described == nil || m.describedKey != m.describeKey(asked) {
		return nil, false
	}
	return m.described, true
}

// answersFor is what every overlay says about one place.
func (m *Map) answersFor(place Place) []Answer {
	var out []Answer
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		answer := m.answerOf(place, id, o)
		answer.Valid = o.Valid
		answer.Stale = overlay.FreshnessAt(o.Valid, o.Keeps, m.wallClock).DrawnStale()
		answer.UnderOneCell = m.underOneCell(answer)
		out = append(out, answer.Cleaned())
	}
	return out
}

// answerOf is the one answer for a place and an overlay, by its shape.
func (m *Map) answerOf(place Place, id string, o Overlay) Answer {
	switch {
	case o.Grid != nil:
		return m.fieldAnswer(place, id, o)
	case o.Image != nil:
		return m.imageAnswer(place, id)
	}
	return m.featureAnswer(place, id, o)
}

// featureAnswer is the answer for an overlay of features: the area the
// place is in or out of, or the nearest point or line.
func (m *Map) featureAnswer(place Place, id string, o Overlay) Answer {
	var rings [][]project.LonLat
	var points, lines []describe.Labelled
	for _, f := range o.Features {
		switch f.Kind {
		case Polygon, Circle:
			rings = append(rings, f.Rings...)
		case Point:
			for _, run := range f.Rings {
				points = append(points, describe.Labelled{Run: run, Label: f.Label})
			}
		default:
			for _, run := range f.Rings {
				lines = append(lines, describe.Labelled{Run: run, Label: f.Label})
			}
		}
	}
	if len(rings) > 0 {
		return describe.OfArea(nameOf(place), id, place.At, rings, m.units)
	}
	if len(points) > 0 {
		near, ok := describe.NearestPoint(place.At, points)
		return describe.OfNear(nameOf(place), id, describe.PointsForm, near, ok, m.units)
	}
	near, ok := describe.NearestLine(place.At, lines)
	return describe.OfNear(nameOf(place), id, describe.LineForm, near, ok, m.units)
}

// fieldAnswer is the answer for a scalar field: the value of the sample the
// place falls in, its band, and which way the field rises.
func (m *Map) fieldAnswer(place Place, id string, o Overlay) Answer {
	grid := describe.Grid{West: o.Grid.West, South: o.Grid.South, East: o.Grid.East, North: o.Grid.North,
		Cols: o.Grid.Cols, Rows: o.Grid.Rows, Values: o.Grid.Values}
	kind, err := overlay.ResolveType(o.Grid.Type)
	if err != nil {
		return Answer{Place: nameOf(place), Overlay: id, Form: describe.FieldForm, NoData: true}
	}
	reading := grid.At(place.At, kind.Breaks)
	return describe.OfField(nameOf(place), id, reading, o.Grid.Type.Preset == "temperature", m.units)
}

// imageAnswer is the answer for an image: the class here, and where the
// nearest heavier class is.
func (m *Map) imageAnswer(place Place, id string) Answer {
	raster, _, ok := m.store.Raster(id)
	if !ok {
		// The image is not classified yet: say there is nothing here rather
		// than an answer made up from the provider's own colours.
		return Answer{Place: nameOf(place), Overlay: id, Form: describe.ImageForm, NoData: true}
	}
	image := describe.Image{West: raster.West, South: raster.South, East: raster.East, North: raster.North,
		Width: raster.Width, Height: raster.Height, Classes: raster.Classes}
	class, here := image.ClassAt(place.At)
	return describe.OfImage(nameOf(place), id, class, here, image.NearestHeavier(place.At), m.units)
}

// underOneCell reports whether an answer's distance is smaller than one
// cell of the view: there the picture cannot settle the question and the
// words have to (D-67).
func (m *Map) underOneCell(a Answer) bool {
	if !m.sized || a.Distance <= 0 {
		return false
	}
	perCol, _, err := m.view.CellSpanKm()
	if err != nil {
		return false
	}
	km := a.Distance
	if a.Unit == "miles" {
		km *= 1.609344
	}
	return describe.UnderOneCell(km, perCol)
}

// noteWallClock keeps the time the last frame was drawn at, which is what
// staleness in a description is judged by (D-114).
func (m *Map) noteWallClock(now time.Time) {
	if now.IsZero() {
		return
	}
	stale := m.stale(now)
	if m.wallClock != now || m.staleNow != stale {
		m.described = nil // the answer would differ; it is worked out again
	}
	m.wallClock, m.staleNow = now, stale
}

package tuimaps

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/describe"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// Answer is what the map says about one place and one overlay: data to
// render or speak, never a sentence of the library's own (D-52).
type Answer = describe.Answer

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
		m.units, m.reported = want, nil
		m.changed++
	}
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
	asked    uint64
	nearby   float64      // what "nearby" means (L-13.6)
	view     project.View // which alerts are in view, and the view's centre, for motion with no place
	landed   uint64       // what work has landed: a picture decoded is a frame motion can read
}

// describeKey is the key for this call.
func (m *Map) describeKey(asked []Place) describeKey {
	return describeKey{overlays: m.overlays, places: m.placesVersion, units: m.units, stale: m.staleNow, asked: askedFingerprint(asked),
		nearby: m.nearby, view: m.view, landed: m.store.Landed()}
}

// askedFingerprint is what was asked about, not how much of it.
//
// **A host may ask about places the map does not store** - that is what the
// argument to Report is for, and a station watching several locations uses
// it for exactly that. `placesVersion` counts changes to the map's OWN places
// and says nothing about these, so keying on the count alone answered a
// question about one place with the answer about another (FR-29 requires the
// key to change whenever the answer would).
//
// What an answer depends on is each place's name and where it is, in order.
func askedFingerprint(asked []Place) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	put := func(v uint64) {
		binary.LittleEndian.PutUint64(buf[:], v)
		_, _ = h.Write(buf[:])
	}
	put(uint64(len(asked)))
	for _, p := range asked {
		_, _ = h.Write([]byte(nameOf(p)))
		put(math.Float64bits(p.At.Lon))
		put(math.Float64bits(p.At.Lat))
	}
	return h.Sum64()
}

// rememberedReport is the report already worked out, when nothing it was
// worked out from has changed. Repeating a call then costs nothing, which is
// what lets a host ask on every frame (FR-29).
func (m *Map) rememberedReport(asked []Place) (Report, bool) {
	if m.reported == nil || m.reportedKey != m.describeKey(asked) {
		return Report{}, false
	}
	return *m.reported, true
}

// answerFor is one overlay's answer for a place, with its valid time, its
// stale mark and whether the picture can settle it, cleaned.
func (m *Map) answerFor(place Place, id string, o Overlay) Answer {
	answer := m.answerOf(place, id, o)
	answer.Valid = o.Valid
	answer.Stale = m.staleOverlay(o)
	answer.UnderOneCell = m.underOneCell(answer)
	return answer.Cleaned()
}

// staleOverlay reports whether one overlay is out of date, by the rule the
// frame's own stale mark follows: a host that has given no time is told
// nothing about time (D-114). Until a frame is drawn there is no wall clock
// to judge by, and the answer carries the valid time so that a host which
// only describes can judge it against a clock of its own.
func (m *Map) staleOverlay(o Overlay) bool {
	if m.wallClock.IsZero() {
		return false
	}
	return overlay.FreshnessAt(overlay.Valid(o), o.Keeps, m.wallClock).DrawnStale()
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
	// **One entry per feature, not one list of every ring.** A feature is one
	// area - an outline and its holes - and two areas of the same hazard may
	// overlap, which even-odd over the lot would cancel (see InAnyArea).
	var areas [][][]project.LonLat
	var points, lines []describe.Labelled
	for _, f := range o.Features {
		switch f.Kind {
		case Polygon, Circle:
			areas = append(areas, f.Rings)
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
	if len(areas) > 0 {
		return describe.OfArea(nameOf(place), id, place.At, areas, m.units)
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
		m.reported = nil // the answer would differ; it is worked out again
	}
	m.wallClock, m.staleNow = now, stale
}

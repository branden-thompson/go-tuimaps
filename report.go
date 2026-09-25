package tuimaps

import (
	"math"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/describe"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Where a place is relative to an alert: inside it, outside it, or nearby -
// outside, but within the distance SetNearby sets.
type Where = describe.Where

// The answers.
const (
	Outside = describe.Outside
	Inside  = describe.Inside
	Nearby  = describe.Nearby
)

// defaultNearby is how close to an alert's edge a place outside it is
// "nearby" until the host says otherwise (L-13.6, D-43).
const defaultNearby = 10.0

// Report is what the map says beside the picture, as data (D-57): the
// alerts shown, and for each place asked about, each alert on its own and
// every other overlay's answer. Every section is present, empty rather than
// nil; every string is cleaned (FR-34); distances are in the units Units
// set. The library words nothing: a host words it, and never says "you" of a
// place (D-29).
type Report struct {
	Alerts []AlertShown
	Places []PlaceReport
	Motion []MotionReport
}

// Trend is how a cell's distance from a place went over a loop.
type Trend = describe.Trend

// The trends: the heavier rain came closer, moved away, or held.
const (
	Held   = describe.Held
	Closer = describe.Closer
	Away   = describe.Away
)

// MotionReport is observed motion (L-1.12, D-42): where the heavier rain was
// at the oldest usable frame of a loop and where it is at the newest,
// relative to a named place - or, with Place empty, to the view's centre -
// and whether it came closer, moved away or held, over the span between.
// It is data about what was seen: nothing in it can say what will happen.
type MotionReport struct {
	Overlay, Place string
	Threshold      int // the class taken as heavier rain, held for the whole loop
	From, To       Sighting
	Trend          Trend
	Span           time.Duration
}

// Sighting is where the heavier rain was seen, and when: how far from the
// place and which way, in the units Units set.
type Sighting struct {
	Valid    time.Time
	At       LonLat
	Distance float64
	Unit     string
	Bearing  float64
	Compass  string
}

// AlertShown is one alert in view (L-13.5): the overlay and feature it is,
// its label, its severity, its times, and whether it is stale by the map's
// own rule.
type AlertShown struct {
	Overlay, Feature, Label string
	Severity                Severity
	Valid, Expires          time.Time // Valid is the feature's own, or the overlay's; Expires is zero when not given
	Stale                   bool
}

// PlaceReport is everything the map says about one place: each alert on its
// own, and every other overlay's answer - an image's class, a field's
// value, the nearest point or line, an area that is no alert.
type PlaceReport struct {
	Place   string
	Alerts  []PlaceAlert
	Answers []Answer
}

// PlaceAlert is one alert against one place (L-13.6): inside, nearby or
// outside, how far its edge is and which way, with its label and severity;
// when its data was valid and whether it is stale, by the map's own rule;
// and whether the distance is under one cell, where the picture cannot
// settle it and the words must (D-67).
type PlaceAlert struct {
	Overlay, Feature, Label string
	Severity                Severity
	Where                   Where
	Distance                float64
	Unit                    string
	Bearing                 float64
	Compass                 string
	Valid                   time.Time
	Stale, UnderOneCell     bool
}

// SetNearby sets how close to an alert's edge a place outside it is said to
// be nearby, in kilometres; zero puts the default of 10 km back (L-13.6).
func (m *Map) SetNearby(km float64) (err error) {
	defer guard("SetNearby", &err)
	m.plant("SetNearby")

	if m == nil {
		return closed()
	}
	if !(km >= 0 && km <= 1000) {
		return fault.Make(fault.OverLimit, textsafe.Const("the nearby distance was not set"),
			textsafe.Const("it is not a distance from 0 to 1,000 kilometres"), textsafe.Const("pass a distance in kilometres, or zero for the default of 10"))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	m.nearby = km
	m.changed++
	return nil
}

// Report answers, beside the picture: the alerts in view, and for each place
// - the ones given, or the map's own - each alert on its own and every other
// overlay's answer. It is computed from the host's geometry, unsimplified,
// never from the drawn cells, which cannot show a distance smaller than one
// cell, and it needs no Work to have run. Asked again with nothing changed,
// it costs nothing, so a host may ask on every frame (FR-29); the slices it
// returns are the map's until the answer changes.
func (m *Map) Report(places []Place) (out Report, err error) {
	defer guard("Report", &err)
	m.plant("Report")

	if m == nil {
		return Report{}, closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return Report{}, closed()
	}
	asked := places
	if len(asked) == 0 {
		asked = m.places
	}
	if kept, ok := m.rememberedReport(asked); ok {
		return kept, nil
	}
	out = Report{Alerts: []AlertShown{}, Places: []PlaceReport{}, Motion: []MotionReport{}}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		stale := m.staleOverlay(o)
		for _, alert := range alertsOf(o.Features) {
			if !m.anyInView(alert) {
				continue
			}
			f := alert[0]
			valid := f.Valid
			if valid.IsZero() {
				valid = o.Valid
			}
			out.Alerts = append(out.Alerts, AlertShown{Overlay: clean(id), Feature: clean(f.ID), Label: clean(f.Label),
				Severity: overlay.SeverityOf(f), Valid: valid, Expires: f.Expires, Stale: stale})
		}
	}
	for _, place := range asked {
		one, err := checkPlace(place)
		if err != nil {
			return Report{}, err
		}
		out.Places = append(out.Places, m.placeReport(one))
	}
	out.Motion = m.observedMotion(asked, out.Motion)
	kept := out // a copy for the memo: keeping out's own address would put every call's result on the heap
	m.reported, m.reportedKey = &kept, m.describeKey(asked)
	return out, nil
}

// observedMotion is each loop's observed motion relative to each place asked about,
// or with none, to the view's centre (D-42).
func (m *Map) observedMotion(asked []Place, out []MotionReport) []MotionReport {
	refs := make([]Place, 0, len(asked))
	for _, p := range asked {
		refs = append(refs, Place{Name: nameOf(p), At: p.At})
	}
	if len(refs) == 0 {
		refs = append(refs, Place{At: LonLat{Lon: m.view.Centre.Lon, Lat: m.view.Centre.Lat}})
	}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		if o.Image == nil || len(o.Image.Frames) == 0 {
			continue
		}
		kind, err := overlay.ResolveType(o.Image.Type)
		if err != nil {
			continue
		}
		threshold := heavierClass(kind)
		var frames []describe.Frame
		for _, f := range m.store.ObservedFrames(id) {
			r := f.Raster
			frames = append(frames, describe.Frame{Valid: f.Valid, Image: describe.Image{West: r.West, South: r.South, East: r.East, North: r.North,
				Width: r.Width, Height: r.Height, Classes: r.Classes}})
		}
		for _, ref := range refs {
			from, to, ok := describe.Track(frames, threshold, ref.At)
			if !ok {
				continue
			}
			a, b := m.sighting(ref.At, from), m.sighting(ref.At, to)
			fromKm, _, _, _ := describe.Measure(ref.At, from.At)
			toKm, _, _, _ := describe.Measure(ref.At, to.At)
			out = append(out, MotionReport{Overlay: clean(id), Place: clean(ref.Name), Threshold: threshold,
				From: a, To: b, Trend: describe.TrendOf(fromKm, toKm), Span: to.Valid.Sub(from.Valid)})
		}
	}
	return out
}

// sighting is a cell seen from a place, in the host's units.
func (m *Map) sighting(from LonLat, s describe.Sighting) Sighting {
	km, bearing, compass, _ := describe.Measure(from, s.At)
	d, unit := m.units.Distance(km)
	return Sighting{Valid: s.Valid, At: s.At, Distance: d, Unit: unit, Bearing: bearing, Compass: compass}
}

// heavierClass is the one class a loop takes as heavier rain, held for the
// whole loop (L-1.12): for radar, the class from 40 dBZ, where rain is heavy;
// for any other type, the top third of its classes.
func heavierClass(kind overlay.Kind) int {
	if kind.Preset == colour.Radar {
		for i, b := range kind.Breaks {
			if b >= 40 {
				return i + 1
			}
		}
	}
	return (2*(len(kind.Breaks)+1) + 2) / 3
}

// placeReport is one place's part of the report.
func (m *Map) placeReport(place Place) PlaceReport {
	out := PlaceReport{Place: clean(nameOf(place)), Alerts: []PlaceAlert{}, Answers: []Answer{}}
	nearby := m.nearby
	if nearby == 0 {
		nearby = defaultNearby
	}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		rest := o
		rest.Features = nil
		for _, f := range o.Features {
			if overlay.SeverityOf(f) == 0 {
				rest.Features = append(rest.Features, f)
			}
		}
		for _, alert := range alertsOf(o.Features) {
			a := m.placeAlert(place, id, alert, nearby)
			if a.Valid.IsZero() {
				a.Valid = o.Valid
			}
			a.Stale = m.staleOverlay(o)
			out.Alerts = append(out.Alerts, a)
		}
		if len(o.Features) > 0 && len(rest.Features) == 0 {
			continue // every feature was an alert, answered on its own above
		}
		out.Answers = append(out.Answers, m.answerFor(place, id, rest))
	}
	return out
}

// alertsOf groups an overlay's alert features into alerts, in the order they
// first appear: the features that share an ID are one alert - one hazard,
// whose areas a service often sends as several overlapping zones - and a
// feature with no ID is an alert of its own.
func alertsOf(features []Feature) [][]Feature {
	var out [][]Feature
	at := map[string]int{}
	for _, f := range features {
		if overlay.SeverityOf(f) == 0 {
			continue
		}
		if i, ok := at[f.ID]; ok && f.ID != "" {
			out[i] = append(out[i], f)
			continue
		}
		at[f.ID] = len(out)
		out = append(out, []Feature{f})
	}
	return out
}

// anyInView reports whether any area of an alert meets the view.
func (m *Map) anyInView(alert []Feature) bool {
	for _, f := range alert {
		if m.inView(f) {
			return true
		}
	}
	return false
}

// placeAlert is one alert against one place: inside if inside any of its
// areas, one at a time so that two overlapping areas never cancel; the
// nearest edge of all of them; and nearby when outside within the distance
// set. Its label and severity are its first area's.
func (m *Map) placeAlert(place Place, id string, alert []Feature, nearbyKm float64) PlaceAlert {
	areas := make([][][]project.LonLat, 0, len(alert))
	for _, f := range alert {
		areas = append(areas, f.Rings)
	}
	f := alert[0]
	a := describe.OfArea(nameOf(place), id, place.At, areas, m.units)
	out := PlaceAlert{Overlay: clean(id), Feature: clean(f.ID), Label: clean(f.Label), Severity: overlay.SeverityOf(f),
		Where: a.Relation, Distance: a.Distance, Unit: a.Unit, Bearing: a.Bearing, Compass: a.Compass, Valid: f.Valid,
		UnderOneCell: m.underOneCell(a)}
	km := a.Distance
	if a.Unit == "miles" {
		km *= 1.609344
	}
	if out.Where == Outside && !a.NoData && km <= nearbyKm {
		out.Where = Nearby
	}
	return out
}

// inView reports whether any of a feature's outline meets the view.
func (m *Map) inView(f Feature) bool {
	w, h := float64(2*m.view.Cols), float64(4*m.view.Rows)
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, ring := range f.Rings {
		for _, at := range ring {
			x, y, err := m.view.ToDot(at)
			if err != nil {
				continue
			}
			minX, minY, maxX, maxY = math.Min(minX, x), math.Min(minY, y), math.Max(maxX, x), math.Max(maxY, y)
		}
	}
	return maxX >= 0 && minX < w && maxY >= 0 && minY < h
}

func clean(s string) string { return textsafe.Clean(s).String() }

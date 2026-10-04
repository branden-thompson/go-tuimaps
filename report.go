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

// MotionReport is observed motion (L-1.12, D-42, D-122) for one loop and
// one named place - or, with Place empty, the view's centre: the way the
// heavier rain near it moves (Heading, towards, not from) and how fast,
// measured over the pairs of frames that could be measured; where the
// nearest heavier rain is at the newest frame (To), and where it stood the
// measured time before by that motion (From, inferred, not seen); whether it
// came closer, moved away or held; and the span measured. When the rain
// barely moved (the contract states the speed), Moving is false,
// HeadingCompass empty, Heading means nothing and Trend is Held. It is data
// about what was seen: nothing in it can say what will happen.
//
// A loop with no motion to tell still has its entry (D-111): Missing says
// why, and every other field but Overlay, Place and Threshold is zero.
type MotionReport struct {
	Overlay, Place string
	Threshold      int // the class taken as heavier rain, held for the whole loop
	From, To       Sighting
	Trend          Trend
	Span           time.Duration
	Heading        float64 // the compass bearing the rain moves towards, in degrees
	HeadingCompass string
	SpeedKmh       float64
	Moving         bool
	Missing        MotionMissing
}

// MotionMissing is why a loop has no motion to tell (D-111).
type MotionMissing = describe.Missing

// The reasons: no heavier rain at the newest frame, fewer than two frames
// that could be measured, or frames still being read.
const (
	MotionNoHeavierRain = describe.NoHeavierRain
	MotionTooFewFrames  = describe.TooFewFrames
	MotionDecoding      = describe.Decoding
)

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
	tracked := make(map[string][]describe.Frame, len(m.tracked))
	defer func() { m.tracked = tracked }() // a loop no longer held is let go
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
		observed := m.store.ObservedFrames(id)
		frames := m.tracked[id]
		if !samePictures(frames, observed) {
			frames = make([]describe.Frame, 0, len(observed))
			for _, f := range observed {
				r := f.Raster
				frames = append(frames, describe.Frame{Valid: f.Valid, Image: describe.Image{West: r.West, South: r.South, East: r.East, North: r.North,
					Width: r.Width, Height: r.Height, Classes: r.Classes, Mercator: r.Projection == uint8(WebMercator)}})
			}
		}
		tracked[id] = frames
		reading := len(observed) < observable(o.Image.Frames)
		for _, ref := range refs {
			out = append(out, m.motionOf(clean(id), ref, frames, threshold, reading))
		}
	}
	return out
}

// motionOf is one loop's motion for one place: measured, or why not.
func (m *Map) motionOf(id string, ref Place, frames []describe.Frame, threshold int, reading bool) MotionReport {
	entry := MotionReport{Overlay: id, Place: clean(ref.Name), Threshold: threshold}
	mo, missing := describe.Track(frames, threshold, ref.At)
	if missing == 0 {
		fromKm, _, _, fromOK := describe.Measure(ref.At, mo.From.At)
		toKm, _, _, toOK := describe.Measure(ref.At, mo.To.At)
		if !fromOK || !toOK {
			missing = describe.TooFewFrames // a sighting that cannot be measured has no trend to tell
		} else {
			entry.From, entry.To = m.sighting(ref.At, mo.From), m.sighting(ref.At, mo.To)
			entry.Trend, entry.Span = describe.Held, mo.To.Valid.Sub(mo.From.Valid)
			entry.SpeedKmh, entry.Moving = mo.SpeedKmh, mo.Moving
			if mo.Moving { // rain that barely moved held, whatever a few kilometres over the span say
				entry.Trend = describe.TrendOf(fromKm, toKm)
				entry.Heading = mo.Heading
				if word, err := project.Compass(mo.Heading); err == nil {
					entry.HeadingCompass = word.String()
				}
			}
			return entry
		}
	}
	if reading {
		missing = describe.Decoding // what is missing may yet be read
	}
	entry.Missing = missing
	return entry
}

// samePictures reports whether frames kept from an earlier report were made
// from these observed frames: the same times and the same decoded classes,
// so what was measured over them still holds.
func samePictures(kept []describe.Frame, observed []overlay.ObservedFrame) bool {
	if len(kept) != len(observed) {
		return false
	}
	for i, f := range observed {
		k := kept[i].Image.Classes
		if !kept[i].Valid.Equal(f.Valid) || len(k) != len(f.Raster.Classes) || (len(k) > 0 && &k[0] != &f.Raster.Classes[0]) {
			return false
		}
	}
	return true
}

// observable is how many of a loop's frames can be observed: neither a gap
// nor a forecast.
func observable(frames []LoopFrame) int {
	n := 0
	for _, f := range frames {
		if !f.Gap && !f.Forecast {
			n++
		}
	}
	return n
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
		for n, alert := range alertsOf(o.Features) {
			a := m.placeAlert(place, id, n, alert, nearby)
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
func (m *Map) placeAlert(place Place, id string, n int, alert []Feature, nearbyKm float64) PlaceAlert {
	f := alert[0]
	a := m.areas.measure(m, areaKey{overlay: id, alert: n, place: nameOf(place), at: place.At}, func() describe.Answer {
		areas := make([][][]project.LonLat, 0, len(alert))
		for _, f := range alert {
			areas = append(areas, f.Rings)
		}
		return describe.OfArea(nameOf(place), id, place.At, areas, m.units)
	})
	out := PlaceAlert{Overlay: clean(id), Feature: clean(f.ID), Label: clean(f.Label), Severity: overlay.SeverityOf(f),
		Where: a.Relation, Distance: a.Distance, Unit: a.Unit, Bearing: a.Bearing, Compass: a.Compass, Valid: f.Valid,
		UnderOneCell: m.underOneCell(a)}
	if out.Where == Outside && !a.NoData && a.Km <= nearbyKm {
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

// areaKey is one measure: a place against one alert of one overlay, the
// alert by its place in the overlay's order.
type areaKey struct {
	overlay string
	alert   int
	place   string
	at      project.LonLat
}

// areaMemo keeps each place's measure against each alert - the nearest edge
// of every area, the description's costliest sum - while nothing it depends
// on changes: the overlays (any Set or Remove) and the units. The view, the
// clock and the work landing do not touch it, and a report is worked out
// again on each of those - after a feed, once for every job that lands (W14).
type areaMemo struct {
	overlays uint64
	units    describe.Units
	kept     map[areaKey]describe.Answer
	measured int // how many measures were worked out, for the library's tests
}

// maxAreaMeasures bounds the kept measures: a host asking about ever more
// places between two changes of its overlays starts afresh, never grows it.
const maxAreaMeasures = 4096

// measure is the kept measure for key, or work's, kept.
func (a *areaMemo) measure(m *Map, key areaKey, work func() describe.Answer) describe.Answer {
	if a.kept == nil || a.overlays != m.overlays || a.units != m.units || len(a.kept) >= maxAreaMeasures {
		a.kept, a.overlays, a.units = map[areaKey]describe.Answer{}, m.overlays, m.units
	}
	if out, ok := a.kept[key]; ok {
		return out
	}
	out := work()
	a.measured++
	a.kept[key] = out
	return out
}

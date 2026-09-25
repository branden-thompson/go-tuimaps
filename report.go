package tuimaps

import (
	"math"
	"time"

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
// own, and every other overlay's answer as Describe gave it - an image's
// class, a field's value, the nearest point or line.
type PlaceReport struct {
	Place   string
	Alerts  []PlaceAlert
	Answers []Answer
}

// PlaceAlert is one alert against one place (L-13.6): inside, nearby or
// outside, how far its edge is and which way, with its label and severity.
type PlaceAlert struct {
	Overlay, Feature, Label string
	Severity                Severity
	Where                   Where
	Distance                float64
	Unit                    string
	Bearing                 float64
	Compass                 string
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
// overlay's answer. Like Describe it is computed from the host's geometry,
// never from the drawn cells, and it needs no Work to have run.
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
	out = Report{Alerts: []AlertShown{}, Places: []PlaceReport{}}
	for _, id := range m.store.IDs() {
		reader, ok := m.store.Read(id)
		if !ok {
			continue
		}
		o := reader.Overlay()
		reader.Done()
		stale := m.staleOverlay(o)
		for _, f := range o.Features {
			if overlay.SeverityOf(f) == 0 || !m.inView(f) {
				continue
			}
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
	return out, nil
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
				continue
			}
			out.Alerts = append(out.Alerts, m.placeAlert(place, id, f, nearby))
		}
		if len(o.Features) > 0 && len(rest.Features) == 0 {
			continue // every feature was an alert, answered on its own above
		}
		out.Answers = append(out.Answers, m.answerFor(place, id, rest))
	}
	return out
}

// placeAlert is one alert against one place: the area test and the nearest
// edge of that alert alone, and nearby when outside within the distance set.
func (m *Map) placeAlert(place Place, id string, f Feature, nearbyKm float64) PlaceAlert {
	a := describe.OfArea(nameOf(place), id, place.At, [][][]project.LonLat{f.Rings}, m.units)
	out := PlaceAlert{Overlay: clean(id), Feature: clean(f.ID), Label: clean(f.Label), Severity: overlay.SeverityOf(f),
		Where: a.Relation, Distance: a.Distance, Unit: a.Unit, Bearing: a.Bearing, Compass: a.Compass}
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

package tuimaps_test

// report_test.go — v0.2.0 WP-L5: Report, the structured answer beside the
// picture (D-57), in v0.1.0's names and units.

import (
	"image/color"
	"math"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// boxAlert is one alert area, a box, with an id and a label.
func boxAlert(id string, role tuimaps.Token, label string, west, south, east, north float64) tuimaps.Feature {
	return tuimaps.Feature{Kind: tuimaps.Polygon, Role: role, Label: label, ID: id,
		Rings: [][]tuimaps.LonLat{{{Lon: west, Lat: south}, {Lon: east, Lat: south}, {Lon: east, Lat: north}, {Lon: west, Lat: north}, {Lon: west, Lat: south}}}}
}

func reportMap(t *testing.T) *tuimaps.Map {
	t.Helper()
	m := world(t, 80, 24)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -90.5, Lat: 35.5}))
	must(t, m.Zoom(5))
	return m
}

// TestAnEmptyReportHasEmptySections is L5.2: an empty map answers with
// every section present and empty, never nil.
func TestAnEmptyReportHasEmptySections(t *testing.T) {
	r, err := reportMap(t).Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Alerts == nil || r.Places == nil || len(r.Alerts) != 0 || len(r.Places) != 0 {
		t.Errorf("an empty map's report: %+v; want empty sections, not nil", r)
	}
}

// TestTheAlertsShown is L5.3 (L-13.5): the alerts in view, needing no place,
// in draw order, each with its label, severity, times and stale mark; an
// alert out of view is not listed; every string comes back cleaned.
func TestTheAlertsShown(t *testing.T) {
	m := reportMap(t)
	o := tuimaps.Overlay{ID: "nws", Valid: noon, Keeps: time.Hour, Features: []tuimaps.Feature{
		boxAlert("a", tuimaps.AlertExtreme, "Tornado Warning", -91, 35, -90, 36),
		boxAlert("b", tuimaps.AlertModerate, "Flood Watch", -92, 34, -91.5, 34.5),
		boxAlert("far", tuimaps.AlertMinor, "Far Away", 10, 10, 11, 11),
		boxAlert("c", tuimaps.AlertUnknown, "Special \x1b[31mStatement\u202e", -90, 36, -89.5, 36.3),
	}}
	o.Features[0].Valid, o.Features[0].Expires = noon.Add(-10*time.Minute), noon.Add(time.Hour)
	mustSet(t, m, o)
	r, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, a := range r.Alerts {
		ids = append(ids, a.Feature)
	}
	if strings.Join(ids, ",") != "a,b,c" {
		t.Fatalf("the alerts shown: %v; want a, b, c in draw order and not far", ids)
	}
	first := r.Alerts[0]
	if first.Overlay != "nws" || first.Label != "Tornado Warning" || first.Severity != tuimaps.SeverityExtreme ||
		!first.Valid.Equal(noon.Add(-10*time.Minute)) || !first.Expires.Equal(noon.Add(time.Hour)) {
		t.Errorf("the first alert: %+v", first)
	}
	if !r.Alerts[1].Valid.Equal(noon) {
		t.Errorf("an alert with no valid time of its own: %v; want the overlay's", r.Alerts[1].Valid)
	}
	if strings.ContainsAny(r.Alerts[2].Label, "\x1b\u202e") {
		t.Errorf("a hostile label came back as it went in: %q", r.Alerts[2].Label)
	}
	// Moving the view changes what is shown, though nothing else changed.
	must(t, m.Recentre(tuimaps.LonLat{Lon: 10.5, Lat: 10.5}))
	if r, err := m.Report(nil); err != nil || len(r.Alerts) != 1 || r.Alerts[0].Feature != "far" {
		t.Errorf("after moving the view to the far alert: %+v, %v", r.Alerts, err)
	}
}

// TestEachAlertIsAnsweredOnItsOwn is L5.4 (L-13.6, D-43): for a named place,
// each alert on its own - inside, nearby within the distance the host sets
// (10 km by default), or outside - with its label and severity, in the host's
// units; and the other answers, an image's among them, beside them.
func TestEachAlertIsAnsweredOnItsOwn(t *testing.T) {
	m := reportMap(t)
	mustSet(t, m, tuimaps.Overlay{ID: "nws", Valid: noon, Keeps: time.Hour, Features: []tuimaps.Feature{
		boxAlert("warn", tuimaps.AlertSevere, "Severe Thunderstorm Warning", -91, 35, -90, 36),
		boxAlert("watch", tuimaps.AlertModerate, "Flood Watch", -92.5, 35, -91.5, 36),
	}})
	mustSet(t, m, rainOver(t, -95, 30, -85, 40, color.NRGBA{R: 200, A: 255}))
	settle(t, m)
	// One degree of longitude at 35.5 north is about 90.6 km.
	six := tuimaps.Place{ID: "six", Name: "Six", At: tuimaps.LonLat{Lon: -90 + 6/90.6, Lat: 35.5}}
	fourteen := tuimaps.Place{ID: "fourteen", Name: "Fourteen", At: tuimaps.LonLat{Lon: -90 + 14/90.6, Lat: 35.5}}
	in := tuimaps.Place{ID: "in", Name: "In", At: tuimaps.LonLat{Lon: -90.5, Lat: 35.5}}
	where := func(r tuimaps.Report, place, feature string) (tuimaps.PlaceAlert, bool) {
		for _, p := range r.Places {
			if p.Place != place {
				continue
			}
			for _, a := range p.Alerts {
				if a.Feature == feature {
					return a, true
				}
			}
		}
		return tuimaps.PlaceAlert{}, false
	}
	r, err := m.Report([]tuimaps.Place{six, fourteen, in})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		place, feature string
		want           tuimaps.Where
	}{
		{"Six", "warn", tuimaps.Nearby}, {"Fourteen", "warn", tuimaps.Outside}, {"In", "warn", tuimaps.Inside},
		{"In", "watch", tuimaps.Outside},
	} {
		a, ok := where(r, c.place, c.feature)
		if !ok || a.Where != c.want {
			t.Errorf("%s against %s: %+v; want %v", c.place, c.feature, a, c.want)
		}
	}
	six6, _ := where(r, "Six", "warn")
	if six6.Label != "Severe Thunderstorm Warning" || six6.Severity != tuimaps.SeveritySevere || six6.Unit != "kilometres" || math.Abs(six6.Distance-6) > 0.3 || six6.Compass != "west" {
		t.Errorf("Six against the warning: %+v; want its label, severe, about 6 kilometres to the west", six6)
	}
	images := 0
	for _, p := range r.Places {
		for _, a := range p.Answers {
			if a.Overlay == "radar" {
				images++
			}
		}
	}
	if images != 3 {
		t.Errorf("the image's answers: %d; want one for each of the three places", images)
	}
	for _, p := range r.Places {
		for _, a := range p.Answers {
			if a.Overlay == "nws" {
				t.Errorf("%s: an overlay of alerts alone also gave a merged answer: %+v", p.Place, a)
			}
		}
	}
	// In miles, "nearby" is still judged in kilometres: 14 km is outside 10.
	m.Units(true, false)
	r, err = m.Report([]tuimaps.Place{fourteen})
	if err != nil {
		t.Fatal(err)
	}
	if a, _ := where(r, "Fourteen", "warn"); a.Where != tuimaps.Outside {
		t.Errorf("14 km away, in miles, with nearby at 10 km: %v; want outside", a.Where)
	}
	must(t, m.SetNearby(20))
	m.Units(true, false)
	r, err = m.Report([]tuimaps.Place{fourteen})
	if err != nil {
		t.Fatal(err)
	}
	if a, _ := where(r, "Fourteen", "warn"); a.Where != tuimaps.Nearby || a.Unit != "miles" || math.Abs(a.Distance-14/1.609344) > 0.3 {
		t.Errorf("with nearby at 20 km and miles: %+v; want nearby, about 8.7 miles", a)
	}
	if err := m.SetNearby(-1); err == nil {
		t.Error("a negative nearby distance was accepted")
	}
}

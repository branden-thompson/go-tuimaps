package tuimaps_test

import (
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// rect is one rectangular ring, closed the way a service sends one.
func rect(w, s, e, n float64) []tuimaps.LonLat {
	return []tuimaps.LonLat{
		{Lon: w, Lat: s}, {Lon: e, Lat: s}, {Lon: e, Lat: n}, {Lon: w, Lat: n}, {Lon: w, Lat: s},
	}
}

// TestAPlaceInsideTwoOverlappingAreasIsInside.
//
// **One hazard commonly covers several areas that overlap.** The National
// Weather Service names a marine zone and the coastal land zone beside it in
// one alert, and they share water; two adjacent counties under one warning
// share their border. Each area is its own feature, because within a feature
// the first ring is an outline and the rest are its holes.
//
// The answer for a place must be about the hazard, not about how many of its
// areas happen to contain that place. **Counting crossings across every
// feature at once makes two overlapping areas cancel**, so a house inside both
// is reported outside - while the fill, which draws each feature separately
// (FR-11), shades it. A map that shows a place inside a warning while the
// words say it is outside is worse than either answer alone, and the words are
// what a screen reader gets (D-52).
func TestAPlaceInsideTwoOverlappingAreasIsInside(t *testing.T) {
	// A covers -85..-83 / 41..43, B covers -84..-82 / 42..44.
	// They overlap on -84..-83 / 42..43, and the house is in the middle of it.
	house := tuimaps.LonLat{Lon: -83.5, Lat: 42.5}
	// One hazard's areas share its ID (v0.2.0 L5.4): Report answers the
	// hazard once, inside if inside any of them.
	areaA := tuimaps.Feature{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, ID: "hazard", Rings: [][]tuimaps.LonLat{rect(-85, 41, -83, 43)}}
	areaB := tuimaps.Feature{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, ID: "hazard", Rings: [][]tuimaps.LonLat{rect(-84, 42, -82, 44)}}

	for _, c := range []struct {
		what  string
		feats []tuimaps.Feature
	}{
		{"one area", []tuimaps.Feature{areaA}},
		{"the two overlapping areas of one hazard", []tuimaps.Feature{areaA, areaB}},
		{"the same two, named the other way round", []tuimaps.Feature{areaB, areaA}},
		{"three areas overlapping at the house", []tuimaps.Feature{areaA, areaB, areaA}},
	} {
		m, err := tuimaps.New()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := m.Set(tuimaps.Overlay{
			ID: "alerts", Valid: time.Now(), Keeps: time.Hour, Features: c.feats,
		}); err != nil {
			t.Fatal(err)
		}
		got, err := m.Report([]tuimaps.Place{{ID: "home", Name: "Home", At: house}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Places) != 1 || len(got.Places[0].Alerts) != 1 {
			t.Fatalf("%s: %+v; want one answer, for the one hazard", c.what, got.Places)
		}
		if a := got.Places[0].Alerts[0]; a.Where != tuimaps.Inside {
			t.Errorf("%s: the house is %v; it is within every area named", c.what, a.Where)
		}
	}
	// Areas with no ID are alerts of their own, each answered.
	a, b := areaA, areaB
	a.ID, b.ID = "", ""
	m, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(tuimaps.Overlay{ID: "alerts", Valid: time.Now(), Keeps: time.Hour, Features: []tuimaps.Feature{a, b}}); err != nil {
		t.Fatal(err)
	}
	got, err := m.Report([]tuimaps.Place{{ID: "home", Name: "Home", At: house}})
	if err != nil {
		t.Fatal(err)
	}
	if n := len(got.Places[0].Alerts); n != 2 || got.Places[0].Alerts[0].Where != tuimaps.Inside || got.Places[0].Alerts[1].Where != tuimaps.Inside {
		t.Errorf("two areas with no ID: %+v; want two answers, both inside", got.Places[0].Alerts)
	}
}

// TestAPlaceInAHoleIsStillOutside is the rule the one above must not break.
// **Within one area, even-odd is exactly right**: a ring inside the outline is
// a hole, and a place in the hole is outside the area. Only the flattening
// ACROSS areas was wrong, and a fix that made everything "inside if inside any
// ring" would lose holes altogether.
func TestAPlaceInAHoleIsStillOutside(t *testing.T) {
	m, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	holed := tuimaps.Feature{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, Rings: [][]tuimaps.LonLat{
		rect(-85, 41, -83, 43),         // the outline
		rect(-84.5, 41.5, -83.5, 42.5), // a lake in the middle of it
	}}
	if _, err := m.Set(tuimaps.Overlay{
		ID: "alerts", Valid: time.Now(), Keeps: time.Hour, Features: []tuimaps.Feature{holed},
	}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		what string
		at   tuimaps.LonLat
		want tuimaps.Where
	}{
		{"in the lake", tuimaps.LonLat{Lon: -84, Lat: 42}, tuimaps.Outside}, // about 41 km from the shore: beyond nearby
		{"on the land around it", tuimaps.LonLat{Lon: -83.2, Lat: 42.8}, tuimaps.Inside},
	} {
		got, err := m.Report([]tuimaps.Place{{ID: "p", Name: "P", At: c.at}})
		if err != nil {
			t.Fatal(err)
		}
		if a := got.Places[0].Alerts[0]; a.Where != c.want {
			t.Errorf("a place %s is %v; it is %v", c.what, a.Where, c.want)
		}
	}
}

// TestDescribingADifferentPlaceDoesNotReuseTheLastAnswer.
//
// `Report` takes the places to ask about, so a host may ask about places the
// map does not store - **which is exactly what a weather station watching
// several locations does**: one alert, and "where is each of my places
// relative to it".
//
// The report is memoised, and the memo must change whenever the answer
// would (FR-29). Counting the places asked about is not enough: two calls
// about *different* places are both a call about one place.
func TestAskingAboutADifferentPlaceDoesNotReuseTheLastAnswer(t *testing.T) {
	m, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(tuimaps.Overlay{
		ID: "alerts", Valid: time.Now(), Keeps: time.Hour,
		Features: []tuimaps.Feature{{
			Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere,
			Rings: [][]tuimaps.LonLat{rect(-85, 41, -83, 43)},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	ask := func(name string, at tuimaps.LonLat) tuimaps.PlaceAlert {
		t.Helper()
		got, err := m.Report([]tuimaps.Place{{ID: name, Name: name, At: at}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Places) != 1 || len(got.Places[0].Alerts) != 1 {
			t.Fatalf("%s: %+v", name, got.Places)
		}
		if got.Places[0].Place != name {
			t.Errorf("asked about %s and was answered about %s", name, got.Places[0].Place)
		}
		return got.Places[0].Alerts[0]
	}
	if in := ask("Fort Wayne", tuimaps.LonLat{Lon: -84, Lat: 42}); in.Where != tuimaps.Inside {
		t.Fatalf("a place within the area is %v", in.Where)
	}
	// A different place, a thousand kilometres away, asked about second.
	if out := ask("Denver", tuimaps.LonLat{Lon: -105, Lat: 39.7}); out.Where != tuimaps.Outside {
		t.Errorf("a place a thousand kilometres from the area is %v - the last answer was reused", out.Where)
	}
}

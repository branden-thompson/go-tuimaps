package tuimaps_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// TestReportIsReadyAtOnce is plan task 11.16 (FR-29, contract section 1),
// ported to Report (v0.2.0 L5.8): the call answers at once, from the host's
// own data, with nothing left to work out later.
func TestReportIsReadyAtOnce(t *testing.T) {
	m := gulfMap(t, 149, 38)
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Home", At: tuimaps.LonLat{Lon: -84.39, Lat: 33.75}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	// No Settle, no Work: the answer is there.
	got, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Places) != 1 || got.Places[0].Place != "Home" {
		t.Fatalf("%+v", got.Places)
	}
	if len(got.Places[0].Alerts) != 1 {
		t.Fatalf("%d alert answers for one alert", len(got.Places[0].Alerts))
	}
	one := got.Places[0].Alerts[0]
	if one.Overlay != "alerts" || one.Where != tuimaps.Outside {
		t.Errorf("Atlanta against the Gulf alert area: %+v; want outside", one)
	}
	if one.Distance <= 0 || one.Unit != "kilometres" || one.Compass == "" {
		t.Errorf("%+v", one)
	}
	if one.Valid.IsZero() {
		t.Error("the answer does not say when the data was valid")
	}
}

// TestReportEveryForm: each shape a host can hand in is described in its
// own terms, and the words are the host's units.
func TestReportEveryForm(t *testing.T) {
	m := gulfMap(t, 149, 38)
	home := tuimaps.Place{Name: "Home", At: tuimaps.LonLat{Lon: -84.39, Lat: 33.75}}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	track := tuimaps.Overlay{ID: "track", Valid: noon, Keeps: time.Hour, Credit: "A survey",
		Features: []tuimaps.Feature{{Kind: tuimaps.Line, Label: "storm track", Role: tuimaps.Track,
			Rings: [][]tuimaps.LonLat{{{Lon: -90, Lat: 33}, {Lon: -80, Lat: 34}}}}}}
	if _, err := m.Set(track); err != nil {
		t.Fatal(err)
	}
	grid := tuimaps.Grid{West: -100, South: 20, East: -70, North: 40, Cols: 2, Rows: 2, Values: []float64{10, 20, 12, 22}}
	if _, err := m.Set(tuimaps.TemperatureGrid("temperature", grid, tuimaps.Celsius, noon)); err != nil {
		t.Fatal(err)
	}
	m.Units(true, true) // miles and Fahrenheit
	got, err := m.Report([]tuimaps.Place{home})
	if err != nil {
		t.Fatal(err)
	}
	forms := map[string]tuimaps.Answer{}
	for _, a := range got.Places[0].Answers {
		forms[a.Overlay] = a
	}
	if alerts := got.Places[0].Alerts; len(alerts) != 1 || alerts[0].Overlay != "alerts" || alerts[0].Unit != "miles" {
		t.Errorf("the alert answer is %+v", alerts)
	}
	if forms["track"].Form != "line" || forms["track"].Label != "storm track" {
		t.Errorf("the track answer is %+v", forms["track"])
	}
	field := forms["temperature"]
	if field.Form != "field" || field.ValueUnit != "degrees Fahrenheit" {
		t.Errorf("the field answer is %+v", field)
	}
	if field.NoData {
		t.Error("a field that covers the place says it has no data")
	}
	// Everything said is speakable: no braille, no glyphs, no escapes.
	for _, a := range got.Places[0].Answers {
		for _, said := range a.Said() {
			if strings.ContainsAny(said, "\x1b⠀░") {
				t.Errorf("an answer carries %q", said)
			}
		}
	}
}

// TestReportMemoised is plan task 11.11 (FR-29), ported to Report: asking
// twice with nothing changed costs nothing, and anything that would change
// the answer makes it be worked out again.
func TestReportMemoised(t *testing.T) {
	m := gulfMap(t, 149, 38)
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Home", At: tuimaps.LonLat{Lon: -84.39, Lat: 33.75}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	first, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	again, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if &first.Places[0].Alerts[0] != &again.Places[0].Alerts[0] {
		t.Error("asking twice with nothing changed worked the answer out again")
	}
	if allocs := testing.AllocsPerRun(20, func() { m.Report(nil) }); allocs > 0 {
		t.Errorf("repeating an unchanged report allocates %v times", allocs)
	}
	// A change to the overlays, the places, the units or nearby is worked out again.
	if _, err := m.Set(warning("more")); err != nil {
		t.Fatal(err)
	}
	third, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(third.Places[0].Alerts) != 2 {
		t.Errorf("after a second overlay there are %d alert answers", len(third.Places[0].Alerts))
	}
	m.Units(true, false)
	fourth, _ := m.Report(nil)
	if fourth.Places[0].Alerts[0].Unit != "miles" {
		t.Error("changing the units did not work the report out again")
	}
	if _, err := m.AddPlace(tuimaps.Place{Name: "Away", At: tuimaps.LonLat{Lon: 0, Lat: 0}}); err != nil {
		t.Fatal(err)
	}
	fifth, _ := m.Report(nil)
	if len(fifth.Places) != 2 {
		t.Errorf("after a second place there are %d place reports", len(fifth.Places))
	}
	near := fifth.Places[0].Alerts[0].Where
	must(t, m.SetNearby(1000))
	sixth, _ := m.Report(nil)
	if sixth.Places[0].Alerts[0].Where == near {
		t.Error("changing what nearby means did not work the report out again")
	}
}

// TestReportNeverInsideRender is plan task 11.12: drawing does not
// report, and reporting does not draw. A host pays for what it asks for.
func TestReportNeverInsideRender(t *testing.T) {
	m := gulfMap(t, 80, 24)
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Home", At: tuimaps.LonLat{Lon: -84.39, Lat: 33.75}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frameAtTime(t, m, 80, 24, noon)
	// Drawing again allocates what drawing allocates; if it described as
	// well, the count would carry the description's own work.
	drawing := testing.AllocsPerRun(10, func() { m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon) })
	m.Report(nil)
	describing := testing.AllocsPerRun(10, func() { m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon) })
	if describing > drawing {
		t.Errorf("drawing allocates %v before a description and %v after it", drawing, describing)
	}
}

// TestReportOnAClosedMap: a closed map reports nothing and says why.
func TestReportOnAClosedMap(t *testing.T) {
	m := world(t, 40, 12)
	m.Close()
	if _, err := m.Report(nil); err == nil {
		t.Error("a closed map reported something")
	}
}

// BenchmarkReportSixtyPlaces is plan task 11.15 (FR-29's cost): sixty
// places against a handful of overlays, which is the fixture the target of
// fifty milliseconds was set against. The figure is recorded, not gated.
func BenchmarkReportSixtyPlaces(b *testing.B) {
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		b.Fatal(err)
	}
	defer m.Close()
	places := make([]tuimaps.Place, 0, 60)
	for i := range 60 {
		places = append(places, tuimaps.Place{
			Name: "place " + string(rune('a'+i%26)),
			At:   tuimaps.LonLat{Lon: -90 + float64(i)/4, Lat: 25 + float64(i)/8},
		})
	}
	if _, err := m.SetPlaces(places); err != nil {
		b.Fatal(err)
	}
	if _, err := m.Set(warningAt("alerts", -86, 24, -82, 28)); err != nil {
		b.Fatal(err)
	}
	if _, err := m.Set(warningAt("more-alerts", -95, 30, -88, 36)); err != nil {
		b.Fatal(err)
	}
	grid := tuimaps.Grid{West: -100, South: 20, East: -70, North: 40, Cols: 8, Rows: 8, Values: make([]float64, 64)}
	for i := range grid.Values {
		grid.Values[i] = float64(i % 30)
	}
	if _, err := m.Set(tuimaps.TemperatureGrid("temperature", grid, tuimaps.Celsius, noon)); err != nil {
		b.Fatal(err)
	}
	version := uint64(0)
	b.ResetTimer()
	for b.Loop() {
		// A place added and removed each round, so that nothing is answered
		// from the memo: this measures the work, not the remembering.
		version++
		if _, err := m.AddPlace(tuimaps.Place{ID: "moving", Name: "moving", At: tuimaps.LonLat{Lon: -84, Lat: 33 + float64(version%7)/100}}); err != nil {
			b.Fatal(err)
		}
		if _, err := m.Report(nil); err != nil {
			b.Fatal(err)
		}
	}
}

// warningAt is an alert area over a box, for the benchmark.
func warningAt(id string, west, south, east, north float64) tuimaps.Overlay {
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "National Weather Service",
		Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, Label: "Warning",
			Rings: [][]tuimaps.LonLat{{{Lon: west, Lat: south}, {Lon: east, Lat: south}, {Lon: east, Lat: north}, {Lon: west, Lat: north}, {Lon: west, Lat: south}}}}}}
}

// TestStalenessIsJudgedByTheFramesOwnRule is a defect the app found: the
// frame's stale mark says plainly that a host which has given no time is
// told nothing about time (D-114), and every answer of a description said
// the opposite - with no frame yet drawn, the wall clock is the zero
// instant, every valid time is far ahead of it, and every overlay read as
// stale. One fact cannot have two answers in one call.
func TestStalenessIsJudgedByTheFramesOwnRule(t *testing.T) {
	const cols, rows = 69, 12
	m := gulfMap(t, cols, rows)
	if _, err := m.SetPlaces([]tuimaps.Place{miami()}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	// Reported before anything is drawn: the host has said nothing about
	// the time, so nothing is said back about it.
	said, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range said.Places {
		for _, a := range one.Alerts {
			if a.Stale {
				t.Errorf("%s in %s is called stale, and no clock has been given", one.Place, a.Overlay)
			}
			if a.Valid.IsZero() {
				t.Errorf("%s in %s carries no valid time, which is what a host judges it by", one.Place, a.Overlay)
			}
		}
	}
	// Once a frame is drawn, the answer follows that frame's clock.
	if _, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon.Add(9*time.Hour)); err != nil {
		t.Fatal(err)
	}
	said, err = m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range said.Places {
		for _, a := range one.Alerts {
			if !a.Stale {
				t.Errorf("%s in %s is not stale nine hours after its data was valid", one.Place, a.Overlay)
			}
		}
	}
	// And a frame drawn while the data is current takes the mark away.
	if _, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	said, err = m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range said.Places {
		for _, a := range one.Alerts {
			if a.Stale {
				t.Errorf("%s in %s is stale a minute after its data was valid", one.Place, a.Overlay)
			}
		}
	}
}

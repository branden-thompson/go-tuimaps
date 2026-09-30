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

// A FRAME DRAWN AT A LATER CLOCK IS NOT A CHANGE TO THE ANSWER (watchpost
// W14, P-11). A host passes its clock to every Render, so keying the kept
// report on the clock worked it out again on every frame - 9.8 ms and 11 MB
// a time with a twelve-frame loop, all of it the loop's motion, which the
// clock does not touch. What the clock does change is which overlays are
// stale, and that is what the report is kept by.
func TestAReportIsKeptWhileOnlyTheClockMoves(t *testing.T) {
	const cols, rows = 69, 12
	m := gulfMap(t, cols, rows)
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	size := tuimaps.Size{Cols: cols, Rows: rows}
	if _, err := m.Render(size, noon.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	first, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(size, noon.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	again, err := m.Report(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Alerts) == 0 || &first.Alerts[0] != &again.Alerts[0] {
		t.Error("a frame a minute later, with nothing gone stale, worked the report out again")
	}
}

// AND EACH OVERLAY'S AGEING IS A CHANGE: one overlay going stale while
// another already is changes the answer, though "is anything stale" does not.
func TestAReportFollowsEachOverlayGoingStale(t *testing.T) {
	const cols, rows = 69, 12
	m := gulfMap(t, cols, rows)
	short, long := warning("short"), warning("long")
	long.Keeps = 6 * time.Hour
	for _, o := range []tuimaps.Overlay{short, long} {
		if _, err := m.Set(o); err != nil {
			t.Fatal(err)
		}
	}
	staleAt := func(at time.Time) map[string]bool {
		t.Helper()
		if _, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, at); err != nil {
			t.Fatal(err)
		}
		said, err := m.Report(nil)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		for _, a := range said.Alerts {
			out[a.Overlay] = a.Stale
		}
		return out
	}
	if got := staleAt(noon.Add(2 * time.Hour)); !got["short"] || got["long"] {
		t.Fatalf("two hours on, the short overlay alone is stale; got %v", got)
	}
	if got := staleAt(noon.Add(7 * time.Hour)); !got["short"] || !got["long"] {
		t.Errorf("seven hours on, both are stale; got %v", got)
	}
}

// A PLACE'S DISTANCE TO AN ALERT IS MEASURED ONCE (watchpost W14): the sum
// is the nearest edge of every area, great circle by great circle - 28 ms a
// report over a day's alerts - and it depends on the place, the alert's
// areas and the units alone. A report is worked out again whenever work
// lands or the view moves, and a host asks one every frame: after a feed,
// twenty-four jobs landed one by one and each paid it again. Now a pan and a
// landing measure nothing; a new overlay, new units or another place do.
func TestAnAlertsDistanceIsMeasuredOnce(t *testing.T) {
	const cols, rows = 69, 12
	m := gulfMap(t, cols, rows)
	if _, err := m.SetPlaces([]tuimaps.Place{miami()}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	report := func() tuimaps.Report {
		t.Helper()
		r, err := m.Report(nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	first := report()
	measured := tuimaps.AreasMeasured(m)
	if measured == 0 {
		t.Fatal("the first report measured nothing")
	}
	if err := m.Recentre(tuimaps.LonLat{Lon: -83, Lat: 27}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	moved := report()
	if got := tuimaps.AreasMeasured(m); got != measured {
		t.Errorf("a pan and the work it landed measured %d more times", got-measured)
	}
	if moved.Places[0].Alerts[0].Distance != first.Places[0].Alerts[0].Distance || moved.Places[0].Alerts[0].Where != first.Places[0].Alerts[0].Where {
		t.Errorf("the kept measure differs: %+v, then %+v", first.Places[0].Alerts[0], moved.Places[0].Alerts[0])
	}
	m.Units(true, false)
	if report().Places[0].Alerts[0].Unit != "miles" || tuimaps.AreasMeasured(m) == measured {
		t.Error("new units were not measured again")
	}
	measured = tuimaps.AreasMeasured(m)
	if _, err := m.Set(warning("more")); err != nil {
		t.Fatal(err)
	}
	if len(report().Places[0].Alerts) != 2 || tuimaps.AreasMeasured(m) == measured {
		t.Error("a new overlay was not measured")
	}
	// The same overlay handed in again with other areas is measured again.
	before := report().Places[0].Alerts[0].Distance
	if _, err := m.Set(warningAt("alerts", -81, 26, -80.5, 26.5)); err != nil {
		t.Fatal(err)
	}
	if got := report().Places[0].Alerts[0].Distance; got == before {
		t.Errorf("an overlay replaced with other areas kept its old measure, %v", got)
	}
	measured = tuimaps.AreasMeasured(m)
	away, err := m.Report([]tuimaps.Place{{Name: "Away", At: tuimaps.LonLat{Lon: -90, Lat: 30}}})
	if err != nil {
		t.Fatal(err)
	}
	if tuimaps.AreasMeasured(m) == measured || away.Places[0].Alerts[0].Distance == first.Places[0].Alerts[0].Distance {
		t.Error("another place was answered with the first place's measure")
	}
	// Two places of one name are two places (a Springfield in each state).
	here, err := m.Report([]tuimaps.Place{{Name: "Away", At: tuimaps.LonLat{Lon: -80, Lat: 25}}})
	if err != nil {
		t.Fatal(err)
	}
	if here.Places[0].Alerts[0].Distance == away.Places[0].Alerts[0].Distance {
		t.Error("a place of the same name elsewhere was answered with the other's measure")
	}
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

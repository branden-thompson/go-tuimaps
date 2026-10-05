package tuimaps_test

// wind_test.go — v0.2.0 L-16 (FR-8; watchpost D-108 to D-110): a wind field
// in one call, drawn as arrows; its answer in words; its legend; what is
// refused.

import (
	"math"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// westerly is a wind of 15 mph from the south-west over the lower 48.
func westerly() (tuimaps.Grid, []float64) {
	const cols, rows = 12, 8
	g := tuimaps.Grid{West: -125, South: 24, East: -66, North: 50, Cols: cols, Rows: rows}
	var from []float64
	for range cols * rows {
		g.Values = append(g.Values, 15)
		from = append(from, 225)
	}
	return g, from
}

// TestAWindGridIsItsArrowsAndItsWords is L-16.1 and L-16.2: WindGrid takes
// speeds and where they blow from; the frame draws arrows and speeds and no
// band; a place's answer is its speed and the compass word it blows from.
func TestAWindGridIsItsArrowsAndItsWords(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	bare := momentFrame(t, m, noon)
	g, from := westerly()
	mustSet(t, m, tuimaps.WindGrid("wind", g, from, tuimaps.MilesPerHour, noon))
	settle(t, m)
	frame := momentFrame(t, m, noon)
	if frame == bare {
		t.Fatal("the wind drew nothing")
	}
	if !strings.Contains(colours.ReplaceAllString(frame, ""), "15") {
		t.Error("no arrow carries its speed")
	}
	r, err := m.Report([]tuimaps.Place{{ID: "p", Name: "Memphis", At: tuimaps.LonLat{Lon: -90, Lat: 35.1}}})
	if err != nil {
		t.Fatal(err)
	}
	var got *tuimaps.Answer
	for i := range r.Places[0].Answers {
		if a := &r.Places[0].Answers[i]; a.Overlay == "wind" {
			got = a
		}
	}
	if got == nil || got.Value != 15 || got.ValueUnit != "mph" || got.From == "" {
		t.Fatalf("the place's wind answer is %+v; want 15 mph and the compass word it blows from", got)
	}
}

// TestTheWindLegendIsItsSixClasses is L-16.3: the legend keys the wind preset
// in its unit, calmest first.
func TestTheWindLegendIsItsSixClasses(t *testing.T) {
	m := reportMap(t)
	g, from := westerly()
	mustSet(t, m, tuimaps.WindGrid("wind", g, from, tuimaps.KilometresPerHour, noon))
	for _, e := range m.Legend() {
		if e.Preset == "wind" {
			if len(e.Classes) != 6 || e.Unit != "km/h" || !strings.Contains(e.Classes[0].Label, "10") {
				t.Errorf("the wind legend is %+v; want six classes in km/h, the calmest under 10", e)
			}
			return
		}
	}
	t.Error("the legend has no wind")
}

// TestAWindGridsDirectionsAreChecked is L-16.4: a direction a speed, each
// from 0 to 360 or NaN.
func TestAWindGridsDirectionsAreChecked(t *testing.T) {
	m := reportMap(t)
	g, from := westerly()
	if _, err := m.Set(tuimaps.WindGrid("short", g, from[:3], tuimaps.MilesPerHour, noon)); !isKind(err, fault.SizeMismatch) {
		t.Errorf("too few directions: %v; want refused", err)
	}
	bad := append([]float64(nil), from...)
	bad[0] = 400
	if _, err := m.Set(tuimaps.WindGrid("wild", g, bad, tuimaps.MilesPerHour, noon)); !isKind(err, fault.InvalidCoordinates) {
		t.Errorf("a direction of 400: %v; want refused", err)
	}
	bad[0] = math.NaN()
	if _, err := m.Set(tuimaps.WindGrid("gap", g, bad, tuimaps.MilesPerHour, noon)); err != nil {
		t.Errorf("a missing direction was refused: %v", err)
	}
}

// TestAGustIsSaidBesideItsSpeed is L-24.1 (watchpost D-136): a wind grid's
// gusts, one a value, are said on an arrow's label as aviation says them -
// "15G30" - where one is given, and the speed alone where it is NaN.
func TestAGustIsSaidBesideItsSpeed(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	g, from := westerly()
	for range g.Values {
		g.Gusts = append(g.Gusts, 30)
	}
	mustSet(t, m, tuimaps.WindGrid("wind", g, from, tuimaps.MilesPerHour, noon))
	settle(t, m)
	if frame := colours.ReplaceAllString(momentFrame(t, m, noon), ""); !strings.Contains(frame, "15G30") {
		t.Errorf("no arrow says its gust:\n%s", frame)
	}
	for i := range g.Gusts {
		g.Gusts[i] = math.NaN()
	}
	mustSet(t, m, tuimaps.WindGrid("wind", g, from, tuimaps.MilesPerHour, noon))
	settle(t, m)
	if frame := colours.ReplaceAllString(momentFrame(t, m, noon), ""); strings.Contains(frame, "15G") || !strings.Contains(frame, "15") {
		t.Errorf("with no gust given, an arrow says one, or no speed:\n%s", frame)
	}
}

// TestAWindGridsGustsAreChecked is L-24.2: gusts belong to a wind grid, one a
// value, each a speed or NaN.
func TestAWindGridsGustsAreChecked(t *testing.T) {
	m := reportMap(t)
	g, from := westerly()
	g.Gusts = []float64{30, 30}
	if _, err := m.Set(tuimaps.WindGrid("short", g, from, tuimaps.MilesPerHour, noon)); !isKind(err, fault.SizeMismatch) {
		t.Errorf("too few gusts: %v; want refused", err)
	}
	g.Gusts = make([]float64, len(g.Values))
	g.Gusts[0] = -5
	if _, err := m.Set(tuimaps.WindGrid("below", g, from, tuimaps.MilesPerHour, noon)); !isKind(err, fault.SizeMismatch) {
		t.Errorf("a gust of -5: %v; want refused", err)
	}
	g.Gusts[0] = 30
	if _, err := m.Set(tuimaps.TemperatureGrid("temp", g, tuimaps.Fahrenheit, noon)); !isKind(err, fault.SizeMismatch) {
		t.Errorf("gusts on a temperature grid: %v; want refused", err)
	}
}

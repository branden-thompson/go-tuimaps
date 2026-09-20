package tuimaps_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestLegendShowsDrawnColours is plan task 12.7: the legend says what is on
// the map, class by class, in the colours it is actually drawn in - at every
// depth, including none at all.
func TestLegendShowsDrawnColours(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	if len(m.Legend()) != 0 {
		t.Errorf("a map with no overlay has a legend: %+v", m.Legend())
	}
	grid := tuimaps.Grid{West: -100, South: 20, East: -70, North: 35, Cols: 2, Rows: 2,
		Values: []float64{-10, 0, 15, 30}}
	if _, err := m.Set(tuimaps.TemperatureGrid("temperature", grid, tuimaps.Celsius, noon)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frameAtTime(t, m, cols, rows, noon)
	for _, depth := range []tuimaps.Depth{tuimaps.Truecolor, tuimaps.Colours256, tuimaps.Colours16, tuimaps.NoColour} {
		m.ColourDepth(depth)
		frameAtTime(t, m, cols, rows, noon)
		legend := m.Legend()
		if len(legend) != 1 {
			t.Fatalf("at %v the legend has %d entries, want the one overlay's", depth, len(legend))
		}
		one := legend[0]
		if one.ID != "temperature" || one.Unit != "C" || len(one.Classes) == 0 {
			t.Errorf("at %v: %+v", depth, one)
		}
		// A ramp has colours at truecolor and at 256. At sixteen it has none,
		// and neither has the frame: what needs a ramp is drawn in its
		// no-colour form at that depth (D-59), and the legend says the same.
		ramped := depth == tuimaps.Truecolor || depth == tuimaps.Colours256
		for _, c := range one.Classes {
			if c.Label == "" {
				t.Errorf("at %v a class has no label: %+v", depth, c)
			}
			if c.Drawn != ramped {
				t.Errorf("at %v a class is drawn %v: %+v", depth, c.Drawn, c)
			}
			if !c.Drawn && c.Colour != (tuimaps.RGB{}) {
				t.Errorf("at %v a class that is not drawn carries a colour: %+v", depth, c)
			}
		}
	}
}

// TestCreditsCombined is plan task 12.8: every credit the map owes, the
// basemap's and each overlay's, once each and in a fixed order.
func TestCreditsCombined(t *testing.T) {
	m := gulfMap(t, 80, 24)
	credits := m.Credits()
	if len(credits) == 0 || !strings.Contains(strings.Join(credits, " "), "OpenStreetMap") {
		t.Fatalf("the basemap's own credit is missing: %v", credits)
	}
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(warning("more-alerts")); err != nil {
		t.Fatal(err)
	}
	credits = m.Credits()
	if n := strings.Count(strings.Join(credits, "\n"), "National Weather Service"); n != 1 {
		t.Errorf("two overlays crediting the same service are credited %d times: %v", n, credits)
	}
	again := m.Credits()
	if strings.Join(credits, "|") != strings.Join(again, "|") {
		t.Error("the credits came back in a different order the second time")
	}
}

// TestScaleExposed is plan task 12.9: the scale as data, in the same terms
// the mark on the frame draws it.
func TestScaleExposed(t *testing.T) {
	m := gulfMap(t, 149, 38)
	scale, ok := m.Scale()
	if !ok {
		t.Fatal("a map of somewhere has no scale")
	}
	if scale.Km <= 0 || scale.Cells <= 0 || scale.KmPerCell <= 0 {
		t.Fatalf("%+v", scale)
	}
	// The bar is the round distance it says it is, within the width of a cell.
	if got := scale.KmPerCell * float64(scale.Cells); got < scale.Km*0.5 || got > scale.Km*1.5 {
		t.Errorf("a bar of %d cells at %v km a cell is %v km, and it says %v", scale.Cells, scale.KmPerCell, got, scale.Km)
	}
	// Zoomed in, a cell is a shorter distance.
	closer := scale.KmPerCell
	if err := m.Zoom(8); err != nil {
		t.Fatal(err)
	}
	after, _ := m.Scale()
	if after.KmPerCell >= closer {
		t.Errorf("a cell is %v km at zoom 8 and %v km at zoom 3", after.KmPerCell, closer)
	}
}

// TestParityP57_Footer is plan task 12.29's row: the footer is upstream's
// own wording of the centre and the zoom, cut with floor as upstream cuts
// it, and it is drawn inside the map only when the host asks.
func TestParityP57_Footer(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	if err := m.Recentre(tuimaps.LonLat{Lon: -84.567891, Lat: 26.987654}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3.97); err != nil {
		t.Fatal(err)
	}
	// Upstream cuts with floor, it does not round.
	if got, want := m.Footer(), "center: 26.987, -84.567   zoom: 3"; got != want {
		t.Errorf("Footer() is %q, want %q", got, want)
	}
	// It is off the frame by default, and drawn when the host turns it on.
	if frame := frameAtTime(t, m, cols, rows, noon); strings.Contains(frame, "center:") {
		t.Error("the footer is drawn with nobody asking for it")
	}
	m.ShowFooter(true)
	if frame := frameAtTime(t, m, cols, rows, noon); !strings.Contains(frame, "center: 26.987, -84.567") {
		t.Errorf("the footer was turned on and is not on the frame:\n%s", frame)
	}
	m.ShowFooter(false)
	if frame := frameAtTime(t, m, cols, rows, noon); strings.Contains(frame, "center:") {
		t.Error("the footer stayed on the frame after it was turned off")
	}
}

// TestFactsOnAClosedMap: the facts a map gives about itself are empty once
// it is closed, and none of them panics.
func TestFactsOnAClosedMap(t *testing.T) {
	m := world(t, 40, 12)
	m.Close()
	if len(m.Legend()) != 0 || len(m.Credits()) != 0 || m.Footer() != "" {
		t.Error("a closed map still says things about itself")
	}
	if _, ok := m.Scale(); ok {
		t.Error("a closed map has a scale")
	}
	if _, ok := m.NextCall(noon); ok {
		t.Error("a closed map is due to be called again")
	}
	_ = time.Now
}

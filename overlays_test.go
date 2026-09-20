package tuimaps_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

var noon = time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

// warning is an alert area over the Gulf coast, with its outline role and a
// plain-word label.
func warning(id string) tuimaps.Overlay {
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "National Weather Service",
		Features: []tuimaps.Feature{{
			Kind:  tuimaps.Polygon,
			Rings: [][]tuimaps.LonLat{{{Lon: -86, Lat: 24}, {Lon: -82, Lat: 24}, {Lon: -82, Lat: 28}, {Lon: -86, Lat: 28}, {Lon: -86, Lat: 24}}},
			Role:  tuimaps.AlertSevere,
			Label: "Tornado Warning",
		}}}
}

// gulfMap is a map of the Gulf coast, settled, with the embedded tiles.
func gulfMap(t *testing.T, cols, rows int) *tuimaps.Map {
	t.Helper()
	m := world(t, cols, rows)
	if err := m.Recentre(tuimaps.LonLat{Lon: -84, Lat: 26}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m
}

// TestOverlayStructsRoundTrip is plan task 12.5: each shape a host can hand
// in is set and removed through the public calls, and the frame shows it.
func TestOverlayStructsRoundTrip(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	_, before := drawn(t, m, cols, rows)

	res, err := m.Set(warning("alerts"))
	if err != nil || !res.Created {
		t.Fatalf("%+v, %v", res, err)
	}
	// One Work call prepares it; the frame then shows it.
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, after := drawn(t, m, cols, rows)
	if after == before {
		t.Error("the frame is unchanged with an alert area set")
	}
	if !strings.Contains(after, "Tornado Warning") {
		t.Errorf("the alert's label is not on the frame:\n%s", after)
	}
	if ids := m.Overlays(); len(ids) != 1 || ids[0] != "alerts" {
		t.Errorf("the overlays are %v", ids)
	}
	gone, err := m.Remove("alerts")
	if err != nil || !gone.Found {
		t.Fatalf("%+v, %v", gone, err)
	}
	_, removed := drawn(t, m, cols, rows)
	if strings.Contains(removed, "Tornado Warning") {
		t.Error("the alert outlived its removal")
	}
	if m.InUse("alerts") {
		t.Error("an overlay that was removed is still in use")
	}
}

// TestPresetOneCall is plan task 12.6 (D-69): a grid that uses a preset needs
// the preset's name and its unit and nothing more.
func TestPresetOneCall(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	_, before := drawn(t, m, cols, rows)
	grid := tuimaps.Grid{West: -100, South: 20, East: -70, North: 35, Cols: 4, Rows: 4,
		Values: []float64{20, 22, 24, 26, 21, 23, 25, 27, 22, 24, 26, 28, 23, 25, 27, 29}}
	res, err := m.Set(tuimaps.TemperatureGrid("temperature", grid, tuimaps.Celsius, noon))
	if err != nil || !res.Created {
		t.Fatalf("%+v, %v", res, err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, after := drawn(t, m, cols, rows)
	if after == before {
		t.Error("the frame is unchanged with a temperature grid set")
	}
	// The radar helper is the same shape of call, for an image.
	if _, err := m.Remove("temperature"); err != nil {
		t.Fatal(err)
	}
}

// TestLargeShapeNeverVanishes is plan task 12.28 (D-92): an overlay over the
// vertex count shows on the very next frame, with no Work call at all,
// because Set built its run index; replaced, every frame shows the old shape
// or the new, and never neither.
func TestLargeShapeNeverVanishes(t *testing.T) {
	const cols, rows = 149, 38
	m := gulfMap(t, cols, rows)
	big := coastOverlay("coast", 40_000)
	res, err := m.Set(big)
	if err != nil || !res.Created {
		t.Fatalf("%+v, %v", res, err)
	}
	// No Settle, no Work: the next frame must already show it.
	_, straightAway := drawn(t, m, cols, rows)
	_, bare := drawn(t, world(t, cols, rows), cols, rows)
	if straightAway == bare {
		t.Error("an overlay too big for the shape cache did not show on the frame after Set")
	}
	// Replaced, no frame is ever without it.
	for range 3 {
		if _, err := m.Set(coastOverlay("coast", 40_000)); err != nil {
			t.Fatal(err)
		}
		_, now := drawn(t, m, cols, rows)
		if now == bare {
			t.Error("a frame drawn while the overlay was being replaced showed neither the old shape nor the new")
		}
	}
	// Seen later at a deep zoom it is still drawn.
	if err := m.Zoom(12); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, deep := drawn(t, m, cols, rows); deep == "" {
		t.Error("nothing at all was drawn at zoom 12")
	}
}

// coastOverlay is a line overlay of many vertices: more than the count above
// which Set builds a run index.
func coastOverlay(id string, n int) tuimaps.Overlay {
	ring := make([]tuimaps.LonLat, 0, n)
	for i := range n {
		lon := -90 + 12*float64(i)/float64(n-1)
		ring = append(ring, tuimaps.LonLat{Lon: lon, Lat: 25 + 2*float64(i%7)/7})
	}
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "A survey",
		Features: []tuimaps.Feature{{Kind: tuimaps.Line, Rings: [][]tuimaps.LonLat{ring}, Role: tuimaps.Track, Label: "Coast"}}}
}

// TestOverlayMistakes: the mistakes a host makes handing in an overlay each
// give a kind from the closed list through the public call.
func TestOverlayMistakes(t *testing.T) {
	m := world(t, 80, 24)
	if _, err := m.Set(tuimaps.Overlay{}); !isKind(err, fault.InvalidID) {
		t.Errorf("an overlay with no id: %v", err)
	}
	bad := warning("alerts")
	bad.Features[0].Rings[0][0].Lat = 100
	if _, err := m.Set(bad); !isKind(err, fault.InvalidCoordinates) {
		t.Errorf("a ring off the world: %v", err)
	}
	if res, err := m.Remove("never-set"); err != nil || res.Found {
		t.Errorf("removing what was never set: %+v, %v", res, err)
	}
}

// TestOverlayWarningsReachTheHost: what the store noticed but did not refuse
// reaches the host through Warnings, once each.
func TestOverlayWarningsReachTheHost(t *testing.T) {
	m := world(t, 80, 24)
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	// An id that differs only in case from one already set is very likely a
	// mistake, and is warned about rather than refused (PL-NC-4).
	if _, err := m.Set(warning("Alerts")); err != nil {
		t.Fatal(err)
	}
	warnings := m.Warnings()
	if len(warnings) == 0 {
		t.Fatal("an id that differs only in case from one already set raised no warning")
	}
	if again := m.Warnings(); len(again) != 0 {
		t.Errorf("a warning was reported twice: %v", again)
	}
}

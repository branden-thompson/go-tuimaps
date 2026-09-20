package tuimaps_test

import (
	"context"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestTwoInstancesIndependent is plan task 12.11 (FR-27): two maps in one
// process share nothing a host did not tell them to share.
func TestTwoInstancesIndependent(t *testing.T) {
	const cols, rows = 80, 24
	one, two := world(t, cols, rows), world(t, cols, rows)
	if _, err := one.SetPlaces([]tuimaps.Place{miami()}); err != nil {
		t.Fatal(err)
	}
	if _, err := one.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	one.ColourDepth(tuimaps.NoColour)
	if err := one.Zoom(6); err != nil {
		t.Fatal(err)
	}
	if len(two.Places()) != 0 || len(two.Overlays()) != 0 {
		t.Errorf("the second map holds the first's places or overlays: %v, %v", two.Places(), two.Overlays())
	}
	if _, zoom := two.Centre(); zoom == 6 {
		t.Error("the second map followed the first's zoom")
	}
	raw, _ := drawn(t, two, cols, rows)
	if !colours.MatchString(raw) {
		t.Error("the second map lost its colour when the first was told to draw without any")
	}
	// Closing one leaves the other working.
	one.Close()
	if _, err := two.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon); err != nil {
		t.Errorf("the second map stopped working when the first closed: %v", err)
	}
	if _, err := two.Settle(context.Background()); err != nil {
		t.Errorf("the second map cannot settle: %v", err)
	}
}

// TestCloseReleasesEverything is plan task 12.12: a closed map holds nothing
// and every call says so, without panicking and without waiting.
func TestCloseReleasesEverything(t *testing.T) {
	m := world(t, 80, 24)
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SetPlaces([]tuimaps.Place{miami()}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if inside := m.Close(); inside != 0 {
		t.Errorf("Close says %d calls are inside a map nobody is calling", inside)
	}
	if len(m.Overlays()) != 0 || len(m.Places()) != 0 || len(m.Legend()) != 0 {
		t.Error("a closed map still holds things")
	}
	if m.InUse("alerts") {
		t.Error("a closed map still reports geometry in use")
	}
	if use := m.CacheUse(); use.Tiles.Held != 0 || use.Shapes.Held != 0 {
		t.Errorf("a closed map still holds caches: %+v", use)
	}
	if m.Pending() != 0 {
		t.Errorf("a closed map has %d jobs pending", m.Pending())
	}
	// And every call that can fail says the same thing.
	for name, err := range map[string]error{
		"Settle":   second(m.Settle(context.Background())),
		"Render":   second(m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)),
		"Set":      second(m.Set(warning("alerts"))),
		"Remove":   second(m.Remove("alerts")),
		"FitTo":    m.FitTo([]tuimaps.LonLat{{Lon: 1, Lat: 1}}, nil, 1),
		"Recentre": m.Recentre(tuimaps.LonLat{}),
		"Source":   m.Source("https://tiles.example.test/"),
		"Purge":    m.Purge(),
	} {
		if !isKind(err, fault.Closed) {
			t.Errorf("%s on a closed map: %v", name, err)
		}
	}
}

// second is the error of a call that answers a value and an error.
func second[T any](_ T, err error) error { return err }

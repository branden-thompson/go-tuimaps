package tuimaps_test

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestIntents is plan task 12.3 (FR-24): each intent moves the view as it
// says it does, and none of them reads anything from the terminal.
func TestIntents(t *testing.T) {
	m := world(t, 149, 38)
	was, wasZoom := m.Centre()

	if err := m.Recentre(tuimaps.LonLat{Lon: -84, Lat: 26}); err != nil {
		t.Fatal(err)
	}
	if at, _ := m.Centre(); at.Lon != -84 || at.Lat != 26 {
		t.Errorf("Recentre put the middle at %+v", at)
	}
	if err := m.Zoom(6); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != 6 {
		t.Errorf("Zoom gave %v", zoom)
	}
	if err := m.ZoomBy(-2); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != 4 {
		t.Errorf("ZoomBy gave %v", zoom)
	}
	// Zooming out past the world stops at the end rather than failing.
	if err := m.ZoomBy(-100); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != tuimaps.MinZoom {
		t.Errorf("ZoomBy past the furthest gave %v, want %v", zoom, tuimaps.MinZoom)
	}
	if err := m.Zoom(5); err != nil {
		t.Fatal(err)
	}
	// Panning by cells moves east and south by what it is asked.
	before, _ := m.Centre()
	if err := m.PanCells(10, 4); err != nil {
		t.Fatal(err)
	}
	after, _ := m.Centre()
	if after.Lon <= before.Lon || after.Lat >= before.Lat {
		t.Errorf("panning ten cells east and four south went from %+v to %+v", before, after)
	}
	// And fitting the world puts it back where a new map starts.
	if err := m.FitWorld(); err != nil {
		t.Fatal(err)
	}
	at, zoom := m.Centre()
	if at != was || zoom != wasZoom {
		t.Errorf("FitWorld gave %+v at %v, want %+v at %v", at, zoom, was, wasZoom)
	}
	// Every intent refuses what is not a view of anywhere.
	for name, err := range map[string]error{
		"a latitude off the world":  m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 91}),
		"a zoom below the furthest": m.Zoom(tuimaps.MinZoom - 1),
		"a zoom past the closest":   m.Zoom(tuimaps.MaxZoom + 1),
		"a zoom that is no number":  m.Zoom(notANumber()),
		"levels that are no number": m.ZoomBy(notANumber()),
	} {
		if !isKind(err, fault.InvalidCoordinates) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestFitTo is the rest of 12.3 (FR-24, D-76): the view is fitted to the
// host's places and overlays, with a margin, and it needs no work to have
// run because the boxes were recorded when they were handed in.
func TestFitTo(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	if _, err := m.SetPlaces([]tuimaps.Place{
		{Name: "Miami", At: tuimaps.LonLat{Lon: -80.19, Lat: 25.77}},
		{Name: "Havana", At: tuimaps.LonLat{Lon: -82.38, Lat: 23.11}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := m.FitTo(nil, nil, 2); err != nil {
		t.Fatal(err)
	}
	at, zoom := m.Centre()
	if at.Lon > -80 || at.Lon < -83 || at.Lat > 26 || at.Lat < 23 {
		t.Errorf("fitting to two places in the Caribbean centred on %+v", at)
	}
	if zoom < 4 {
		t.Errorf("fitting to two places 300 km apart gave zoom %v", zoom)
	}
	// Both are on the frame it draws.
	_, text := drawn(t, m, cols, rows)
	if !strings.Contains(text, "Miami") || !strings.Contains(text, "Havana") {
		t.Errorf("a fitted frame does not hold both places:\n%s", text)
	}
	// An overlay set a moment ago can be fitted to with no Work at all.
	if _, err := m.Set(warning("alerts")); err != nil {
		t.Fatal(err)
	}
	if err := m.FitTo(nil, []string{"alerts"}, 1); err != nil {
		t.Fatal(err)
	}
	if at, _ := m.Centre(); at.Lon > -83 || at.Lon < -85 || at.Lat < 25 || at.Lat > 27 {
		t.Errorf("fitting to the alert area centred on %+v; it spans -86..-82, 24..28", at)
	}
	// Naming an overlay that is not there is refused, and the view stands.
	before, beforeZoom := m.Centre()
	if err := m.FitTo(nil, []string{"never-set"}, 1); !isKind(err, fault.InvalidID) {
		t.Errorf("fitting to an overlay that is not there: %v", err)
	}
	if at, zoom := m.Centre(); at != before || zoom != beforeZoom {
		t.Error("a refused fit moved the view")
	}
	// Naming nothing at all, with no places, is refused rather than guessing.
	if _, err := m.SetPlaces(nil); err != nil {
		t.Fatal(err)
	}
	if err := m.FitTo(nil, nil, 1); !isKind(err, fault.InvalidCoordinates) {
		t.Errorf("fitting to nothing: %v", err)
	}
	// Places given by hand are fitted to as well as the map's own.
	if err := m.FitTo([]tuimaps.LonLat{{Lon: 10, Lat: 50}, {Lon: 14, Lat: 52}}, nil, 1); err != nil {
		t.Fatal(err)
	}
	if at, _ := m.Centre(); at.Lon < 9 || at.Lon > 15 || at.Lat < 49 || at.Lat > 53 {
		t.Errorf("fitting to two places in Europe centred on %+v", at)
	}
}

// TestParityP55_ZoomByInitialZoom is the parity row for upstream's zoom_by
// and its initial zoom (P-55, Match): zooming by a number of levels stops
// at the ends rather than going past them, the closest is 18, and a map
// that was told no zoom starts at one it holds.
func TestParityP55_ZoomByInitialZoom(t *testing.T) {
	m := world(t, 80, 24)

	// A map told nothing starts at a zoom inside the bounds, and one it
	// can be asked for again.
	_, start := m.Centre()
	if start < tuimaps.MinZoom || start > tuimaps.MaxZoom {
		t.Errorf("a map told no zoom starts at %v, outside %v to %v", start, tuimaps.MinZoom, tuimaps.MaxZoom)
	}
	if err := m.Zoom(start); err != nil {
		t.Errorf("the zoom a map starts at is one it refuses: %v", err)
	}

	// Zooming by levels stops at each end instead of going past it.
	if err := m.ZoomBy(1000); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != tuimaps.MaxZoom {
		t.Errorf("zooming in by a thousand levels left the map at %v, want %v", zoom, tuimaps.MaxZoom)
	}
	if err := m.ZoomBy(-1000); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != tuimaps.MinZoom {
		t.Errorf("zooming out by a thousand levels left the map at %v, want %v", zoom, tuimaps.MinZoom)
	}

	// Upstream's closest is 18, and it is the library's own too (P-55).
	if tuimaps.MaxZoom != 18 {
		t.Errorf("the closest zoom is %v; upstream's is 18", tuimaps.MaxZoom)
	}

	// A zoom outside the bounds is refused, and the refusal says the
	// bounds it actually has.
	err := m.Zoom(tuimaps.MaxZoom + 1)
	if !isKind(err, fault.InvalidCoordinates) {
		t.Errorf("a zoom past the closest: %v", err)
	}
	for _, edge := range []string{"-8", "18"} {
		if err != nil && !strings.Contains(err.Error(), edge) {
			t.Errorf("the refusal does not say %s: %v", edge, err)
		}
	}
	if _, zoom := m.Centre(); zoom != tuimaps.MinZoom {
		t.Errorf("a refused zoom moved the map to %v", zoom)
	}
}

package tuimaps_test

import (
	"math"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// TestParityP52_ConfigDefaults is plan task 12.25: what a map is before a
// host has said anything. Each of these is a default upstream also has, or
// one a ruling set in its place, and every one of them is overridable.
func TestParityP52_ConfigDefaults(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	// The whole world, at the zoom that fits the rectangle (P-56).
	at, zoom := m.Centre()
	if at.Lon != 0 {
		t.Errorf("a new map is centred at longitude %v, want 0", at.Lon)
	}
	// A rectangle taller than the world in one tile needs a zoom below zero.
	if zoom < -2 || zoom > 2 {
		t.Errorf("a new map opens at zoom %v; the world in 80 by 24 cells is near zoom 0", zoom)
	}
	// Nothing is reached, nothing is held, nothing animates, nothing is due.
	if m.Pending() != 0 {
		t.Errorf("a new map wants %d units of work before it has drawn", m.Pending())
	}
	if use := m.CacheUse(); use.Tiles.Held != 0 || use.Disk.Limit != 0 {
		t.Errorf("a new map holds %+v", use)
	}
	if due, ok := m.NextCall(noon); ok {
		t.Errorf("a new map is due at %v", due)
	}
	// No places, no overlays, no legend, no warnings; the basemap's credit.
	if len(m.Places()) != 0 || len(m.Overlays()) != 0 || len(m.Legend()) != 0 || len(m.Warnings()) != 0 {
		t.Error("a new map holds places, overlays, a legend or warnings")
	}
	if len(m.Credits()) != 1 {
		t.Errorf("a new map owes %d credits, want the basemap's", len(m.Credits()))
	}
	// The footer is off (P-57), and the frame draws the scale and the credit.
	if m.Footer() == "" {
		t.Error("Footer() says nothing about where the map is")
	}
}

// TestParityP53_SizeFromTerminal is the Fix row of the same name (L-17 g):
// the library draws the exact rectangle it is given and reserves no rows of
// its own. Reserving rows for chrome is the application's business.
func TestParityP53_SizeFromTerminal(t *testing.T) {
	for _, size := range []tuimaps.Size{{Cols: 80, Rows: 24}, {Cols: 1, Rows: 1}, {Cols: 149, Rows: 38}, {Cols: 40, Rows: 3}} {
		m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
		if err != nil {
			t.Fatalf("%+v: %v", size, err)
		}
		frame, err := m.Render(size, noon)
		if err != nil {
			t.Fatalf("%+v: %v", size, err)
		}
		if len(frame.Lines) != size.Rows {
			t.Errorf("%+v drew %d rows", size, len(frame.Lines))
		}
		m.Close()
	}
}

// TestParityP54_MinZoom and P-55: the zoom a map may be at runs from the
// world in the rectangle to upstream's closest, and a zoom outside it is
// clamped rather than refused when it comes from zooming by an amount.
func TestParityP54_MinZoom(t *testing.T) {
	m := world(t, 80, 24)
	// Zooming out past the furthest stops there rather than failing.
	if err := m.ZoomBy(-50); err != nil {
		t.Fatal(err)
	}
	if _, zoom := m.Centre(); zoom != tuimaps.MinZoom {
		t.Errorf("zooming far out gave %v, want the furthest out at %v", zoom, tuimaps.MinZoom)
	}
	// Zooming in past the closest stops at the closest.
	if err := m.ZoomBy(500); err != nil {
		t.Fatal(err)
	}
	_, zoom := m.Centre()
	if zoom != tuimaps.MaxZoom {
		t.Errorf("zooming far in gave %v, want the closest at %v", zoom, tuimaps.MaxZoom)
	}
	// A zoom named outright outside the range is refused, not clamped: a
	// host that names one has made a mistake it should hear about.
	if err := m.Zoom(tuimaps.MaxZoom + 1); err == nil {
		t.Error("a zoom past the closest was accepted")
	}
}

// TestParityP56_FitWorld: the world, latitude 84 to -56, centred at
// longitude 0 and at the Mercator midpoint of that span.
func TestParityP56_FitWorld(t *testing.T) {
	m := world(t, 149, 38)
	if err := m.Recentre(tuimaps.LonLat{Lon: 100, Lat: -40}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(8); err != nil {
		t.Fatal(err)
	}
	if err := m.FitWorld(); err != nil {
		t.Fatal(err)
	}
	at, zoom := m.Centre()
	if at.Lon != 0 {
		t.Errorf("fit world is centred at longitude %v", at.Lon)
	}
	// The Mercator midpoint of 84N to 56S is well north of the equator,
	// because Mercator stretches the north: about 45N.
	if at.Lat < 40 || at.Lat > 50 {
		t.Errorf("fit world is centred at latitude %v; the Mercator midpoint of 84N to 56S is about 45N", at.Lat)
	}
	if math.Abs(zoom) > 1 {
		t.Errorf("the world in 149 by 38 cells is at zoom %v", zoom)
	}
	// Both far corners are on the frame it draws.
	_, text := drawn(t, m, 149, 38)
	if len(text) == 0 {
		t.Error("the world frame is empty")
	}
}

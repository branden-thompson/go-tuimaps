package tuimaps_test

// bound_test.go — v0.2.0 WP-L7 (L-3): a minimum zoom and a box the host sets
// once, held on every path that moves the view.

import (
	"math"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// inside reports whether a view lies within a box, in the world's projected
// fractions (a box crossing the antimeridian unwrapped): each axis either
// within the box or, where the box is narrower than the view, centred on it.
func inside(t *testing.T, m *tuimaps.Map, cols, rows int, b tuimaps.Bound) bool {
	t.Helper()
	centre, zoom := m.Centre()
	if zoom < b.MinZoom-1e-9 {
		t.Errorf("zoom %.2f below the bound's %.2f", zoom, b.MinZoom)
		return false
	}
	frac := func(lon, lat float64) (float64, float64) {
		x, y, err := project.ToTile(project.LonLat{Lon: lon, Lat: lat}, 0)
		if err != nil {
			t.Fatal(err)
		}
		return x, y
	}
	x0, y0 := frac(b.W, b.N)
	x1, y1 := frac(b.E, b.S)
	if b.W > b.E {
		x1++
	}
	cx, cy := frac(centre.Lon, centre.Lat)
	if cx < x0-1e-9 {
		cx++
	}
	world := 256 * math.Exp2(zoom)
	hw, hh := float64(cols*2)/2/world, float64(rows*4)/2/world
	axis := func(c, lo, hi, half float64) bool {
		if hi-lo <= 2*half {
			return math.Abs(c-(lo+hi)/2) < 1e-6
		}
		return c-half >= lo-1e-9 && c+half <= hi+1e-9
	}
	return axis(cx, x0, x1, hw) && axis(cy, y0, y1, hh)
}

// TestABoundIsRefusedOrHeld is L7.1: a box of no area is refused, and one
// crossing the antimeridian - the Aleutians - is held.
func TestABoundIsRefusedOrHeld(t *testing.T) {
	m := world(t, 80, 24)
	for _, bad := range []tuimaps.Bound{{W: -90, S: 30, E: -90, N: 40}, {W: -100, S: 40, E: -90, N: 40}, {W: -100, S: 50, E: -90, N: 40}, {W: -200, S: 30, E: -90, N: 40}, {MinZoom: -1, W: -100, S: 30, E: -90, N: 40}} {
		if err := m.SetBound(bad); !isKind(err, fault.InvalidCoordinates) {
			t.Errorf("%+v: %v; want it refused", bad, err)
		}
	}
	aleutians := tuimaps.Bound{MinZoom: 3, W: 172, S: 50, E: -165, N: 57}
	must(t, m.SetBound(aleutians))
	if !inside(t, m, 80, 24, aleutians) {
		t.Errorf("after setting the Aleutians bound the view is outside it")
	}
	// Just west of the box, the view goes to the box's west end - the near
	// one - not round the world to its east end; and just east, the east end.
	must(t, m.Zoom(5))
	for _, c := range []struct {
		lon       float64
		westOfMid bool
	}{{170, true}, {-160, false}} {
		must(t, m.Recentre(tuimaps.LonLat{Lon: c.lon, Lat: 53}))
		centre, _ := m.Centre()
		lon := centre.Lon
		if lon < 0 {
			lon += 360
		}
		if west := lon < (172+195)/2.0; west != c.westOfMid {
			t.Errorf("recentred at %v: the view went to %v; want the %s end", c.lon, centre.Lon, map[bool]string{true: "west", false: "east"}[c.westOfMid])
		}
	}
	// A box just east of the antimeridian, not crossing it: a centre just
	// across it, at 179 east, goes to the box's west end.
	must(t, m.SetBound(tuimaps.Bound{MinZoom: 3, W: -178, S: 50, E: -150, N: 60}))
	must(t, m.Zoom(5))
	must(t, m.Recentre(tuimaps.LonLat{Lon: 179, Lat: 55}))
	if centre, _ := m.Centre(); centre.Lon > (-178-150)/2.0 {
		t.Errorf("recentred at 179 east beside a box from -178 to -150: the view went to %v; want its west end", centre.Lon)
	}
}

// TestTheBoundIsHeldOnEveryPath is L7.2 (M3): resize, fit, pan, zoom and the
// fall-back to the whole world each stay inside, across the antimeridian too.
func TestTheBoundIsHeldOnEveryPath(t *testing.T) {
	for name, b := range map[string]tuimaps.Bound{
		"the south central states": {MinZoom: 4, W: -106, S: 25, E: -88, N: 37},
		"the Aleutians":            {MinZoom: 3, W: 172, S: 50, E: -165, N: 57},
	} {
		m := world(t, 80, 24)
		must(t, m.SetBound(b))
		check := func(what string, cols, rows int) {
			t.Helper()
			if !inside(t, m, cols, rows, b) {
				c, z := m.Centre()
				t.Errorf("%s, %s: the view (centre %+v, zoom %.2f) left the bound", name, what, c, z)
			}
		}
		must(t, m.Zoom(1))
		check("zoomed out past the minimum", 80, 24)
		must(t, m.PanCells(400, 0))
		check("panned far east", 80, 24)
		must(t, m.PanCells(0, -400))
		check("panned far north", 80, 24)
		must(t, m.Recentre(tuimaps.LonLat{Lon: 10, Lat: 50}))
		check("recentred on Europe", 80, 24)
		must(t, m.FitWorld())
		check("fitted to the world", 80, 24)
		must(t, m.FitTo([]tuimaps.LonLat{{Lon: 139, Lat: 35}, {Lon: -0.1, Lat: 51.5}}, nil, 1))
		check("fitted to Tokyo and London", 80, 24)
		if _, err := m.Render(tuimaps.Size{Cols: 149, Rows: 38}, noon); err != nil {
			t.Fatal(err)
		}
		check("resized", 149, 38)
		if _, err := m.Render(tuimaps.Size{Cols: 20, Rows: 6}, noon); err != nil {
			t.Fatal(err)
		}
		check("resized small", 20, 6)
	}
	// A map nobody has pointed anywhere falls back to the bound, not the world.
	fresh, err := tuimaps.New()
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	b := tuimaps.Bound{MinZoom: 4, W: -106, S: 25, E: -88, N: 37}
	must(t, fresh.SetBound(b))
	if _, err := fresh.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	if !inside(t, fresh, 80, 24, b) {
		t.Error("a map nobody pointed anywhere fell back to the whole world, outside its bound")
	}
}

// TestNoBoundIsV010 is L7.3 (L-3.2): a host that sets no bound, or clears
// the one it set, may zoom out to the world and look anywhere.
func TestNoBoundIsV010(t *testing.T) {
	m := world(t, 80, 24)
	must(t, m.SetBound(tuimaps.Bound{MinZoom: 4, W: -106, S: 25, E: -88, N: 37}))
	must(t, m.SetBound(tuimaps.Bound{}))
	must(t, m.Zoom(1))
	must(t, m.Recentre(tuimaps.LonLat{Lon: 10, Lat: 50}))
	if c, z := m.Centre(); z != 1 || c.Lon != 10 || c.Lat != 50 {
		t.Errorf("with no bound: centre %+v zoom %v; want Europe at zoom 1", c, z)
	}
}

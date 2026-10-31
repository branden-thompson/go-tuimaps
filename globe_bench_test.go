package tuimaps_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// globeField is a 2° whole-globe field shaped as a recorded MUF map is
// (watchpost D-261: a field varying with latitude alone drew straight
// contours, easier than any real one): highest on the day side, its crests
// either side of the equator, structure along the longitudes and two closed
// patches, 4 to 30 MHz - every class of the MUF preset somewhere.
func globeField(overWater bool) tuimaps.Grid {
	g := tuimaps.Grid{West: -180, South: -90, East: 180, North: 90, Cols: 180, Rows: 90, Values: make([]float64, 180*90), Lines: true, OverWater: overWater}
	r := math.Pi / 180
	for i := range g.Values {
		lat, lon := 89-2*float64(i/180), -179+2*float64(i%180)
		day := 1 + 0.6*math.Cos((lon-30)*r)
		crests := 1 + 0.25*math.Cos(4*lat*r)
		v := 6 + 14*math.Cos(lat*r)*day*crests + 2*math.Sin(4*lon*r)*math.Cos(3*lat*r)
		v += 5*math.Exp(-((lat-35)*(lat-35)+(lon+100)*(lon+100))/150) - 4*math.Exp(-((lat+30)*(lat+30)+(lon-140)*(lon-140))/200)
		g.Values[i] = min(max(v, 4), 30)
	}
	return g
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// reachField is a frequency's reach as a host hands it in: a whole-globe
// field of two classes - reached, 1, in a band of latitudes and longitudes
// around the origin, and not, 0 - asking for its edge alone (L-2.6).
func reachField() tuimaps.Grid {
	g := tuimaps.Grid{West: -180, South: -90, East: 180, North: 90, Cols: 180, Rows: 90, Values: make([]float64, 180*90), OverWater: true}
	for i := range g.Values {
		lat, lon := 89-2*float64(i/180), -179+2*float64(i%180)
		if abs(lat-20) < 35 && abs(lon) < 60 && !(abs(lat-20) < 8 && abs(lon) < 8) {
			g.Values[i] = 1
		}
	}
	g.Type = tuimaps.Type{Breaks: []float64{0.5}, Edge: true}
	return g
}

// globePan is M5's frame (NFR-1, watchpost D-209): a map of size over the
// whole world under a 2° field filled in its colours and drawn with its
// contour lines, over the sea or stopping at the shore, with the night
// shading the ground and its terminator, and a reach's edge over it, settled.
// step is the i-th frame of a one-cell pan back and forth, drawn.
func globePan(tb testing.TB, size tuimaps.Size, overWater bool) (step func(i int)) {
	tb.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { m.Close() })
	if err := m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 20}); err != nil {
		tb.Fatal(err)
	}
	if err := m.Zoom(1); err != nil {
		tb.Fatal(err)
	}
	field := globeField(overWater)
	reach := reachField()
	night := tuimaps.Overlay{ID: "night", Valid: noon, Keeps: 3600e9, Features: []tuimaps.Feature{
		{Kind: tuimaps.Polygon, Rings: tuimaps.NightSide(noon), Role: tuimaps.NightTint},
		{Kind: tuimaps.Line, Rings: [][]tuimaps.LonLat{tuimaps.Terminator(noon)}, Role: tuimaps.TerminatorLine},
	}}
	for _, o := range []tuimaps.Overlay{tuimaps.MUFGrid("muf", field, noon), night, {ID: "reach", Valid: noon, Keeps: 3600e9, Grid: &reach}} {
		if _, err := m.Set(o); err != nil {
			tb.Fatal(err)
		}
	}
	if _, err := m.Settle(context.Background()); err != nil {
		tb.Fatal(err)
	}
	return func(i int) {
		step := 1
		if i%2 == 1 {
			step = -1
		}
		if err := m.PanCells(step, 0); err != nil {
			tb.Fatal(err)
		}
		if _, err := m.Render(size, noon); err != nil {
			tb.Fatal(err)
		}
	}
}

// m5Sizes are the sizes M5 is held at, each with its target in milliseconds a
// frame, and 800 by 220, recorded as a known limit, not M5's (watchpost D-261).
var m5Sizes = []struct {
	size   tuimaps.Size
	target float64
}{{tuimaps.Size{Cols: 400, Rows: 110}, 16}, {tuimaps.Size{Cols: 200, Rows: 56}, 8}, {tuimaps.Size{Cols: 800, Rows: 220}, 0}}

// waterName names a field's choice over water.
var waterName = map[bool]string{true: "over water", false: "stopping at the shore"}

// BenchmarkGlobeFieldPan is NFR-1's shape (M5, watchpost D-209): globePan's
// frame at each of m5Sizes, over the sea and stopping at the shore - the frame
// a host's Propagation map changes.
func BenchmarkGlobeFieldPan(b *testing.B) {
	for _, s := range m5Sizes {
		for _, overWater := range []bool{true, false} {
			b.Run(fmt.Sprintf("%dx%d %s", s.size.Cols, s.size.Rows, waterName[overWater]), func(b *testing.B) {
				step := globePan(b, s.size, overWater)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; b.Loop(); i++ {
					step(i)
				}
			})
		}
	}
}

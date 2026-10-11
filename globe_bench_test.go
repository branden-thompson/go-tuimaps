package tuimaps_test

import (
	"context"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// globeField is a 2° whole-globe field rising from the poles to the
// equator, 4 to 30 MHz: every class of the MUF preset somewhere.
func globeField(overWater bool) tuimaps.Grid {
	g := tuimaps.Grid{West: -180, South: -90, East: 180, North: 90, Cols: 180, Rows: 90, Values: make([]float64, 180*90), Lines: true, OverWater: overWater}
	for i := range g.Values {
		lat := 89 - 2*float64(i/180)
		g.Values[i] = 4 + 26*(1-abs(lat)/90)
	}
	return g
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkGlobeFieldPan is NFR-1's shape (M5): a one-cell pan over the
// whole world under a 2° field drawn as lines over its colours, over the
// sea and stopping at the shore - what a host's changed frame costs.
func BenchmarkGlobeFieldPan(b *testing.B) {
	for _, overWater := range []bool{true, false} {
		name := map[bool]string{true: "over water", false: "stopping at the shore"}[overWater]
		b.Run(name, func(b *testing.B) {
			m, err := tuimaps.New(tuimaps.WithSize(149, 38))
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			if err := m.Recentre(tuimaps.LonLat{Lon: 0, Lat: 20}); err != nil {
				b.Fatal(err)
			}
			if err := m.Zoom(1); err != nil {
				b.Fatal(err)
			}
			if _, err := m.Set(tuimaps.MUFGrid("muf", globeField(overWater), noon)); err != nil {
				b.Fatal(err)
			}
			if _, err := m.Settle(context.Background()); err != nil {
				b.Fatal(err)
			}
			size := tuimaps.Size{Cols: 149, Rows: 38}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				step := 1
				if i%2 == 1 {
					step = -1
				}
				if err := m.PanCells(step, 0); err != nil {
					b.Fatal(err)
				}
				if _, err := m.Render(size, noon); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

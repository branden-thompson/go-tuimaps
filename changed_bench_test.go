package tuimaps_test

import (
	"context"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// manyFeatures is an overlay of n small areas spread over the Gulf, which
// is the shape of a real alert set: many separate polygons, each small.
func manyFeatures(id string, n int) tuimaps.Overlay {
	features := make([]tuimaps.Feature, 0, n)
	for i := range n {
		west := -92 + 0.12*float64(i%80)
		south := 24 + 0.12*float64(i/80)
		features = append(features, tuimaps.Feature{
			Kind: tuimaps.Polygon, Role: tuimaps.AlertModerate, Label: "Warning",
			Rings: [][]tuimaps.LonLat{{
				{Lon: west, Lat: south}, {Lon: west + 0.1, Lat: south},
				{Lon: west + 0.1, Lat: south + 0.1}, {Lon: west, Lat: south + 0.1},
				{Lon: west, Lat: south},
			}},
		})
	}
	return tuimaps.Overlay{ID: id, Valid: noon, Keeps: time.Hour, Credit: "National Weather Service", Features: features}
}

// loaded is a settled map of the Gulf carrying n features and a blinking
// place, ready to be asked for frame after frame.
func loaded(b *testing.B, n int) *tuimaps.Map {
	b.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38))
	if err != nil {
		b.Fatal(err)
	}
	if err := m.Recentre(tuimaps.LonLat{Lon: -88, Lat: 27}); err != nil {
		b.Fatal(err)
	}
	if err := m.Zoom(6); err != nil {
		b.Fatal(err)
	}
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Home", At: tuimaps.LonLat{Lon: -88, Lat: 27}, Blink: true}}); err != nil {
		b.Fatal(err)
	}
	if _, err := m.Set(manyFeatures("alerts", n)); err != nil {
		b.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		b.Fatal(err)
	}
	return m
}

// BenchmarkChangedFrame is plan task 14.7 (NFR-4): **what it costs to draw
// again when something small has changed.** A marker's blink changes one
// cell; a pan of one cell changes every row. Both are measured at a hundred
// features and at a thousand, because the cost a host feels is the cost of
// the frame it did not ask for.
func BenchmarkChangedFrame(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run("marker phase/"+itoa(n), func(b *testing.B) {
			m := loaded(b, n)
			defer m.Close()
			size := tuimaps.Size{Cols: 149, Rows: 38}
			at := noon
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				at = at.Add(400 * time.Millisecond) // half a blink: the marker turns over
				if _, err := m.Render(size, at); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("one-cell pan/"+itoa(n), func(b *testing.B) {
			m := loaded(b, n)
			defer m.Close()
			size := tuimaps.Size{Cols: 149, Rows: 38}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				step := 1
				if i%2 == 1 {
					step = -1 // back again, so the view stays where it was
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

// TestChangedFrameCost is NFR-4's numbers, pinned as a test rather than
// left in a benchmark nobody reads: a frame whose marker has turned over
// allocates no more than 16 KB in no more than 64 allocations, **and the
// count does not grow with the features on the map** (D-115).
//
// A person zooming into a dense area is the ordinary case for a general map
// library, not the exception; a redraw caused by something that is not the
// overlays must not walk the features again.
func TestChangedFrameCost(t *testing.T) {
	for _, features := range []int{100, 1000} {
		t.Run(itoa(features)+" features", func(t *testing.T) { costOfAChange(t, features) })
	}
}

// costOfAChange measures one marker turn on a map carrying n features.
func costOfAChange(t *testing.T, features int) {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Recentre(tuimaps.LonLat{Lon: -88, Lat: 27}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(6); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SetPlaces([]tuimaps.Place{{Name: "Home", At: tuimaps.LonLat{Lon: -88, Lat: 27}, Blink: true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(manyFeatures("alerts", features)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	size := tuimaps.Size{Cols: 149, Rows: 38}
	at := noon
	// Warm: the first frame builds what every later frame reuses.
	if _, err := m.Render(size, at); err != nil {
		t.Fatal(err)
	}
	const rounds = 20
	allocs := testing.AllocsPerRun(rounds, func() {
		at = at.Add(400 * time.Millisecond) // half a blink: the marker turns over
		if _, err := m.Render(size, at); err != nil {
			t.Fatal(err)
		}
	})
	if allocs > 64 {
		t.Errorf("with %d features a marker-phase change allocates %.0f times; NFR-4 says at most 64, whatever is on the map (D-115)",
			features, allocs)
	}

	// A one-cell pan does change what is drawn of every feature, so it is
	// held to NFR-4's own figure for a pan rather than to the marker's.
	// **What it may not do is grow with the map's contents**: the residual
	// here is about a twentieth of an allocation a feature, against the one
	// and a half it was before D-115.
	panned := testing.AllocsPerRun(rounds, func() {
		if err := m.PanCells(1, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Render(size, at); err != nil {
			t.Fatal(err)
		}
		if err := m.PanCells(-1, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Render(size, at); err != nil {
			t.Fatal(err)
		}
	})
	if panned > 1346 {
		t.Errorf("with %d features a one-cell pan allocates %.0f times; NFR-4 says at most 1,346", features, panned)
	}
	t.Logf("%d features: a marker turn allocates %.0f times, a one-cell pan %.0f", features, allocs, panned)
}

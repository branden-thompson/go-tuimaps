//go:build !race

package tuimaps_test

// idle_bench_test.go — W14 (the host's performance pass): what Map.Render
// costs when nothing has changed. The renderer's own unchanged frame is pinned
// at zero allocations (internal/render/allocs_test.go), but everything
// Map.Render does before it reaches the renderer - planning the tiles, the
// overlays' state, the report's clock - was pinned by nothing. A host renders
// on every Update, so this is the cost an idle map pays for every key, tick
// and data landing. The allocations are gated; the time is recorded, never
// gated (D-53). Built without the race detector, which allocates on its own
// account.

import (
	"context"
	"image/color"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// idleFrameAllocs is the budget an unchanged Map.Render may allocate: set
// from the first measurement, and only ever lowered by the work that follows.
const idleFrameAllocs = 63 // measured 2026-09-30 at 3.8 µs a frame (W14): pinned, lowered only

// idling is a 149x38 map over the embedded tiles with a stopped twelve-frame
// radar loop and an alert area - the host's map at rest - drawn once and
// settled; render draws it again, with nothing changed.
func idling(tb testing.TB) (render func()) {
	tb.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { m.Close() })
	var frames []tuimaps.LoopFrame
	for i := 11; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(tb, 298, 152, color.NRGBA{R: 200, A: 255})})
	}
	if _, err := m.Set(tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -110, South: 25, East: -70, North: 50,
		Projection: tuimaps.PlateCarree, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 45}}, Exact: true}, noon)); err != nil {
		tb.Fatal(err)
	}
	ring := []tuimaps.LonLat{{Lon: -100, Lat: 35}, {Lon: -95, Lat: 35}, {Lon: -95, Lat: 39}, {Lon: -100, Lat: 39}, {Lon: -100, Lat: 35}}
	if _, err := m.Set(tuimaps.Overlay{ID: "alert/1", Valid: noon.Add(-time.Hour), Keeps: 6 * time.Hour,
		Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Rings: [][]tuimaps.LonLat{ring}, Role: tuimaps.AlertSevere, Label: "Wind Warning", Severity: tuimaps.SeveritySevere, ID: "1"}}}); err != nil {
		tb.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		tb.Fatal(err)
	}
	size := tuimaps.Size{Cols: 149, Rows: 38}
	for range 3 { // drawn and settled: every later draw is the same frame
		if _, err := m.Render(size, noon); err != nil {
			tb.Fatal(err)
		}
	}
	return func() {
		if _, err := m.Render(size, noon); err != nil {
			tb.Fatal(err)
		}
	}
}

// TestAnIdleFrameCostsItsBudget holds the allocations of an unchanged
// Map.Render to idleFrameAllocs.
func TestAnIdleFrameCostsItsBudget(t *testing.T) {
	render := idling(t)
	allocs := testing.AllocsPerRun(50, render)
	t.Logf("an idle Map.Render: %.0f allocations", allocs)
	if allocs > idleFrameAllocs {
		t.Errorf("an idle Map.Render cost %.0f allocations; its budget is %d", allocs, idleFrameAllocs)
	}
}

// BenchmarkIdleFrame is the time an unchanged Map.Render takes: recorded,
// never gated.
func BenchmarkIdleFrame(b *testing.B) {
	render := idling(b)
	b.ReportAllocs()
	for b.Loop() {
		render()
	}
}

package tuimaps_test

// advance_bench_test.go — v0.2.0 L3.14 (L-12.5, D-61): what one frame
// advance costs at 149x38 over a twelve-frame loop at the view's dot grid.
// The time is measured on the reference machine and recorded at SHIP, never
// gated - timing in a gate is flaky (D-53); the gate holds the allocations.

import (
	"context"
	"image/color"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// advancing is a 149x38 map over the embedded tiles with a playing
// twelve-frame loop, every other frame rain, so that each advance redraws;
// step draws the frame one advance on.
func advancing(tb testing.TB) (step func()) {
	tb.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { m.Close() })
	var frames []tuimaps.LoopFrame
	for i := 11; i >= 0; i-- {
		c := color.NRGBA{R: 200, A: 255}
		if i%2 == 0 {
			c = color.NRGBA{A: 0}
		}
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(tb, 298, 152, c)})
	}
	if _, err := m.Set(tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -110, South: 25, East: -70, North: 50,
		Projection: tuimaps.PlateCarree, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 45}}, Exact: true}, noon)); err != nil {
		tb.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		tb.Fatal(err)
	}
	if err := m.SetPlayback(tuimaps.PlaybackOn); err != nil {
		tb.Fatal(err)
	}
	size := tuimaps.Size{Cols: 149, Rows: 38}
	at := noon
	m.Animate(at)
	if err := m.Play(); err != nil {
		tb.Fatal(err)
	}
	if _, err := m.Render(size, noon); err != nil {
		tb.Fatal(err)
	}
	return func() {
		at = at.Add(500 * time.Millisecond)
		m.Animate(at)
		if _, err := m.Render(size, noon); err != nil {
			tb.Fatal(err)
		}
	}
}

// BenchmarkFrameAdvance is L-12.5's measure: one advance, drawn. Its time is
// recorded at SHIP on the reference machine, against 15 ms.
func BenchmarkFrameAdvance(b *testing.B) {
	step := advancing(b)
	for b.Loop() {
		step()
	}
}

// TestAFrameAdvanceAllocates is the gate's hold on L3.14: the allocations
// one advance makes, pinned so that a change that makes every frame allocate
// more is seen.
func TestAFrameAdvanceAllocates(t *testing.T) {
	step := advancing(t)
	if n := testing.AllocsPerRun(24, step); n > frameAdvanceAllocs {
		t.Errorf("a frame advance allocates %.0f times; pinned at %d", n, frameAdvanceAllocs)
	}
}

// frameAdvanceAllocs is the pin: 469 measured on 2026-09-25 (most of them
// the rows the advance redraws), with about ten per cent of room for noise.
// The same run measured 3.8 ms an advance on the reference machine.
const frameAdvanceAllocs = 520

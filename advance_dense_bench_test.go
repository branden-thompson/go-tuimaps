package tuimaps_test

// advance_dense_bench_test.go — v0.2.0 OW-22 (L-12.5): one frame advance on
// a dense map, where BenchmarkFrameAdvance draws a near-empty one - a region
// at zoom 6, a twelve-frame web map loop of noisy rain in six classes, a
// thousand alert areas and twenty places in view. Its time is recorded on
// the reference machine at VALIDATE, against 15 ms, and never gated (D-53).

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// denseClasses are six rain colours, light to heavy.
var denseClasses = []tuimaps.RGB{{G: 200}, {R: 40, G: 160}, {R: 200, G: 200}, {R: 230, G: 120}, {R: 220}, {R: 200, B: 200}}

// noisyFrame is a w x h picture of rain in the six classes, every pixel
// drawn at random, a third of them clear.
func noisyFrame(tb testing.TB, rng *rand.Rand, w, h int) []byte {
	tb.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			if k := rng.IntN(9); k < 6 {
				c := denseClasses[k]
				img.SetNRGBA(x, y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		tb.Fatal(err)
	}
	return buf.Bytes()
}

// denseAdvancing is the dense map, playing; step draws the frame one
// advance on.
func denseAdvancing(tb testing.TB) (step func()) {
	tb.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { m.Close() })
	if err := m.Recentre(tuimaps.LonLat{Lon: -87, Lat: 35}); err != nil {
		tb.Fatal(err)
	}
	if err := m.Zoom(6); err != nil {
		tb.Fatal(err)
	}
	rng := rand.New(rand.NewPCG(7, 9))
	var frames []tuimaps.LoopFrame
	for i := 11; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: noisyFrame(tb, rng, 298, 152)})
	}
	var table []tuimaps.TableEntry
	for i, c := range denseClasses {
		table = append(table, tuimaps.TableEntry{Colour: c, Value: 10 + 10*float64(i)})
	}
	if _, err := m.Set(tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -92, South: 32, East: -82, North: 38,
		Projection: tuimaps.WebMercator, Table: table, Exact: true}, noon)); err != nil {
		tb.Fatal(err)
	}
	alerts := tuimaps.Overlay{ID: "alerts", Valid: noon, Keeps: time.Hour}
	roles := []tuimaps.Token{tuimaps.AlertExtreme, tuimaps.AlertSevere, tuimaps.AlertModerate, tuimaps.AlertMinor}
	for i := range 1000 {
		lon, lat := -92+10*rng.Float64(), 32+6*rng.Float64()
		alerts.Features = append(alerts.Features, tuimaps.Feature{Kind: tuimaps.Polygon, Role: roles[i%4], Label: "Warning " + strconv.Itoa(i),
			Rings: [][]tuimaps.LonLat{{{Lon: lon, Lat: lat}, {Lon: lon + 0.3, Lat: lat}, {Lon: lon + 0.3, Lat: lat + 0.2}, {Lon: lon, Lat: lat + 0.2}, {Lon: lon, Lat: lat}}}})
	}
	if _, err := m.Set(alerts); err != nil {
		tb.Fatal(err)
	}
	var places []tuimaps.Place
	for i := range 20 {
		places = append(places, tuimaps.Place{Name: "Place " + strconv.Itoa(i), At: tuimaps.LonLat{Lon: -91 + 0.4*float64(i), Lat: 33 + 0.2*float64(i%10)}})
	}
	if _, err := m.SetPlaces(places); err != nil {
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

// BenchmarkFrameAdvanceDense is OW-22's measure: one advance of the dense
// map, drawn.
func BenchmarkFrameAdvanceDense(b *testing.B) {
	step := denseAdvancing(b)
	for b.Loop() {
		step()
	}
}

// TestADenseAdvanceDraws holds the dense map to drawing: every advance
// renders without error, and the benchmark measures a map that draws.
func TestADenseAdvanceDraws(t *testing.T) {
	step := denseAdvancing(t)
	for range 3 {
		step()
	}
}

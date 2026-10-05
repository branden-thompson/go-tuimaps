package tuimaps

// blink_test.go — v0.2.0 L11.35 (L-29; watchpost W14 P-10): the blink phase
// redraws only what blinks.

import (
	"context"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/render"
)

// TestAPhaseFlipRedrawsOnlyWhatBlinks is L-29.1: a map whose places are all
// steady draws nothing new when the blink phase flips - the frame is the same
// in either half - and a map with a blinking place is redrawn.
func TestAPhaseFlipRedrawsOnlyWhatBlinks(t *testing.T) {
	m, err := New(WithSize(100, 30), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.SetPlaces([]Place{{ID: "here", Name: "Here", At: LonLat{Lon: -98, Lat: 38}, Marker: MarkerDot}}); err != nil {
		t.Fatal(err)
	}
	size := Size{Cols: 100, Rows: 30}
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	draw := func() {
		t.Helper()
		if _, err := m.Render(size, at); err != nil {
			t.Fatal(err)
		}
	}
	draw()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = m.Settle(ctx)
	draw()
	before := Redraws(m)
	for range 4 { // four half-periods: the phase flips at each
		at = at.Add(render.BlinkPeriod / 2)
		draw()
	}
	if got := Redraws(m) - before; got != 0 {
		t.Errorf("the phase flipped four times with nothing blinking and the map was redrawn %d times", got)
	}

	if _, err := m.SetPlaces([]Place{{ID: "here", Name: "Here", At: LonLat{Lon: -98, Lat: 38}, Marker: MarkerDot, Blink: true}}); err != nil {
		t.Fatal(err)
	}
	draw()
	before = Redraws(m)
	at = at.Add(render.BlinkPeriod / 2)
	draw()
	if Redraws(m) == before {
		t.Error("a blinking place's phase flipped and the map was not redrawn")
	}
}

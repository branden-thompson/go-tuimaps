package tuimaps_test

// refresh_test.go — the BUILD-exit red team's perf F1: after a loop is
// refreshed and the map settles, the new frames are drawn, not the
// stand-in kept from the old ones.

import (
	"image/color"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// colouredLoop is a loop of n solid frames of one colour, five minutes
// apart, ending at noon, read with a table of two classes.
func colouredLoop(t *testing.T, n int, c color.NRGBA) tuimaps.Overlay {
	t.Helper()
	var frames []tuimaps.LoopFrame
	for i := n - 1; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(t, 160, 96, c)})
	}
	return tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -90, South: 30, East: -80, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{G: 200}, Value: 20}, {Colour: tuimaps.RGB{R: 200}, Value: 55}}, Exact: true}, noon)
}

// TestARefreshedLoopDrawsItsNewFrames: heavy frames, then a refresh to
// light ones; once the map settles, the light class is drawn and the heavy
// is gone.
func TestARefreshedLoopDrawsItsNewFrames(t *testing.T) {
	m := world(t, 80, 24)
	m.ColourDepth(tuimaps.Truecolor)
	heavy, light := color.NRGBA{R: 200, A: 255}, color.NRGBA{G: 200, A: 255}
	if _, err := m.Set(colouredLoop(t, 12, heavy)); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	var classes []tuimaps.Class
	for _, e := range m.Legend() {
		if e.Preset == "radar" {
			classes = e.Classes
		}
	}
	before := momentFrame(t, m, noon)
	if _, err := m.Set(colouredLoop(t, 13, light)); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	after := momentFrame(t, m, noon)
	if before == after {
		t.Fatalf("after a refresh from heavy frames to light ones and a settle, the frame drawn is unchanged: the old frames are still shown (%d radar grounds)", groundsOf(classes, after))
	}
}

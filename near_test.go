package tuimaps_test

// near_test.go — OW-10's wiring (L-2.2, D-104): a near colour reaches the
// host as a warning.

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestANearColourReachesTheHost is OW-10's wiring (L-2.2, D-104): a picture a
// few steps off its own table's colours is read, and the host is told with
// near-image-colours and the count.
func TestANearColourReachesTheHost(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := range 10 {
		for x := range 10 {
			img.Set(x, y, color.NRGBA{R: 43, G: 160, B: 40, A: 255}) // three steps off the table's green
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	m := world(t, 80, 24)
	mustSet(t, m, tuimaps.RadarImage("radar", tuimaps.Image{PNG: buf.Bytes(), West: -95, South: 30, East: -85, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 40, G: 160, B: 40}, Value: 20}, {Colour: tuimaps.RGB{R: 200, G: 30, B: 30}, Value: 50}}}, noon))
	settle(t, m)
	for _, w := range m.Warnings() {
		if w.Kind == tuimaps.NearImageColours && w.Count == 100 {
			return
		}
	}
	t.Errorf("no near-image-colours warning for 100 pixels: %+v", m.Warnings())
}

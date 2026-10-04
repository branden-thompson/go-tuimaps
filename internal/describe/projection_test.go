package describe

import (
	"math"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// webRow is the row a latitude falls in, in equal steps of the web map's
// height between north and south - worked out here from the projection's
// definition, not from the code under test.
func webRow(lat, north, south float64, rows int) int {
	h := func(l float64) float64 { return math.Atanh(math.Sin(l * math.Pi / 180)) }
	return int((h(north) - h(lat)) / (h(north) - h(south)) * float64(rows))
}

// A WEB MERCATOR IMAGE IS READ IN ITS OWN PROJECTION (F-3): its rows are equal
// steps of the web map's height, not of latitude, so at 60 north on an image
// from the equator to 70 north the class is the one the renderer draws there.
func TestAWebMercatorImageIsReadInItsOwnProjection(t *testing.T) {
	classes := make([]int8, 10)
	for i := range classes {
		classes[i] = int8(i)
	}
	img := Image{West: -10, South: 0, East: 10, North: 70, Width: 1, Height: 10, Classes: classes, Mercator: true}
	for _, lat := range []float64{5, 30, 45, 58, 60, 69} {
		got, ok := img.ClassAt(project.LonLat{Lon: 0, Lat: lat})
		if want := webRow(lat, 70, 0, 10); !ok || got != want {
			t.Errorf("at %v north: class %d (%v), want %d", lat, got, ok, want)
		}
	}
	flat := img
	flat.Mercator = false
	if got, _ := flat.ClassAt(project.LonLat{Lon: 0, Lat: 60}); got != 1 {
		t.Errorf("a plate carree image at 60 north: class %d, want 1", got)
	}
}

// A PIXEL'S CENTRE IS IN ITS PIXEL, in either projection: the centre a
// sighting is placed at reads back as the pixel it came from.
func TestAPixelsCentreIsInItsPixel(t *testing.T) {
	for _, merc := range []bool{false, true} {
		img := Image{West: -10, South: 0, East: 10, North: 70, Width: 4, Height: 12, Classes: make([]int8, 48), Mercator: merc}
		for y := range img.Height {
			for x := range img.Width {
				if px, py, ok := img.pixel(img.centre(x, y)); !ok || px != x || py != y {
					t.Errorf("mercator %v: the centre of (%d,%d) reads as (%d,%d) %v", merc, x, y, px, py, ok)
				}
			}
		}
	}
}

// TestTrackingKeepsEachFramesCells (BUILD-exit red team, perf F2): Track
// keeps the cells it works out on the frames, so a second place tracked
// over the same frames shares the flood fill.
func TestTrackingKeepsEachFramesCells(t *testing.T) {
	cls := make([]int8, 100)
	cls[55] = 5
	frames := []Frame{{Valid: time.Unix(0, 0), Image: Image{West: 0, South: 0, East: 1, North: 1, Width: 10, Height: 10, Classes: cls}},
		{Valid: time.Unix(300, 0), Image: Image{West: 0, South: 0, East: 1, North: 1, Width: 10, Height: 10, Classes: cls}}}
	if _, _, ok := Track(frames, 3, project.LonLat{Lon: 0.5, Lat: 0.5}); !ok {
		t.Fatal("no motion tracked over a held cell")
	}
	for i, f := range frames {
		if !f.celled || f.threshold != 3 || len(f.cells) != 1 {
			t.Errorf("frame %d kept no cells after tracking (celled %v, %d cells)", i, f.celled, len(f.cells))
		}
	}
}

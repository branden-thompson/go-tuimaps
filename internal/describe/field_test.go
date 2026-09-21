package describe

import (
	"math"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// warmEast is a field that rises from west to east: 10, 20, 30 across, the
// same at every latitude.
func warmEast() Grid {
	return Grid{West: -10, South: -10, East: 20, North: 10, Cols: 3, Rows: 2,
		Values: []float64{10, 20, 30, 10, 20, 30}}
}

// TestFieldValueBandAndRise is plan task 11.6 (D-68): the value at a place,
// the class it falls in, and which way the field rises - exact on a flat
// day, where the answer is that it does not rise at all.
func TestFieldValueBandAndRise(t *testing.T) {
	breaks := []float64{0, 15, 25}
	got := warmEast().At(project.LonLat{Lon: 0, Lat: 0}, breaks)
	if got.NoData {
		t.Fatal("the middle of the field has no data")
	}
	if got.Value != 20 {
		t.Errorf("the value is %v, want the sample's own 20", got.Value)
	}
	if got.Band != 2 {
		t.Errorf("20 is in band %d; the breaks are 0, 15 and 25", got.Band)
	}
	if !got.Rising {
		t.Fatal("a field that rises from west to east is said not to rise")
	}
	if math.Abs(got.Rises-90) > 1 {
		t.Errorf("it rises towards bearing %v, want due east", got.Rises)
	}
	// The western sample: its class is lower, and it still rises east.
	west := warmEast().At(project.LonLat{Lon: -9, Lat: 0}, breaks)
	if west.Value != 10 || west.Band != 1 {
		t.Errorf("the western sample is %v in band %d", west.Value, west.Band)
	}
	// A flat day: every value the same. Nothing rises, and that is the
	// answer rather than a bearing made up from rounding (D-68).
	flat := Grid{West: -10, South: -10, East: 10, North: 10, Cols: 2, Rows: 2, Values: []float64{7, 7, 7, 7}}
	still := flat.At(project.LonLat{Lon: 0, Lat: 0}, breaks)
	if still.NoData || still.Value != 7 {
		t.Errorf("a flat day reads %+v", still)
	}
	if still.Rising {
		t.Errorf("a flat day rises towards %v", still.Rises)
	}
	// A field that rises north: bearing 0.
	northward := Grid{West: -10, South: -10, East: 10, North: 10, Cols: 2, Rows: 2, Values: []float64{30, 30, 10, 10}}
	up := northward.At(project.LonLat{Lon: 0, Lat: -5}, breaks)
	if !up.Rising || math.Abs(up.Rises) > 1 {
		t.Errorf("a field that rises north rises towards %v (rising %v)", up.Rises, up.Rising)
	}
	// Outside the grid is no data, said plainly.
	if out := warmEast().At(project.LonLat{Lon: 100, Lat: 0}, breaks); !out.NoData {
		t.Errorf("a place outside the grid reads %+v", out)
	}
	// A value that is not a number is no data, not a reading of zero.
	broken := Grid{West: -10, South: -10, East: 10, North: 10, Cols: 1, Rows: 1, Values: []float64{math.NaN()}}
	if got := broken.At(project.LonLat{Lon: 0, Lat: 0}, breaks); !got.NoData {
		t.Errorf("a value that is not a number reads %+v", got)
	}
}

// TestImageClassAndNearestHeavier is plan task 11.7: the class here, and
// where the nearest heavier class is.
func TestImageClassAndNearestHeavier(t *testing.T) {
	// Three pixels across: light in the west, heavy in the east.
	img := Image{West: -30, South: -10, East: 30, North: 10, Width: 3, Height: 1,
		Classes: []int8{1, 2, 5}}
	class, ok := img.ClassAt(project.LonLat{Lon: -20, Lat: 0})
	if !ok || class != 1 {
		t.Errorf("the western pixel is class %d (%v)", class, ok)
	}
	heavier := img.NearestHeavier(project.LonLat{Lon: -20, Lat: 0})
	if !heavier.Found {
		t.Fatal("nothing heavier than the lightest class was found")
	}
	if heavier.Class != 2 {
		t.Errorf("the nearest heavier is class %d, want the middle pixel's 2", heavier.Class)
	}
	if math.Abs(heavier.Bearing-90) > 2 {
		t.Errorf("it is at bearing %v, want due east", heavier.Bearing)
	}
	// From the heaviest pixel, nothing is heavier, and that is said.
	if from := img.NearestHeavier(project.LonLat{Lon: 20, Lat: 0}); from.Found {
		t.Errorf("something heavier than the heaviest was found: %+v", from)
	}
	// No data here means anything at all counts as heavier.
	empty := Image{West: -30, South: -10, East: 30, North: 10, Width: 3, Height: 1, Classes: []int8{-1, 0, 3}}
	if _, ok := empty.ClassAt(project.LonLat{Lon: -20, Lat: 0}); ok {
		t.Error("a pixel of no data answered with a class")
	}
	if got := empty.NearestHeavier(project.LonLat{Lon: -20, Lat: 0}); !got.Found || got.Class != 0 {
		t.Errorf("from no data, the nearest heavier is %+v, want the class-0 pixel", got)
	}
	// Outside the image, there is no class at all.
	if _, ok := img.ClassAt(project.LonLat{Lon: 0, Lat: 80}); ok {
		t.Error("a place outside the image has a class")
	}
}

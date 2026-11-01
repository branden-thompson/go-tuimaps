package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// globe is a whole-world 2° field whose class changes every few cells both
// ways, so a column or a row placed one off reads another class.
func globe() scene.Field {
	f := scene.Field{West: -180, South: -90, East: 180, North: 90, Cols: 180, Rows: 90, Classes: make([]int8, 180*90), Preset: uint8(colour.MUF), ClassCount: 10, OverWater: true}
	for i := range f.Classes {
		f.Classes[i] = int8((i%180/3 + i/180/2) % 10)
	}
	return f
}

// TestAFieldsGridIsPlacedAsEachDotIsSampled (go-tuiMaps D-53): a field's
// class at every dot, read through the grid fieldAxes places once a frame,
// is the class fieldClass samples at the dot's centre - over a region, over
// the whole world, across the antimeridian, and for a grid whose classes do
// not fill it, which has none.
func TestAFieldsGridIsPlacedAsEachDotIsSampled(t *testing.T) {
	broken := bands(40, 30)
	broken.Classes = broken.Classes[:10]
	views := []project.View{gulfView(),
		{Centre: project.LonLat{Lon: 0, Lat: 20}, Zoom: 1, Cols: 200, Rows: 56},
		{Centre: project.LonLat{Lon: 180, Lat: -10}, Zoom: 1.5, Cols: 120, Rows: 40}}
	fields := []scene.Field{bands(40, 30), globe(), reach(), broken}
	classed := make([]int, len(fields)) // dots with a class, a field, over every view
	for _, v := range views {
		in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), OverlaysVersion: 1, Depth: colour.Truecolor}
		r, _ := drawn(t, in)
		if !r.axes(v) {
			t.Fatalf("view %+v: the dots could not be placed", v)
		}
		for i, f := range fields {
			cols, rows := r.fieldAxes(&f)
			for y := range r.lats {
				for x := range r.lons {
					want := fieldClass(&f, r.lons[x], r.lats[y])
					if got := classAt(&f, cols, rows, x, y); got != want {
						t.Fatalf("view %+v, field %v..%v: dot %d,%d reads class %d, sampled %d", v, f.West, f.East, x, y, got, want)
					}
					if want >= 0 {
						classed[i]++
					}
				}
			}
		}
	}
	for i, f := range fields {
		if whole := len(f.Classes) == f.Cols*f.Rows; whole != (classed[i] > 0) {
			t.Errorf("field %v..%v, its classes filling its grid %v: %d dots have a class", f.West, f.East, whole, classed[i])
		}
	}
}

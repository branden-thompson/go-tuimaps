package render

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// gulfView looks at the Gulf coast: land to the north, sea to the south.
func gulfView() project.View {
	return project.View{Centre: project.LonLat{Lon: -90, Lat: 29}, Zoom: 3.8, Cols: 120, Rows: 40}
}

// bands is a temperature field whose class rises from west to east.
func bands(cols, rows int) scene.Field {
	f := scene.Field{West: -110, South: 15, East: -70, North: 45, Cols: cols, Rows: rows, Classes: make([]int8, cols*rows), Preset: uint8(colour.Temperature), ClassCount: 17}
	for row := range rows {
		for col := range cols {
			f.Classes[row*cols+col] = int8(4 + col*10/cols)
		}
	}
	return f
}

func drawn(t *testing.T, in Input) (*Renderer, Frame) {
	t.Helper()
	r, err := NewRenderer(in.View.Cols, in.View.Rows)
	if err != nil {
		t.Fatal(err)
	}
	f, err := r.Draw(in)
	if err != nil {
		t.Fatal(err)
	}
	return r, f
}

// TestGridSampledAtDrawTime is part of plan task 09.23, and TestWaterMasksField
// of 09.7 (D-32): a field colours the cells of land, each sampled at its
// centre, and stops at the shore.
func TestGridSampledAtDrawTime(t *testing.T) { testField(t) }
func TestWaterMasksField(t *testing.T)       { testField(t) }

func testField(t *testing.T) {
	v := gulfView()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Fields: []scene.Field{bands(40, 30)}, OverlaysVersion: 1}
	r, _ := drawn(t, in)
	land, sea, classes := 0, 0, map[uint8]bool{}
	for i, c := range r.grid.cells {
		water := r.painter.Water(i%120, i/120)
		if water && c.under != 0 {
			t.Fatalf("cell %d is water and carries the field's colour; temperature stops at the shore", i)
		}
		if water {
			sea++
		}
		if c.under != 0 {
			land++
			classes[c.under] = true
		}
	}
	if land < 1000 || sea < 500 || len(classes) < 5 {
		t.Errorf("%d cells of field, %d of sea, %d classes; the view is half land, and the field has ten bands across it", land, sea, len(classes))
	}
	west, east := r.grid.cells[5*120+10].under, r.grid.cells[5*120+110].under
	if west == 0 || east == 0 || east <= west {
		t.Errorf("inks %d in the west and %d in the east; the classes rise eastwards", west, east)
	}
	if west < uint8(colour.Temperature1) || east > uint8(colour.Temperature17) {
		t.Errorf("inks %d and %d are not temperature tokens", west, east)
	}
	// A host can flip it: the field over water as well.
	in.FieldsOverWater = true
	in.OverlaysVersion = 2
	flipped, _ := drawn(t, in)
	over := 0
	for i, c := range flipped.grid.cells {
		if flipped.painter.Water(i%120, i/120) && c.under != 0 {
			over++
		}
	}
	if over < 500 {
		t.Errorf("%d cells of field over water with the mask turned off", over)
	}
}

// rain is an image with one heavy cell of rain over the sea, and a light
// shower beside it one pixel wide.
func rain() scene.Raster {
	ra := scene.Raster{West: -95, South: 22, East: -85, North: 28, Projection: 1, Width: 200, Height: 120, Classes: make([]int8, 200*120), Preset: uint8(colour.Radar), ClassCount: 7}
	for i := range ra.Classes {
		ra.Classes[i] = -1
	}
	for y := 40; y < 80; y++ {
		for x := 60; x < 140; x++ {
			ra.Classes[y*200+x] = 2
		}
		ra.Classes[y*200+100] = 6 // a line of the heaviest rain, one pixel wide
	}
	return ra
}

// TestWaterNeverMasksImage is the other half of 09.7 (D-87): rain over the
// sea is drawn. TestHeaviestInCell is D-78: of a cell's eight samples the
// heaviest class wins, so a storm's core one pixel wide is not averaged away.
func TestWaterNeverMasksImage(t *testing.T) { testImage(t) }
func TestHeaviestInCell(t *testing.T)       { testImage(t) }

func testImage(t *testing.T) {
	v := gulfView()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Rasters: []scene.Raster{rain()}, OverlaysVersion: 1}
	r, _ := drawn(t, in)
	overSea, heaviest, light := 0, 0, 0
	for i, c := range r.grid.cells {
		if c.under == 0 {
			continue
		}
		if r.painter.Water(i%120, i/120) {
			overSea++
		}
		switch c.under {
		case uint8(colour.Radar6):
			heaviest++
		case uint8(colour.Radar1 + 1):
			light++
		}
	}
	if overSea < 50 {
		t.Errorf("%d cells of rain over the sea; the image lies wholly over the Gulf", overSea)
	}
	if heaviest < 5 {
		t.Errorf("%d cells show the heaviest class; the core is one pixel wide and must not be lost", heaviest)
	}
	if light <= heaviest {
		t.Errorf("%d light and %d heavy cells", light, heaviest)
	}
}

// TestImageResampledAtDrawTime is the rest of 09.23 (PL-PF-4): a pan redraws
// the same prepared image in its new place, with no work done in between.
func TestImageResampledAtDrawTime(t *testing.T) {
	v := gulfView()
	ra := rain()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Rasters: []scene.Raster{ra}, OverlaysVersion: 1}
	r, _ := drawn(t, in)
	first := -1
	for i, c := range r.grid.cells {
		if c.under != 0 {
			first = i % 120
			break
		}
	}
	in.View.Centre.Lon -= 3 // the view moves west, so the rain moves east on the screen
	if _, err := r.Draw(in); err != nil {
		t.Fatal(err)
	}
	moved := -1
	for i, c := range r.grid.cells {
		if c.under != 0 {
			moved = i % 120
			break
		}
	}
	if first < 0 || moved <= first {
		t.Errorf("the rain's first column went from %d to %d", first, moved)
	}
}

// TestContoursAtBreaks and TestFlatDayNoContour are plan task 10.14, which
// lives here because contours are found at draw time (D-35, D-68): with no
// colour a field is lines where its class changes, each carrying its value.
func TestContoursAtBreaks(t *testing.T) {
	v := gulfView()
	// The values a contour carries ride on the prepared field itself, put
	// there when the grid was classified and its breaks were still known.
	field := bands(40, 30)
	field.Labels = []string{"", "", "", "", "-10", "-5", "0", "5", "10", "15", "20", "25", "30", "35", "40", "45", ""}
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Depth: colour.NoColour, Fields: []scene.Field{field}, OverlaysVersion: 1}
	r, f := drawn(t, in)
	contour := 0
	for _, c := range r.grid.cells {
		if c.ink == uint8(colour.LabelRegion) && c.text == "" && c.glyph != blank {
			contour++
		}
	}
	if contour < 60 {
		t.Errorf("%d cells of contour; ten bands cross forty rows", contour)
	}
	text := ""
	for _, l := range f.Lines {
		text += plain(l) + "\n"
	}
	if !strings.Contains(text, "10") || !strings.Contains(text, "20") {
		t.Errorf("the contours do not carry their values:\n%s", text)
	}
	if strings.Contains(strings.Join(f.Lines, ""), "\x1b") {
		t.Error("a frame with no colour holds a sequence")
	}
}

func TestFlatDayNoContour(t *testing.T) {
	v := gulfView()
	flat := bands(40, 30)
	for i := range flat.Classes {
		flat.Classes[i] = 9
	}
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Depth: colour.NoColour, Fields: []scene.Field{flat}, OverlaysVersion: 1}
	with, _ := drawn(t, in)
	in.Fields, in.OverlaysVersion = nil, 2
	without, _ := drawn(t, in)
	for i := range with.grid.cells {
		if with.grid.cells[i].glyph != without.grid.cells[i].glyph {
			t.Fatalf("cell %d differs; a field all of one class has no line to draw, and the description carries it (D-68)", i)
		}
	}
}

// TestImageBlockShadesNoColour is plan task 09.27 (FR-18): with no colour an
// image is block shades, heavier for heavier classes.
func TestImageBlockShadesNoColour(t *testing.T) {
	v := gulfView()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Depth: colour.NoColour, Rasters: []scene.Raster{rain()}, OverlaysVersion: 1}
	_, f := drawn(t, in)
	text := ""
	for _, l := range f.Lines {
		text += plain(l)
	}
	for _, shade := range []string{"\u2591", "\u2593"} {
		if !strings.Contains(text, shade) {
			t.Errorf("no %q in a frame with light rain and a heavy core", shade)
		}
	}
	for i, l := range f.Lines {
		if got := len([]rune(plain(l))); got != 120 {
			t.Fatalf("line %d is %d characters", i, got)
		}
	}
}

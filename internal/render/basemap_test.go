package render

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// square is a tile with one layer and one feature: a ring or a line through
// the given corners, in a 4096 grid.
func tileWith(layer string, f scene.Feature, coords ...int16) *scene.Tile {
	f.FirstPart, f.EndPart = 0, 1
	return &scene.Tile{Layers: []scene.Layer{{Name: layer, Extent: 4096, Features: []scene.Feature{f}, Coords: coords, Parts: []uint32{uint32(len(coords))}}}}
}

// here is the tile the painter tests draw: deep enough that the built-in
// style draws every role.
var here = scene.TileID{Z: 12, X: 2048, Y: 2048}

// worldView is a view that one tile fills exactly: 128 by 64 cells, 256 dots
// square, at the tile's own zoom and centred on it.
func worldView() project.View {
	return viewOf(here, 0)
}

// viewOf is the view centred on a tile, deeper than the tile by extra zooms.
func viewOf(tile scene.TileID, extra float64) project.View {
	centre, err := project.FromTile(float64(tile.X)+0.5, float64(tile.Y)+0.5, tile.Z)
	if err != nil {
		panic(err)
	}
	return project.View{Centre: centre, Zoom: float64(tile.Z) + extra, Cols: 128, Rows: 64}
}

func painter(t *testing.T, v project.View) *Painter {
	t.Helper()
	p, err := NewPainter(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestWaterwaysAndParksDrawn is plan task 09.26 (FR-2), and with it how any
// line feature reaches the canvas: through the style, in its role's ink.
func TestWaterwaysAndParksDrawn(t *testing.T) {
	v := worldView()
	p := painter(t, v)
	river := tileWith("waterway", scene.Feature{Kind: scene.GeomLine, Class: "river"}, 0, 2048, 4095, 2048)
	park := tileWith("park", scene.Feature{Kind: scene.GeomPolygon, Class: "national_park"}, 1024, 1024, 3072, 1024, 3072, 1536, 1024, 1536, 1024, 1024)
	for _, tile := range []*scene.Tile{river, park} {
		if err := p.Tile(v, tile, here, style.BuiltIn()); err != nil {
			t.Fatal(err)
		}
	}
	lines := p.Lines()
	if !lines.Lit(10, 128) || !lines.Lit(250, 128) {
		t.Errorf("the river is not drawn across the middle of the world")
	}
	if _, ink := lines.Cell(5, 32); ink != uint8(colour.River) {
		t.Errorf("the river's cell has ink %d, want the river token's", ink)
	}
	if !lines.Lit(64, 64) || !lines.Lit(192, 64) || !lines.Lit(128, 96) {
		t.Error("the park's outline is not drawn")
	}
	if lines.Lit(128, 80) {
		t.Error("the park is filled; a park is an outline")
	}
	if _, ink := lines.Cell(64, 16); ink != uint8(colour.Park) {
		t.Errorf("the park's cell has ink %d", ink)
	}
	// What no rule takes is not drawn.
	q := painter(t, v)
	if err := q.Tile(v, tileWith("poi", scene.Feature{Kind: scene.GeomLine, Class: "shop"}, 0, 0, 4095, 4095), here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(picture(q.Lines()), "#") {
		t.Error("a feature no rule takes was drawn")
	}
}

// TestWaterOwnsItsCells: water is a cell's background, not dots, and its edge
// is a line. The edge along a tile's own border is not a coast.
func TestWaterOwnsItsCells(t *testing.T) {
	v := worldView()
	p := painter(t, v)
	// The eastern half of the world tile is ocean, running out past the
	// tile's edge into its buffer, as real tiles do.
	ocean := tileWith("water", scene.Feature{Kind: scene.GeomPolygon, Class: "ocean"}, 2048, -64, 4160, -64, 4160, 4160, 2048, 4160, 2048, -64)
	if err := p.Tile(v, ocean, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if !p.Water(100, 30) || p.Water(20, 30) {
		t.Error("the cells of the eastern half are water and the western are not")
	}
	lines := p.Lines()
	if !lines.Lit(128, 100) {
		t.Error("the coast down the middle is not drawn")
	}
	if _, ink := lines.Cell(64, 25); ink != uint8(colour.Coast) {
		t.Errorf("the coast has ink %d", ink)
	}
	for _, dot := range [][2]int{{200, 0}, {255, 100}, {200, 255}} {
		if lines.Lit(dot[0], dot[1]) {
			t.Errorf("dot %v is lit: the polygon's edge along the tile's border is not a coast", dot)
		}
	}
}

// TestBorderSliverNotACoast (S26-4): the real zoom-0 tile has an ocean edge
// down the antimeridian that starts one unit inside the tile. It is where
// the world was cut, not a coast.
func TestBorderSliverNotACoast(t *testing.T) {
	v := worldView()
	p := painter(t, v)
	ocean := tileWith("water", scene.Feature{Kind: scene.GeomPolygon, Class: "ocean"}, 1, 4054, 0, 2239, 0, 1000, 2048, 1000, 2048, 4054, 1, 4054)
	if err := p.Tile(v, ocean, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	for _, y := range []int{80, 150, 200, 250} {
		if p.Lines().Lit(0, y) {
			t.Errorf("dot 0,%d is lit: an edge within a sliver of the tile's side was stroked as a coast", y)
		}
	}
	if !p.Lines().Lit(128, 150) {
		t.Error("the real coast down the middle is not drawn")
	}
	// A coast that merely passes near the side is still a coast.
	if onBorder(1, 100, 40, 900, 4096) || onBorder(4000, 5, 4094, 5, 4096) {
		t.Error("an edge that only ends near the side was taken for the border")
	}
	if !onBorder(1, 4054, 0, 2239, 4096) || !onBorder(4095, 10, 4096, 900, 4096) || !onBorder(5, 1, 900, 0, 4096) {
		t.Error("an edge lying within a sliver of the side was not taken for the border")
	}
}

// TestStandInDrawnLarger: an ancestor drawn in place of a missing tile is
// drawn at the missing tile's scale, and only its part of the view matters.
func TestStandInDrawnLarger(t *testing.T) {
	v := viewOf(here, 1) // the middle of the tile, one zoom deeper
	p := painter(t, v)
	river := tileWith("waterway", scene.Feature{Kind: scene.GeomLine, Class: "river"}, 0, 2048, 4095, 2048)
	if err := p.Tile(v, river, here, style.BuiltIn()); err != nil { // the shallower tile, standing in
		t.Fatal(err)
	}
	if !p.Lines().Lit(3, 128) || !p.Lines().Lit(250, 128) {
		t.Errorf("the river crosses the whole view when drawn from the shallower tile, one zoom deeper")
	}
}

// TestProjectionRounded is plan task 09.28 (NFR-6): a coordinate is rounded
// to 1/256 of a dot before it is rastered, so the same tile gives the same
// dots wherever it is computed.
func TestProjectionRounded(t *testing.T) {
	cases := map[float64]int{0: 0, 0.99: 0, 1: 1, 0.9999: 1, 127.998: 127, 127.9981: 128, -0.001: 0, -0.5: -1, -1.0001: -1}
	for value, want := range cases {
		if got := toDot(value); got != want {
			t.Errorf("%v rasters to dot %d, want %d", value, got, want)
		}
	}
	if got := place(100.25, 4095, 256.0/4096); got != toDot(100.25+float64(4095*(256.0/4096))) {
		t.Errorf("place: %d", got)
	}
}

func TestPainterRefusals(t *testing.T) {
	if _, err := NewPainter(0, 5); err == nil {
		t.Error("a painter with no size")
	}
	v := worldView()
	p := painter(t, v)
	if err := p.Tile(v, nil, here, style.BuiltIn()); err == nil {
		t.Error("no tile must be an error")
	}
	if err := p.Tile(v, &scene.Tile{}, scene.TileID{}, nil); err == nil {
		t.Error("no style must be an error")
	}
	if err := p.Tile(project.View{}, &scene.Tile{}, here, style.BuiltIn()); err == nil {
		t.Error("a view with no size must be an error")
	}
	broken := tileWith("waterway", scene.Feature{Kind: scene.GeomLine, Class: "river"}, 0, 0, 10, 10)
	broken.Layers[0].Parts[0] = 99
	if err := p.Tile(v, broken, here, style.BuiltIn()); err == nil {
		t.Error("a tile whose parts run past its coordinates must be an error")
	}
}

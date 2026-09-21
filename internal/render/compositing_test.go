package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// The nine layers of the compositing order (FR-12, L2 Render), numbered in
// the order they are drawn, bottom to top. Within the three parts of a cell
// - its background, its dots and its text - the later layer is what shows.
const (
	layerGround = iota + 1
	layerWater
	layerImage
	layerTint
	layerBasemapLine
	layerOverlayLine
	layerName
	layerMarker
	layerFurniture
)

// composed is one layer: how it is put into a frame, and how it is known in
// a cell once the frame is drawn.
type composed struct {
	name  string
	rank  int
	add   func(t *testing.T, in *Input, at spot)
	shows func(c cell) bool
}

// spot is where a pair of layers is made to meet: a place, its cell, and the
// row of dots through it.
type spot struct {
	at       project.LonLat
	col, row int
	dotRow   int
}

// where is the spot of one cell of the compositing frame.
func where(t *testing.T, col, row int) spot {
	t.Helper()
	v := worldView()
	at, err := v.FromDot(float64(col*2)+0.5, float64(row*4)+1.5)
	if err != nil {
		t.Fatal(err)
	}
	return spot{at: at, col: col, row: row, dotRow: row*4 + 1}
}

// across is a line straight through a spot's row of dots: in the tile's own
// coordinates, and as an overlay's two vertices.
func across(t *testing.T, s spot) ([]int16, []scene.Vertex) {
	t.Helper()
	v := worldView()
	west, err := v.FromDot(0, float64(s.dotRow))
	if err != nil {
		t.Fatal(err)
	}
	east, err := v.FromDot(float64(v.Cols*2), float64(s.dotRow))
	if err != nil {
		t.Fatal(err)
	}
	y := int16(float64(s.dotRow) / float64(v.Rows*4) * 4096) // one tile fills the view exactly
	return []int16{0, y, 4095, y}, []scene.Vertex{vertex(t, west.Lon, west.Lat), vertex(t, east.Lon, east.Lat)}
}

// wholeTile is a ring round the whole tile, in its own coordinates.
func wholeTile() []int16 {
	return []int16{0, 0, 4095, 0, 4095, 4095, 0, 4095, 0, 0}
}

// wholeWorld is a ring round the whole view, as an overlay's.
func wholeWorld(t *testing.T) [][]scene.Vertex {
	t.Helper()
	v := worldView()
	west, err := v.FromDot(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	east, err := v.FromDot(float64(v.Cols*2-1), float64(v.Rows*4-1))
	if err != nil {
		t.Fatal(err)
	}
	return [][]scene.Vertex{lonLatBox(t, west.Lon, east.Lat, east.Lon, west.Lat)}
}

// alertTint is the token an alert's area is filled with.
const alertTint = uint8(colour.AlertSevereTint)

// wholeRaster is an image of one class over the whole world.
func wholeRaster() scene.Raster {
	return scene.Raster{West: -179, South: -84, East: 179, North: 84, Projection: 2, Width: 2, Height: 2,
		Classes: []int8{3, 3, 3, 3}, Preset: uint8(colour.Radar), ClassCount: 6}
}

// layers are the nine, each made to land on the spot it is given.
func layers() []composed {
	return []composed{
		{"the ground", layerGround, func(t *testing.T, in *Input, s spot) {},
			func(c cell) bool { return c.area == 0 && c.under == 0 }},
		{"water", layerWater, func(t *testing.T, in *Input, s spot) {
			in.Tiles = append(in.Tiles, Drawn{Tile: tileWith("water", scene.Feature{Kind: scene.GeomPolygon, Class: "ocean"}, wholeTile()...), At: here, Exact: true})
		}, func(c cell) bool { return c.area == uint8(colour.WaterFill) && c.under == 0 }},
		{"an image", layerImage, func(t *testing.T, in *Input, s spot) {
			in.Rasters = append(in.Rasters, wholeRaster())
			in.OverlaysVersion++
		}, func(c cell) bool { return c.under != 0 && c.area != alertTint }},
		{"an alert's tint", layerTint, func(t *testing.T, in *Input, s spot) {
			in.Shapes = append(in.Shapes, scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Rings: wholeWorld(t)})
			in.OverlaysVersion++
		}, func(c cell) bool { return c.area == alertTint }},
		{"a basemap line", layerBasemapLine, func(t *testing.T, in *Input, s spot) {
			line, _ := across(t, s)
			in.Tiles = append(in.Tiles, Drawn{Tile: tileWith("waterway", scene.Feature{Kind: scene.GeomLine, Class: "river"}, line...), At: here, Exact: true})
		}, func(c cell) bool { return c.glyph != blank && c.ink == uint8(colour.River) }},
		{"an overlay's line", layerOverlayLine, func(t *testing.T, in *Input, s spot) {
			_, ends := across(t, s)
			in.Shapes = append(in.Shapes, scene.Shape{Kind: scene.ShapeLine, Role: uint8(colour.Track), Rings: [][]scene.Vertex{ends}})
			in.OverlaysVersion++
		}, func(c cell) bool { return c.glyph != blank && c.ink == uint8(colour.Track) }},
		{"a name", layerName, func(t *testing.T, in *Input, s spot) {
			v := worldView()
			x := int16(float64(s.col*2) / float64(v.Cols*2) * 4096)
			y := int16(float64(s.dotRow) / float64(v.Rows*4) * 4096)
			in.Tiles = append(in.Tiles, Drawn{Tile: tileWith("place", scene.Feature{Kind: scene.GeomPoint, Class: "city", Name: textsafe.Clean("N"), Rank: 1}, x, y), At: here, Exact: true})
			in.Labels = true
		}, func(c cell) bool { return c.text == "N" }},
		{"a marker", layerMarker, func(t *testing.T, in *Input, s spot) {
			in.Markers = append(in.Markers, Marker{At: s.at, Shape: MarkerGlyph, Text: "◉", Ink: uint8(colour.Marker)})
		}, func(c cell) bool { return c.text == "◉" }},
		{"the furniture", layerFurniture, func(t *testing.T, in *Input, s spot) {
			in.Credit = textsafe.Const("D")
		}, func(c cell) bool { return c.ink == uint8(colour.Credit) && c.text != "" }},
	}
}

// TestCompositingOrder is plan task 09.6 (FR-12): the nine layers, each pair
// of them, the later one seen where both want the same cell.
func TestCompositingOrder(t *testing.T) {
	all := layers()
	v := worldView()
	middle := where(t, v.Cols/2, v.Rows/2)
	credit := where(t, v.Cols-1, v.Rows-1) // the right end of the last row, where the credit sits
	for i, under := range all {
		for _, over := range all[i+1:] {
			s := middle
			if over.rank == layerFurniture {
				s = credit
			}
			// A tile with nothing on it: the frame has one, so it is not the
			// empty rectangle, whose notice would take the middle cell.
			in := Input{View: v, Style: style.BuiltIn(), Depth: Truecolor,
				Tiles: []Drawn{{Tile: &scene.Tile{}, At: here, Exact: true}}}
			under.add(t, &in, s)
			over.add(t, &in, s)
			r, err := NewRenderer(v.Cols, v.Rows)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = r.Draw(in); err != nil {
				t.Fatal(err)
			}
			c := r.grid.cells[s.row*v.Cols+s.col]
			if !over.shows(c) {
				t.Errorf("%s under %s: the cell does not show the one on top (%+v)", under.name, over.name, c)
			}
			if sameChannel(under.rank, over.rank) && under.shows(c) {
				t.Errorf("%s under %s: the cell still shows the one underneath (%+v)", under.name, over.name, c)
			}
		}
	}
}

// sameChannel reports whether two layers want the same part of a cell. A
// cell has one background, one set of dots and one piece of text; within
// each the later layer wins outright, and across them all three are seen.
func sameChannel(a, b int) bool {
	return channelOf(a) == channelOf(b)
}

func channelOf(rank int) int {
	switch {
	case rank <= layerTint:
		return 1 // the background
	case rank <= layerOverlayLine:
		return 2 // the dots
	}
	return 3 // the text
}

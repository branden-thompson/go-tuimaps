package project

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestTilePlace: where a tile sits in a view's dots, and how big it is drawn,
// for the tile's own zoom and for a stand-in from another.
func TestTilePlace(t *testing.T) {
	world := View{Centre: LonLat{}, Zoom: 1, Cols: 256, Rows: 128} // 512 by 512 dots: the whole world at zoom 1
	cases := []struct {
		tile       scene.TileID
		x, y, side float64
	}{
		{scene.TileID{Z: 1, X: 0, Y: 0}, 0, 0, 256},
		{scene.TileID{Z: 1, X: 1, Y: 1}, 256, 256, 256},
		{scene.TileID{Z: 0, X: 0, Y: 0}, 0, 0, 512},    // an ancestor standing in: drawn twice the size
		{scene.TileID{Z: 3, X: 4, Y: 4}, 256, 256, 64}, // a deeper tile: drawn smaller
	}
	for _, c := range cases {
		x, y, side, err := world.TilePlace(c.tile)
		if err != nil || x != c.x || y != c.y || side != c.side {
			t.Errorf("%v: %v, %v, side %v, %v; want %v, %v, side %v", c.tile, x, y, side, err, c.x, c.y, c.side)
		}
	}
	half := View{Centre: LonLat{}, Zoom: 1.5, Cols: 10, Rows: 5}
	_, _, side, err := half.TilePlace(scene.TileID{Z: 1, X: 0, Y: 0})
	if err != nil || side < 362 || side > 363 {
		t.Errorf("at zoom 1.5 a zoom-1 tile is %v dots, want 256 times the square root of 2; %v", side, err)
	}
	if _, _, _, err := world.TilePlace(scene.TileID{Z: 40}); err == nil {
		t.Error("a tile deeper than any zoom has a place")
	}
	if _, _, _, err := (View{}).TilePlace(scene.TileID{}); err == nil {
		t.Error("a view with no size places a tile")
	}
}

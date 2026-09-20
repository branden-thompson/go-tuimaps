package project

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// DotsPerCol is the braille cell's width: two dots across.
	DotsPerCol = 2
	// DotsPerRow is the braille cell's height: four dots down. A cell is
	// about twice as tall as it is wide, so a row covers twice the ground of
	// a column.
	DotsPerRow = 4
	// MaxCells bounds a view's width and height, in cells.
	MaxCells = 10000
	// MinViewZoom is the furthest a view zooms out: a whole world a few dots
	// wide. The closest is MaxViewZoom.
	MinViewZoom = -8
	// maxTilesForView bounds the tiles one view may ask for.
	maxTilesForView = 4096
)

// View is what the map shows: a centre, a zoom - fractional zooms are
// allowed - and a rectangle of cells.
type View struct {
	Centre LonLat
	Zoom   float64
	Cols   int
	Rows   int
}

// badView is the error for a view that cannot be drawn.
func badView() error {
	return fault.Make(fault.InvalidCoordinates,
		textsafe.Const("the view cannot be drawn"),
		textsafe.Const("its size is not between 1 and 10000 cells each way, its zoom is not between -8 and 18, or its centre is not a position on the map"),
		textsafe.Const("check the size, the zoom and the centre"))
}

// Validate reports whether the view can be drawn.
func (v View) Validate() error {
	if v.Cols < 1 || v.Cols > MaxCells {
		return badView()
	}
	if v.Rows < 1 || v.Rows > MaxCells {
		return badView()
	}
	if !(v.Zoom >= MinViewZoom && v.Zoom <= MaxViewZoom) { // NaN fails this too
		return badView()
	}
	if !(v.Centre.Lon >= -math.MaxFloat64 && v.Centre.Lon <= math.MaxFloat64) {
		return badView()
	}
	if !(v.Centre.Lat >= -90 && v.Centre.Lat <= 90) {
		return badView()
	}
	return nil
}

// origin returns the world-dot coordinates of the view's top-left corner.
func (v View) origin() (x, y float64, err error) {
	if err := v.Validate(); err != nil {
		return 0, 0, err
	}
	cx, cy, err := ToTile(v.Centre, 0)
	if err != nil {
		return 0, 0, err
	}
	w := TileSize * math.Exp2(v.Zoom) // the side of the whole world, in dots, at this zoom
	return cx*w - float64(v.Cols*DotsPerCol)/2, cy*w - float64(v.Rows*DotsPerRow)/2, nil
}

// ToDot projects a position into the view's dots: x grows east and y grows
// south from the view's top-left corner. A position outside the view gets
// coordinates outside the rectangle.
func (v View) ToDot(p LonLat) (x, y float64, err error) {
	ox, oy, err := v.origin()
	if err != nil {
		return 0, 0, err
	}
	px, py, err := ToTile(p, 0)
	if err != nil {
		return 0, 0, err
	}
	w := TileSize * math.Exp2(v.Zoom) // the side of the whole world, in dots, at this zoom
	return px*w - ox, py*w - oy, nil
}

// FromDot is the inverse of ToDot.
func (v View) FromDot(x, y float64) (LonLat, error) {
	ox, oy, err := v.origin()
	if err != nil {
		return LonLat{}, err
	}
	w := TileSize * math.Exp2(v.Zoom) // the side of the whole world, in dots, at this zoom
	return FromTile((x+ox)/w, (y+oy)/w, 0)
}

// Tiles returns the tiles the view is drawn from, ordered by zoom, then
// column, then row, so that two renders of one view read them in one order
// (NFR-6). The map does not repeat across the antimeridian, as upstream: a
// view that reaches past the edge of the world asks for nothing there.
func (v View) Tiles() ([]scene.TileID, error) {
	ox, oy, err := v.origin()
	if err != nil {
		return nil, err
	}
	zoom, err := TileZoom(v.Zoom)
	if err != nil {
		return nil, err
	}
	size, err := TileSizeAt(v.Zoom)
	if err != nil {
		return nil, err
	}
	last := float64(uint32(1)<<zoom - 1)
	clamp := func(dot float64) uint32 { return uint32(math.Max(0, math.Min(last, math.Floor(dot/size)))) }
	x0, x1 := clamp(ox), clamp(ox+float64(v.Cols*DotsPerCol)-1)
	y0, y1 := clamp(oy), clamp(oy+float64(v.Rows*DotsPerRow)-1)
	if uint64(x1-x0+1)*uint64(y1-y0+1) > maxTilesForView {
		return nil, badView()
	}
	tiles := make([]scene.TileID, 0, (x1-x0+1)*(y1-y0+1))
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			tiles = append(tiles, scene.TileID{Z: zoom, X: x, Y: y})
		}
	}
	return tiles, nil
}

// TilePlace is where a tile sits in the view's dots - its top-left corner -
// and the side it is drawn at: 256 dots at the tile's own zoom, larger for an
// ancestor standing in, smaller for a deeper tile.
func (v View) TilePlace(t scene.TileID) (x, y, side float64, err error) {
	err = t.Validate()
	if err != nil {
		return 0, 0, 0, err
	}
	ox, oy, err := v.origin()
	if err != nil {
		return 0, 0, 0, err
	}
	side = TileSize * math.Exp2(v.Zoom-float64(t.Z))
	return float64(t.X)*side - ox, float64(t.Y)*side - oy, side, nil
}

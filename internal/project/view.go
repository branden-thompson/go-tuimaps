package project

import (
	"cmp"
	"math"
	"slices"

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
	// Shift draws the world this many times its width east (negative: west)
	// of where it is: the renderer draws each copy of the world a view reaches
	// through a view of that copy, so a view past the antimeridian is filled
	// (v0.2.0 D-86). Zero for the view the host set.
	Shift int
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
	if v.Shift != 0 {
		// A COPY OF THE WORLD (D-86): its origin a whole number of worlds
		// west, so everything drawn through this view lands that far east.
		cx -= float64(v.Shift)
	}
	// **Each product is converted before the half-window is taken off it**,
	// so that no machine fuses the two into one rounding. This is the
	// origin every dot of a frame is measured from: a fused multiply-add
	// keeps more bits than the separate steps, so a machine that fuses
	// puts a line's dots in different cells from one that does not, and
	// the same data draws two pictures (NFR-6, constants section 6).
	return float64(cx*w) - float64(v.Cols*DotsPerCol)/2, float64(cy*w) - float64(v.Rows*DotsPerRow)/2, nil
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
	p, err := FromTile((x+ox)/w, (y+oy)/w, 0)
	if err != nil {
		return LonLat{}, err
	}
	return Normalize(p) // past the antimeridian, the other side's longitude (D-86)
}

// Placement is a tile the view draws and where: the tile itself, and how
// many worlds east (or, negative, west) of its own place it is drawn - the
// world repeats across the antimeridian (L-14.6, v0.2.0 D-86; from watchpost D-90).
type Placement struct {
	ID    scene.TileID
	Shift int
}

// Tiles returns the tiles the view is drawn from, each once, ordered by
// zoom, then column, then row, so that two renders of one view read them in
// one order (NFR-6). A view past the antimeridian wants the tiles of the
// other side: the world repeats (D-86).
func (v View) Tiles() ([]scene.TileID, error) {
	placed, err := v.Placements()
	if err != nil {
		return nil, err
	}
	tiles := make([]scene.TileID, 0, len(placed))
	for _, p := range placed {
		tiles = append(tiles, p.ID)
	}
	slices.SortFunc(tiles, func(a, b scene.TileID) int { // no reflection, no allocation: a frame advance is counted
		if a.X != b.X {
			return cmp.Compare(a.X, b.X)
		}
		return cmp.Compare(a.Y, b.Y)
	})
	return slices.Compact(tiles), nil // a tile two copies of the world reach, once
}

// Placements are every tile the view draws, with its shift in worlds, in
// column then row order. THE WORLD REPEATS ACROSS THE ANTIMERIDIAN (D-86): a
// view reaching past the edge of the world draws the other side there, so a
// window is always filled. Upstream drew nothing past the edge; the rows
// still stop at the poles.
func (v View) Placements() ([]Placement, error) {
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
	n := int64(1) << zoom
	last := float64(n - 1)
	clamp := func(dot float64) uint32 { return uint32(math.Max(0, math.Min(last, math.Floor(dot/size)))) }
	x0, x1 := int64(math.Floor(ox/size)), int64(math.Floor((ox+float64(v.Cols*DotsPerCol)-1)/size))
	y0, y1 := clamp(oy), clamp(oy+float64(v.Rows*DotsPerRow)-1)
	if x1 < x0 || uint64(x1-x0+1)*uint64(y1-y0+1) > maxTilesForView {
		return nil, badView()
	}
	placed := make([]Placement, 0, (x1-x0+1)*int64(y1-y0+1))
	for x := x0; x <= x1; x++ {
		shift := floorDiv(x, n)
		for y := y0; y <= y1; y++ {
			placed = append(placed, Placement{ID: scene.TileID{Z: zoom, X: uint32(x - shift*n), Y: y}, Shift: int(shift)})
		}
	}
	return placed, nil
}

// floorDiv is a divided by b, rounded down: -1 for the first column west of
// the antimeridian.
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
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
	// **The product is converted before the origin is taken off it**, so
	// that no machine fuses the two into one rounding: a fused
	// multiply-add keeps more bits than the separate steps, and this is
	// where a tile's corner is decided. A corner a fraction of a dot apart
	// puts a line's dots in different cells, and the same data then draws
	// two pictures on two machines (NFR-6, constants section 6).
	return float64(float64(t.X)*side) - ox, float64(float64(t.Y)*side) - oy, side, nil
}

// Shifts are the copies of the world the view reaches, from the westmost to
// the eastmost: 0 and 0 for a view that stays within one (D-86).
func (v View) Shifts() (lo, hi int, err error) {
	ox, _, err := v.origin()
	if err != nil {
		return 0, 0, err
	}
	w := TileSize * math.Exp2(v.Zoom)
	return int(math.Floor(ox / w)), int(math.Floor((ox + float64(v.Cols*DotsPerCol) - 1) / w)), nil
}

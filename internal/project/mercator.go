// Package project holds the arithmetic of the map: web mercator, the view
// and its dots, the tiles a view needs, distances and bearings on the
// sphere, and fit-to. It draws nothing and fetches nothing.
//
// The forward projection is written as atanh(sin(lat)) and the inverse as
// asin(tanh(y)): the same web mercator every tile source uses, in functions
// that have no assembly version on the architectures whose frames are
// compared byte for byte (constants, section 6).
package project

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// TileSize is the side of a tile in dots at its own zoom.
	TileSize = 256
	// MaxSourceZoom is the deepest zoom a tile source has; a view zoomed
	// closer is drawn from zoom-14 tiles, stretched.
	MaxSourceZoom = 14
	// MaxLatitude is where web mercator's square world ends.
	MaxLatitude = 85.0511
)

// LonLat is a position in degrees: longitude east, latitude north.
type LonLat struct {
	Lon float64
	Lat float64
}

// notFinite is the one error this file raises: a number that is not a
// number, is infinite or lies outside the world must never reach the
// arithmetic, still less a conversion to an integer.
func notFinite() error {
	return fault.Make(fault.InvalidCoordinates,
		textsafe.Const("a coordinate or a zoom is not a finite number"),
		textsafe.Const("it is NaN or infinite, a latitude beyond the poles, or a zoom deeper than the library addresses"),
		textsafe.Const("check the values handed in"))
}

// Normalize brings a position into the map's range as upstream does:
// longitude wraps by a single turn, latitude is clamped to where the
// projection's square world ends.
func Normalize(p LonLat) (LonLat, error) {
	if !(p.Lon >= -math.MaxFloat64 && p.Lon <= math.MaxFloat64) { // NaN and both infinities fail this
		return LonLat{}, notFinite()
	}
	if !(p.Lat >= -math.MaxFloat64 && p.Lat <= math.MaxFloat64) { // NaN and both infinities fail this
		return LonLat{}, notFinite()
	}
	if p.Lon < -180 {
		p.Lon += 360
	}
	if p.Lon > 180 {
		p.Lon -= 360
	}
	p.Lat = math.Max(-MaxLatitude, math.Min(MaxLatitude, p.Lat))
	return p, nil
}

// ToTile projects a position into tile coordinates at a whole zoom: x grows
// east and y grows south, each from 0 to 2^zoom across the world. Latitude
// is clamped to the projection's range; longitude is not wrapped.
func ToTile(p LonLat, zoom uint8) (x, y float64, err error) {
	if zoom > scene.MaxTileZoom {
		return 0, 0, notFinite()
	}
	if !(p.Lon >= -math.MaxFloat64 && p.Lon <= math.MaxFloat64) { // NaN and both infinities fail this
		return 0, 0, notFinite()
	}
	if !(p.Lat >= -math.MaxFloat64 && p.Lat <= math.MaxFloat64) { // NaN and both infinities fail this
		return 0, 0, notFinite()
	}
	if p.Lat < -90 || p.Lat > 90 {
		return 0, 0, notFinite()
	}
	n := math.Exp2(float64(zoom))
	lat := math.Max(-MaxLatitude, math.Min(MaxLatitude, p.Lat))
	merc := math.Atanh(math.Sin(lat * math.Pi / 180))
	return (p.Lon + 180) / 360 * n, (1 - merc/math.Pi) / 2 * n, nil
}

// FromTile is the inverse of ToTile.
func FromTile(x, y float64, zoom uint8) (LonLat, error) {
	if zoom > scene.MaxTileZoom {
		return LonLat{}, notFinite()
	}
	if !(x >= -math.MaxFloat64 && x <= math.MaxFloat64) { // NaN and both infinities fail this
		return LonLat{}, notFinite()
	}
	if !(y >= -math.MaxFloat64 && y <= math.MaxFloat64) { // NaN and both infinities fail this
		return LonLat{}, notFinite()
	}
	n := math.Exp2(float64(zoom))
	merc := math.Pi * (1 - 2*y/n)
	return LonLat{Lon: x/n*360 - 180, Lat: math.Asin(math.Tanh(merc)) * 180 / math.Pi}, nil
}

// TileZoom is the zoom of the tiles a view at the given zoom is drawn from:
// its whole part, held between 0 and the deepest zoom a source has.
func TileZoom(zoom float64) (uint8, error) {
	if !(zoom >= -math.MaxFloat64 && zoom <= math.MaxFloat64) { // NaN and both infinities fail this
		return 0, notFinite()
	}
	return uint8(math.Max(0, math.Min(MaxSourceZoom, math.Floor(zoom)))), nil
}

// TileSizeAt is the side, in dots, of one tile as a view at the given zoom
// draws it: 256 at the tile's own zoom, growing with the fraction, and
// stretched beyond that above the deepest source zoom.
func TileSizeAt(zoom float64) (float64, error) {
	if !(zoom >= -math.MaxFloat64 && zoom <= math.MaxFloat64) { // NaN and both infinities fail this
		return 0, notFinite()
	}
	base, err := TileZoom(zoom)
	if err != nil {
		return 0, err
	}
	return TileSize * math.Exp2(zoom-float64(base)), nil
}

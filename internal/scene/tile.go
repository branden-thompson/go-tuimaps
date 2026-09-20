package scene

import (
	"errors"

	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// What the kept form's own structures cost, in bytes, on a 64-bit machine.
// A test holds each to the real size.
const (
	tileBytes    = 24
	layerBytes   = 96
	featureBytes = 48
)

// GeomKind is what a feature's geometry is.
type GeomKind uint8

// The kinds of geometry a vector tile carries.
const (
	GeomPoint GeomKind = iota + 1
	GeomLine
	GeomPolygon
)

// Tile is a decoded vector tile in its kept form: only the layers the map
// draws, only the attributes it reads, and every coordinate as a 16-bit
// integer (D-75).
type Tile struct {
	Layers []Layer
}

// Layer is one kept layer of a tile. Its geometry lives in two slabs of its
// own, each allocated once at its exact size.
type Layer struct {
	Name   string
	Extent uint16 // the side of the layer's coordinate grid; 4096 unless the tile says otherwise
	// Untyped counts the features passed over for want of a declared type.
	// A feature with no type cannot be drawn - there is no telling whether
	// to fill it or stroke it - so it is dropped; this says how many were,
	// so that a reader can tell "dropped because it said nothing" from
	// "lost" (D-118). It sits here, beside the extent, because the bytes
	// either side of it are padding and the layer's size is accounted for.
	Untyped  uint16
	Features []Feature
	// Coords holds every feature's coordinates as x, y pairs.
	Coords []int16
	// Parts holds, for each part - a point, a line, a ring - the index in
	// Coords one past its last value. Part i runs from Parts[i-1], or from
	// 0 for the first, up to Parts[i].
	Parts []uint32
}

// Feature is one kept feature. A polygon feature is one outline and its
// holes: a tile feature holding several polygons becomes several of these,
// each carrying the same attributes.
type Feature struct {
	Kind       GeomKind
	AdminLevel uint8         // a boundary's administrative level: 2 a country's, 4 a region's; 0 if it has none
	Maritime   bool          // a boundary drawn across the sea
	Rank       int32         // importance, lower first: the tile's local rank, else its scale rank, else its rank, else 0
	Class      string        // shared between the features of a tile that have the same class
	Name       textsafe.Text // cleaned where it is read, once per tile, so that no frame cleans it again (D-120)
	FirstPart  uint32        // the feature's parts are its layer's Parts[FirstPart:EndPart]
	EndPart    uint32
}

// Bytes is what the tile's kept form costs in memory, for the caches' byte
// caps and the decoder's retained-size limit. Class strings are shared, so
// they are counted once for each run of features that share one. Slabs are
// counted at their capacity: what is held, not what is used.
func (t *Tile) Bytes() int {
	if t == nil {
		return 0
	}
	n := tileBytes
	for _, l := range t.Layers {
		n += layerBytes + len(l.Name) + cap(l.Features)*featureBytes + cap(l.Coords)*2 + cap(l.Parts)*4
		last := ""
		for _, f := range l.Features {
			n += len(f.Name.String())
			if f.Class != last {
				n += len(f.Class)
				last = f.Class
			}
		}
	}
	return n
}

// ErrBadPart is returned by Part for an index outside the layer, or for
// slabs that do not agree with each other.
var ErrBadPart = errors.New("scene: no such part, or the layer's slabs disagree")

// Part returns the x, y pairs of one part of the layer: a point, a line or
// a ring. It is the one place the slab arithmetic lives.
func (l *Layer) Part(i int) ([]int16, error) {
	if l == nil {
		return nil, ErrBadPart
	}
	if i < 0 || i >= len(l.Parts) {
		return nil, ErrBadPart
	}
	start, end := 0, int(l.Parts[i])
	if i > 0 {
		start = int(l.Parts[i-1])
	}
	if start > end || end > len(l.Coords) {
		return nil, ErrBadPart
	}
	if (end-start)%2 != 0 {
		return nil, ErrBadPart
	}
	return l.Coords[start:end], nil
}

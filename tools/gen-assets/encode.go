// Command gen-assets builds the tiles embedded in the library's assets
// package: zoom 0 to 3, read from the pinned planet archive, decoded with
// the library's own decoder, stripped to what the library reads, and written
// back as small vector tiles (FR-28a). The encoder lives here and nowhere
// else: the library never encodes.
package main

import (
	"errors"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// drawn is what the generator keeps: the layers the map draws, and English,
// the one language the embedded tiles carry (D-82).
func drawn() mvt.Want {
	return mvt.Want{
		Layers:   []string{"water", "waterway", "landcover", "park", "boundary", "transportation", "aeroway", "place", "water_name", "aerodrome_label"},
		Language: "en",
	}
}

// The geometry commands, and the wire types the encoder writes.
const (
	cmdMoveTo    = 1
	cmdLineTo    = 2
	cmdClosePath = 7
	wireVarint   = 0
	wireBytes    = 2
)

func putVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func putBytes(b []byte, field uint32, body []byte) []byte {
	b = putVarint(b, uint64(field)<<3|wireBytes)
	b = putVarint(b, uint64(len(body)))
	return append(b, body...)
}

func putUint(b []byte, field uint32, v uint64) []byte {
	b = putVarint(b, uint64(field)<<3|wireVarint)
	return putVarint(b, v)
}

func zigzag(v int32) uint64 { return uint64(uint32(v<<1) ^ uint32(v>>31)) }

// values is a layer's table of attribute values, each written once.
type values struct {
	index   map[string]uint64
	encoded [][]byte
}

func (v *values) text(s string) uint64 {
	key := "s" + s
	if i, ok := v.index[key]; ok {
		return i
	}
	v.index[key] = uint64(len(v.encoded))
	v.encoded = append(v.encoded, putBytes(nil, 1, []byte(s)))
	return v.index[key]
}

func (v *values) rank(n int32) uint64 {
	key := "i" + strconv.Itoa(int(n))
	if i, ok := v.index[key]; ok {
		return i
	}
	v.index[key] = uint64(len(v.encoded))
	v.encoded = append(v.encoded, putUint(nil, 6, zigzag(n))) // sint64, so a negative rank survives
	return v.index[key]
}

// Encode writes a tile's kept form as a vector tile. What the library's
// decoder keeps of the result is exactly what it kept of the source.
func Encode(tile *scene.Tile) ([]byte, error) {
	if tile == nil {
		return nil, errors.New("gen-assets: no tile to encode")
	}
	if len(tile.Layers) == 0 {
		return nil, errors.New("gen-assets: a tile with no layer the map draws; every tile of zoom 0 to 3 has land or water")
	}
	var out []byte
	for i := range tile.Layers {
		layer, err := encodeLayer(&tile.Layers[i])
		if err != nil {
			return nil, err
		}
		out = putBytes(out, 3, layer)
	}
	return out, nil
}

// The keys a stripped layer carries, by index.
const (
	keyClass = iota
	keyName
	keyRank
	keyLevel
	keyMaritime
)

func encodeLayer(l *scene.Layer) ([]byte, error) {
	if l == nil || l.Name == "" {
		return nil, errors.New("gen-assets: a layer with no name")
	}
	if l.Extent == 0 {
		return nil, errors.New("gen-assets: a layer with no extent")
	}
	vals := &values{index: map[string]uint64{}}
	b := putUint(nil, 15, 2)
	b = putBytes(b, 1, []byte(l.Name))
	for _, f := range l.Features {
		feature, err := encodeFeature(l, f, vals)
		if err != nil {
			return nil, err
		}
		b = putBytes(b, 2, feature)
	}
	for _, k := range []string{"class", "name", "localrank", "admin_level", "maritime"} {
		b = putBytes(b, 3, []byte(k))
	}
	for _, v := range vals.encoded {
		b = putBytes(b, 4, v)
	}
	return putUint(b, 5, uint64(l.Extent)), nil
}

func encodeFeature(l *scene.Layer, f scene.Feature, vals *values) ([]byte, error) {
	if l == nil || vals == nil {
		return nil, errors.New("gen-assets: a feature with no layer")
	}
	if f.Kind < scene.GeomPoint || f.Kind > scene.GeomPolygon {
		return nil, errors.New("gen-assets: a feature that is not a point, a line or a polygon")
	}
	if f.FirstPart >= f.EndPart {
		return nil, errors.New("gen-assets: a feature with no geometry")
	}
	var tags []byte
	if f.Class != "" {
		tags = putVarint(putVarint(tags, keyClass), vals.text(f.Class))
	}
	if f.Name != "" {
		tags = putVarint(putVarint(tags, keyName), vals.text(f.Name))
	}
	if f.Rank != 0 {
		tags = putVarint(putVarint(tags, keyRank), vals.rank(f.Rank))
	}
	if f.AdminLevel != 0 {
		tags = putVarint(putVarint(tags, keyLevel), vals.rank(int32(f.AdminLevel)))
	}
	if f.Maritime {
		tags = putVarint(putVarint(tags, keyMaritime), vals.rank(1))
	}
	geometry, err := encodeGeometry(l, f)
	if err != nil {
		return nil, err
	}
	var b []byte
	if len(tags) > 0 {
		b = putBytes(b, 2, tags)
	}
	b = putUint(b, 3, uint64(f.Kind))
	return putBytes(b, 4, geometry), nil
}

// encodeGeometry writes a feature's parts as commands. Points share one
// MoveTo; a line is a MoveTo and a LineTo; a ring that ends where it began
// is written with ClosePath, which pushes that point again on decoding.
func encodeGeometry(l *scene.Layer, f scene.Feature) ([]byte, error) {
	if l == nil {
		return nil, errors.New("gen-assets: a geometry with no layer")
	}
	if f.FirstPart >= f.EndPart {
		return nil, errors.New("gen-assets: a feature with no geometry")
	}
	var b []byte
	var cx, cy int32
	step := func(x, y int16) {
		b = putVarint(putVarint(b, zigzag(int32(x)-cx)), zigzag(int32(y)-cy))
		cx, cy = int32(x), int32(y)
	}
	if f.Kind == scene.GeomPoint {
		b = putVarint(b, cmdMoveTo|uint64(f.EndPart-f.FirstPart)<<3)
	}
	for p := f.FirstPart; p < f.EndPart; p++ {
		pts, err := l.Part(int(p))
		if err != nil {
			return nil, err
		}
		if len(pts) < 2 {
			return nil, errors.New("gen-assets: a part with no point")
		}
		if f.Kind == scene.GeomPoint {
			step(pts[0], pts[1])
			continue
		}
		n := len(pts) / 2
		closed := f.Kind == scene.GeomPolygon && n >= 2 && pts[0] == pts[len(pts)-2] && pts[1] == pts[len(pts)-1]
		if closed {
			n-- // ClosePath pushes the first point again
		}
		b = putVarint(b, cmdMoveTo|1<<3)
		step(pts[0], pts[1])
		if n > 1 {
			b = putVarint(b, cmdLineTo|uint64(n-1)<<3)
			for i := 1; i < n; i++ {
				step(pts[2*i], pts[2*i+1])
			}
		}
		if closed {
			b = putVarint(b, cmdClosePath|1<<3)
		}
	}
	return b, nil
}

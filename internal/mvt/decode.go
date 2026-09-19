package mvt

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// defaultExtent is a layer's coordinate grid when the tile does not say.
const defaultExtent = 4096

// Limits bound what one tile may cost. Each is checked before memory is set
// aside for what it bounds (NFR-10). A host may lower them, never raise
// them past what the coordinate type can hold.
type Limits struct {
	BodyBytes         int // the tile as received
	DecompressedBytes int // the tile after gzip
	Layers            int // layers in the tile, kept or not
	Features          int // features in the kept layers
	GeometryIntegers  int // geometry integers in the kept layers
	Keys              int // attribute names in one kept layer
	Values            int // attribute values in one kept layer
	RetainedBytes     int // the kept form
	MaxExtent         int // the largest coordinate grid a layer may declare
}

// DefaultLimits are the numbers of NFR-10 and the constants file, each with
// measured headroom over the heaviest real tiles.
func DefaultLimits() Limits {
	return Limits{
		BodyBytes:         2 << 20,
		DecompressedBytes: 8 << 20,
		Layers:            64,
		Features:          100000,
		GeometryIntegers:  2000000,
		Keys:              4096,
		Values:            400000,
		RetainedBytes:     4 << 20,
		MaxExtent:         8192,
	}
}

// Validate reports whether the limits can be used. A host may lower any of
// them. It may not remove one, and it may not raise one past its default:
// the extent because a coordinate is 16 bits wide, the others because the
// memory the library promises is stated against them.
func (l Limits) Validate() error {
	d := DefaultLimits()
	if l.BodyBytes < 1 || l.BodyBytes > d.BodyBytes {
		return overLimit()
	}
	if l.DecompressedBytes < 1 || l.DecompressedBytes > d.DecompressedBytes {
		return overLimit()
	}
	if l.RetainedBytes < 1 || l.RetainedBytes > d.RetainedBytes {
		return overLimit()
	}
	if l.MaxExtent < 1 || l.MaxExtent > d.MaxExtent {
		return overLimit()
	}
	return l.validateCounts(d)
}

// validateCounts is the half of Validate that bounds how many of each thing
// a tile may hold.
func (l Limits) validateCounts(d Limits) error {
	if l.Layers < 1 || l.Layers > d.Layers {
		return overLimit()
	}
	if l.Features < 1 || l.Features > d.Features {
		return overLimit()
	}
	if l.GeometryIntegers < 1 || l.GeometryIntegers > d.GeometryIntegers {
		return overLimit()
	}
	if l.Keys < 1 || l.Keys > d.Keys {
		return overLimit()
	}
	if l.Values < 1 || l.Values > d.Values {
		return overLimit()
	}
	return nil
}

// Want says what to keep: the layers the schema mapping draws, and the one
// label language (D-82). Everything else is dropped while decoding.
type Want struct {
	Layers   []string
	Language string
}

// overLimit is the error for a tile that is larger than a limit allows.
func overLimit() error {
	return fault.New(fault.OverLimit,
		textsafe.Const("the tile was refused"),
		textsafe.Const("it is larger than a limit allows: its size, its layers, its features or its geometry"),
		textsafe.Const("check the tile source; the limits can be lowered by a host, not raised past what the library can hold"))
}

// unsupported is the error for bytes that are some other format.
func unsupported() error {
	return fault.New(fault.UnsupportedTile,
		textsafe.Const("the tile is not a vector tile"),
		textsafe.Const("it is an image, a page of text, or compressed in a way other than gzip"),
		textsafe.Const("name a source of vector tiles in the supported schema"))
}

// Decode turns a tile's bytes into its kept form. The bytes are hostile
// until proved otherwise: every limit is checked before anything is
// allocated for what it bounds.
func Decode(body []byte, want Want, lim Limits) (*scene.Tile, error) {
	err := lim.Validate()
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, malformed()
	}
	if len(want.Layers) == 0 {
		return &scene.Tile{}, nil // nothing is drawn, so nothing is decoded
	}
	if len(body) > lim.BodyBytes {
		return nil, overLimit()
	}
	data, err := inflate(body, lim)
	if err != nil {
		return nil, err
	}
	return decodeTile(data, want, lim)
}

// inflate returns the tile's protobuf bytes: the body itself, or what gzip
// holds, known by its first two bytes as upstream knows it (P-37). Other
// formats are told apart here, so a wrong source gets a clear error.
func inflate(body []byte, lim Limits) ([]byte, error) {
	if len(body) < 2 {
		return nil, malformed()
	}
	if body[0] != 0x1f || body[1] != 0x8b {
		if isOtherFormat(body) {
			return nil, unsupported()
		}
		return body, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, malformed()
	}
	data, err := io.ReadAll(io.LimitReader(r, int64(lim.DecompressedBytes)+1))
	if err != nil {
		return nil, malformed()
	}
	if len(data) > lim.DecompressedBytes {
		return nil, overLimit()
	}
	return data, nil
}

// isOtherFormat reports bytes that are plainly something else: an image, a
// page, a JSON error, or a compression the library does not read.
func isOtherFormat(b []byte) bool {
	for _, magic := range [][]byte{{0x89, 'P', 'N', 'G'}, {0xff, 0xd8, 0xff}, []byte("RIFF"), {0x28, 0xb5, 0x2f, 0xfd}, []byte("<"), []byte("{")} {
		if bytes.HasPrefix(b, magic) {
			return true
		}
	}
	return false
}

// decodeTile walks the tile's layers. Every layer counts against the limit,
// kept or not; a layer the map does not draw is passed over without
// anything being allocated for it.
func decodeTile(data []byte, want Want, lim Limits) (*scene.Tile, error) {
	tile := &scene.Tile{}
	layers, rest := 0, data
	d := layerDecoder{tile: tile, want: want, lim: lim} // one decoder, so the limits hold across layers
	for range len(data) {                               // a field is at least two bytes, so this bounds the walk
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return nil, err
		}
		rest = next
		if f.num != 3 || f.wire != wireBytes {
			continue // an unknown field is skipped, as the format allows
		}
		layers++
		if layers > lim.Layers {
			return nil, overLimit()
		}
		err = d.decode(f.body)
		if err != nil {
			return nil, err
		}
		if tile.Bytes() > lim.RetainedBytes {
			return nil, overLimit()
		}
	}
	return tile, nil
}

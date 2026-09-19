package mvt

import (
	"math"
	"slices"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// What an attribute key is for. Every other key is dropped while decoding
// and its values are never read.
const (
	roleNone uint8 = iota
	roleClass
	roleNameLang // name_<lang> or name:<lang>, the one configured language (D-82)
	roleName     // the local name
	roleHouseNum
	roleLocalRank
	roleScaleRank
)

// layerDecoder decodes a tile's layers, one after another, into the tile.
// The counters run across layers, so the limits hold for the whole tile.
type layerDecoder struct {
	tile     *scene.Tile
	want     Want
	lim      Limits
	features int // kept features so far, in the whole tile
	integers int // geometry integers so far, in the whole tile

	roles   []uint8 // what each key of this layer is for
	values  []span  // where each Value message of this layer lies in the layer's bytes
	body    []byte  // the layer's bytes, which the spans index
	classes map[string]string

	layer *scene.Layer // the layer being decoded
	count counts       // what it holds, counted before anything is allocated
}

// counts is what a layer's features hold, found without decoding them.
type counts struct {
	features int // features in the bytes; a feature of several polygons becomes more
	integers int // geometry integers, for the limit
	coords   int // coordinate values the geometry decodes to, exactly
	parts    int // parts: points, lines, rings
}

// span is a stretch of the layer's bytes: eight bytes, where a slice is 24.
type span struct {
	start uint32
	end   uint32
}

// header is what a layer says about itself before its features are read.
type header struct {
	name    []byte
	extent  uint64
	version uint64
	keys    int // how many keys and values it holds, so their tables are
	values  int // allocated once, at their exact size
}

// decode reads one layer. The first pass finds the name and stops there if
// the map does not draw the layer: nothing is allocated for it.
func (d *layerDecoder) decode(body []byte) error {
	h, err := readHeader(body)
	if err != nil {
		return err
	}
	if !wanted(h.name, d.want) {
		return nil
	}
	if h.version != 1 && h.version != 2 {
		return malformed()
	}
	if h.extent == 0 || h.extent > uint64(d.lim.MaxExtent) {
		return overLimit()
	}
	if h.keys > d.lim.Keys || h.values > d.lim.Values {
		return overLimit()
	}
	err = d.readTables(body, h)
	if err != nil {
		return err
	}
	// The counts are known; the limits are checked before the slabs grow.
	d.integers += d.count.integers
	if d.integers > d.lim.GeometryIntegers || d.features+d.count.features > d.lim.Features {
		return overLimit()
	}
	d.tile.Layers = append(d.tile.Layers, scene.Layer{
		Name:     string(h.name),
		Extent:   uint16(h.extent),
		Features: make([]scene.Feature, 0, d.count.features),
		Coords:   make([]int16, 0, d.count.coords),
		Parts:    make([]uint32, 0, d.count.parts),
	})
	d.layer = &d.tile.Layers[len(d.tile.Layers)-1]
	return d.readFeatures(body)
}

// readHeader finds a layer's name, extent and version. It allocates nothing.
func readHeader(body []byte) (header, error) {
	h, rest := header{extent: defaultExtent, version: 1}, body
	for range len(body) { // a field is at least two bytes
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return header{}, err
		}
		rest = next
		switch {
		case f.num == 1 && f.wire == wireBytes:
			h.name = f.body
		case f.num == 5 && f.wire == wireVarint:
			h.extent = f.value
		case f.num == 15 && f.wire == wireVarint:
			h.version = f.value
		case f.num == 3 && f.wire == wireBytes:
			h.keys++
		case f.num == 4 && f.wire == wireBytes:
			h.values++
		}
	}
	return h, nil
}

// wanted reports whether the map draws the layer called name. The
// comparison allocates nothing.
func wanted(name []byte, want Want) bool {
	for _, w := range want.Layers {
		if string(name) == w {
			return true
		}
	}
	return false
}

// readTables notes what each key is for and where each value lies. Keys and
// values may follow the features in the bytes, so this is a pass of its own.
// A key's text is compared and never kept; a value is not read until a kept
// key points at it.
func (d *layerDecoder) readTables(body []byte, h header) error {
	if len(body) > math.MaxUint32 {
		return overLimit()
	}
	d.body = body
	d.roles, d.values = slices.Grow(d.roles[:0], h.keys), slices.Grow(d.values[:0], h.values)
	d.count = counts{}
	rest := body
	for range len(body) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return err
		}
		rest = next
		if f.wire != wireBytes {
			continue
		}
		switch f.num {
		case 2:
			err = d.count.add(f.body)
			if err != nil {
				return err
			}
		case 3:
			if len(d.roles) >= d.lim.Keys {
				return overLimit()
			}
			d.roles = append(d.roles, roleOf(f.body, d.want.Language))
		case 4:
			if len(d.values) >= d.lim.Values {
				return overLimit()
			}
			end := len(body) - len(next)
			d.values = append(d.values, span{start: uint32(end - len(f.body)), end: uint32(end)})
		}
	}
	return nil
}

// roleOf says what a key is for. The label's lookup order is upstream's
// (P-36) with one language kept: name_<lang>, name:<lang>, the local name,
// the house number. No other language's name is ever materialised.
func roleOf(key []byte, language string) uint8 {
	switch string(key) {
	case "class":
		return roleClass
	case "name":
		return roleName
	case "house_num":
		return roleHouseNum
	case "localrank":
		return roleLocalRank
	case "scalerank":
		return roleScaleRank
	}
	if len(key) == len("name_")+len(language) && string(key[:4]) == "name" && (key[4] == '_' || key[4] == ':') && string(key[5:]) == language {
		return roleNameLang
	}
	return roleNone
}

// readFeatures decodes the layer's features into the tile.
func (d *layerDecoder) readFeatures(body []byte) error {
	rest := body
	for range len(body) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return err
		}
		rest = next
		if f.num != 2 || f.wire != wireBytes {
			continue
		}
		d.features++
		if d.features > d.lim.Features {
			return overLimit()
		}
		err = d.readFeature(f.body)
		if err != nil {
			return err
		}
	}
	return nil
}

// add counts one feature's geometry without decoding its coordinates: the
// integers, for the limit, and exactly how many coordinate values and parts
// it will decode to, so the layer's slabs are allocated once and at their
// true size. A count that lies is caught later, when the commands are run.
func (c *counts) add(feature []byte) error {
	rest := feature
	c.features++
	for range len(feature) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return err
		}
		rest = next
		if f.num != 4 || f.wire != wireBytes {
			continue
		}
		err = c.addGeometry(f.body)
		if err != nil {
			return err
		}
	}
	return nil
}

// addGeometry walks the command integers only, skipping each command's
// points by counting their terminating bytes.
func (c *counts) addGeometry(geometry []byte) error {
	rest := geometry
	for range len(geometry) {
		if len(rest) == 0 {
			break
		}
		command, n, err := uvarint(rest)
		if err != nil {
			return err
		}
		rest = rest[n:]
		c.integers++
		id, count := command&7, command>>3
		if id == cmdClosePath {
			c.coords += 2 // the ring's first point, pushed again
			continue
		}
		if count > uint64(len(rest))/2 {
			return malformed()
		}
		if id == cmdMoveTo {
			c.parts += int(count)
		}
		c.integers += 2 * int(count)
		c.coords += 2 * int(count)
		rest, err = skipVarints(rest, 2*int(count))
		if err != nil {
			return err
		}
	}
	return nil
}

// skipVarints passes over n varints: each ends in exactly one byte with the
// high bit clear.
func skipVarints(b []byte, n int) ([]byte, error) {
	if n == 0 {
		return b, nil
	}
	for i, c := range b {
		if c >= 0x80 {
			continue
		}
		n--
		if n == 0 {
			return b[i+1:], nil
		}
	}
	return nil, malformed()
}

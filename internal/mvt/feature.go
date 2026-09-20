package mvt

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// attrs is what is kept of one feature's attributes.
type attrs struct {
	class     string
	nameLang  string
	name      string
	houseNum  string
	localRank []byte // the Value message, read only if no better rank is found
	scaleRank []byte
	rankValue []byte
	level     uint8
	maritime  bool
}

// readFeature decodes one feature: its kind, the attributes the map reads,
// and its geometry. A feature of unknown kind (0) is passed over, as the
// format allows.
func (d *layerDecoder) readFeature(body []byte) error {
	if d == nil || d.layer == nil {
		return malformed() // a feature arrived with no layer to put it in
	}
	kind, tags, geometry, err := featureFields(body)
	if err != nil {
		return err
	}
	if kind == 0 {
		d.passOver()
		return nil
	}
	if kind > uint64(scene.GeomPolygon) {
		return malformed()
	}
	a, err := d.readTags(tags)
	if err != nil {
		return err
	}
	feature := scene.Feature{Kind: scene.GeomKind(kind), Class: a.class, Name: textsafe.Clean(a.label()), Rank: a.rank(), AdminLevel: a.level, Maritime: a.maritime}
	return d.readGeometry(geometry, feature)
}

// featureFields walks one feature's bytes and answers what it says it is,
// where its tags are and where its geometry is. Tags and geometry must be
// packed, as the format requires; anything else is a feature this decoder
// will not guess at.
func featureFields(body []byte) (kind uint64, tags, geometry []byte, err error) {
	if len(body) == 0 {
		return 0, nil, nil, nil // a feature of no bytes says nothing, and is passed over
	}
	rest := body
	for range len(body) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return 0, nil, nil, err
		}
		rest = next
		switch {
		case f.num == 3 && f.wire == wireVarint:
			kind = f.value
		case f.num == 2 && f.wire == wireBytes:
			tags = f.body
		case f.num == 4 && f.wire == wireBytes:
			geometry = f.body
		case f.num == 2 || f.num == 4:
			return 0, nil, nil, malformed() // tags and geometry must be packed
		}
	}
	return kind, tags, geometry, nil
}

// passOver counts a feature that does not say what it is. Such a feature
// cannot be drawn - there is no telling whether to fill it or stroke it -
// so it is dropped, and counted so that a reader can tell this from a
// feature that was lost (D-118).
func (d *layerDecoder) passOver() {
	if d == nil || d.layer == nil {
		return
	}
	if d.layer.Untyped == math.MaxUint16 {
		return // a tile with sixty-five thousand untyped features says enough
	}
	d.layer.Untyped++
}

// readTags walks a feature's key and value indexes and keeps what the map
// reads. An index outside its table is an error.
func (d *layerDecoder) readTags(tags []byte) (attrs, error) {
	var a attrs
	rest := tags
	for range len(tags) {
		if len(rest) == 0 {
			break
		}
		key, n, err := uvarint(rest)
		if err != nil {
			return attrs{}, err
		}
		rest = rest[n:]
		value, n, err := uvarint(rest) // an odd count of integers fails here
		if err != nil {
			return attrs{}, err
		}
		rest = rest[n:]
		if key >= uint64(len(d.roles)) || value >= uint64(len(d.values)) {
			return attrs{}, malformed()
		}
		v := d.values[value]
		err = d.keep(&a, d.roles[key], d.body[v.start:v.end])
		if err != nil {
			return attrs{}, err
		}
	}
	return a, nil
}

// keep stores one attribute by its role. A class is shared between the
// features of a tile; a name is copied, once.
func (d *layerDecoder) keep(a *attrs, role uint8, value []byte) error {
	if role == roleNone {
		return nil
	}
	if role == roleLocalRank {
		a.localRank = value
		return nil
	}
	if role == roleScaleRank {
		a.scaleRank = value
		return nil
	}
	if role >= roleRank {
		a.keepSchema(role, value)
		return nil
	}
	text, err := stringOf(value)
	if err != nil {
		return err
	}
	switch role {
	case roleClass:
		a.class = d.intern(text)
	case roleNameLang:
		a.nameLang = string(text)
	case roleName:
		a.name = string(text)
	case roleHouseNum:
		a.houseNum = string(text)
	}
	return nil
}

// keepSchema stores the attributes OpenMapTiles carries where upstream's
// schema had a class or a local rank. A level outside 1 to 11, which is
// every level there is, is kept as none.
func (a *attrs) keepSchema(role uint8, value []byte) {
	if a == nil {
		return
	}
	switch role {
	case roleRank:
		a.rankValue = value
	case roleAdminLevel:
		if n := intOf(value); n >= 1 && n <= 11 {
			a.level = uint8(n)
		}
	case roleMaritime:
		a.maritime = intOf(value) == 1
	}
}

// intern returns one shared string for each distinct class in a tile.
func (d *layerDecoder) intern(text []byte) string {
	if d.classes == nil {
		d.classes = make(map[string]string)
	}
	if s, ok := d.classes[string(text)]; ok { // the lookup allocates nothing
		return s
	}
	s := string(text)
	d.classes[s] = s
	return s
}

// label is the feature's name by upstream's order with one language kept.
func (a attrs) label() string {
	if a.nameLang != "" {
		return a.nameLang
	}
	if a.name != "" {
		return a.name
	}
	return a.houseNum
}

// rank is upstream's sort key: the local rank, else the scale rank, else -
// what OpenMapTiles really carries - the rank, else 0; integers only. As upstream, a local rank that is present and not an
// integer gives 0: the scale rank is not read in its place.
func (a attrs) rank() int32 {
	value := a.localRank
	if value == nil {
		value = a.scaleRank
	}
	if value == nil {
		value = a.rankValue
	}
	if value == nil {
		return 0
	}
	return intOf(value)
}

// stringOf returns the string a Value message holds, aliasing its bytes; a
// value of another type gives nothing.
func stringOf(value []byte) ([]byte, error) {
	rest := value
	for range len(value) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return nil, err
		}
		rest = next
		if f.num == 1 && f.wire == wireBytes {
			return f.body, nil
		}
	}
	return nil, nil
}

// intOf returns the integer a Value message holds - int, unsigned or
// zig-zag signed - held to 32 bits; any other type gives 0.
func intOf(value []byte) int32 {
	const limit = 1 << 31
	rest := value
	found, fields := int32(0), 0
	for range len(value) {
		if len(rest) == 0 {
			break
		}
		f, next, err := nextField(rest)
		if err != nil {
			return 0
		}
		rest = next
		fields++
		if f.wire != wireVarint || f.value >= limit {
			continue
		}
		switch f.num {
		case 4, 5:
			found = int32(f.value)
		case 6:
			found = zigzag(uint32(f.value))
		}
	}
	// **A value carries one field.** One with several is damaged, and
	// taking a number out of it is guessing at what was meant: the first
	// number and the last number were both tried against the proven
	// decoder and each disagreed with it, in opposite directions, on
	// different damaged values. This decoder refuses to guess, which is
	// the rule it keeps everywhere else (D-75: the decoder refuses a
	// malformed stream rather than reading on).
	if fields != 1 {
		return 0
	}
	return found
}

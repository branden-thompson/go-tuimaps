package mvt

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// The geometry commands.
const (
	cmdMoveTo    = 1
	cmdLineTo    = 2
	cmdClosePath = 7
)

// cursor is where the geometry's pen is, and the part being drawn.
type cursor struct {
	x, y      int64
	partStart int // index in the tile's Coords where the open part began; -1 when none is open
}

// readGeometry decodes a feature's commands into the layer's slabs and adds
// the feature - or, for polygons, one feature for each outline with its
// holes - to the layer. The integers were counted, and the limit checked,
// before the layer's slabs were grown (layer.go).
func (d *layerDecoder) readGeometry(geometry []byte, feature scene.Feature) error {
	firstPart := uint32(len(d.layer.Parts))
	err := d.runCommands(geometry, feature.Kind)
	if err != nil {
		return err
	}
	endPart := uint32(len(d.layer.Parts))
	if endPart == firstPart {
		return nil // no geometry: nothing to draw
	}
	if feature.Kind != scene.GeomPolygon {
		feature.FirstPart, feature.EndPart = firstPart, endPart
		d.layer.Features = append(d.layer.Features, feature)
		return nil
	}
	return d.groupRings(feature, firstPart, endPart)
}

// runCommands executes MoveTo, LineTo and ClosePath. A malformed stream is
// an error, not a guess: a count beyond the integers left, LineTo or
// ClosePath with no part open, a ClosePath count other than one, a count of
// zero, an unknown command, or a pen that leaves the 16-bit range.
func (d *layerDecoder) runCommands(geometry []byte, kind scene.GeomKind) error {
	if kind < scene.GeomPoint || kind > scene.GeomPolygon {
		return malformed() // only the three kinds have commands to run
	}
	c := cursor{partStart: -1}
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
		id, count := command&7, command>>3
		if count == 0 {
			return malformed()
		}
		if id != cmdClosePath && count > uint64(len(rest))/2 {
			return malformed() // each point is two integers of at least a byte each
		}
		switch {
		case id == cmdMoveTo:
			rest, err = d.moveTo(&c, rest, int(count), kind)
		case kind == scene.GeomPoint:
			// **A point has no line and no ring.** Reading these anyway grew
			// a single point into a run of positions, which the proven
			// decoder does not do - the second disagreement the oracle's
			// fuzzer found (D-126). A damaged stream is refused (D-75).
			err = malformed()
		case id == cmdLineTo:
			rest, err = d.lineTo(&c, rest, int(count))
		case id == cmdClosePath:
			err = d.closePath(&c, count)
		default:
			err = malformed()
		}
		if err != nil {
			return err
		}
	}
	d.endPart(&c)
	return nil
}

// step reads one dx, dy pair, moves the pen and appends the point.
func (d *layerDecoder) step(c *cursor, rest []byte) ([]byte, error) {
	dx, n, err := uvarint(rest)
	if err != nil {
		return nil, err
	}
	rest = rest[n:]
	dy, n, err := uvarint(rest)
	if err != nil {
		return nil, err
	}
	if dx > math.MaxUint32 || dy > math.MaxUint32 {
		return nil, malformed()
	}
	c.x += int64(zigzag(uint32(dx)))
	c.y += int64(zigzag(uint32(dy)))
	if c.x < math.MinInt16 || c.x > math.MaxInt16 || c.y < math.MinInt16 || c.y > math.MaxInt16 {
		return nil, malformed() // an error, never a wrap
	}
	d.layer.Coords = append(d.layer.Coords, int16(c.x), int16(c.y))
	return rest[n:], nil
}

// moveTo starts a new part at each of its points. For points, each is a
// part of its own, as upstream has it.
func (d *layerDecoder) moveTo(c *cursor, rest []byte, count int, kind scene.GeomKind) ([]byte, error) {
	if kind != scene.GeomPoint && count != 1 {
		return nil, malformed()
	}
	for range count {
		d.endPart(c)
		c.partStart = len(d.layer.Coords)
		next, err := d.step(c, rest)
		if err != nil {
			return nil, err
		}
		rest = next
	}
	return rest, nil
}

// lineTo extends the open part.
func (d *layerDecoder) lineTo(c *cursor, rest []byte, count int) ([]byte, error) {
	if c.partStart < 0 {
		return nil, malformed() // LineTo before MoveTo
	}
	for range count {
		next, err := d.step(c, rest)
		if err != nil {
			return nil, err
		}
		rest = next
	}
	return rest, nil
}

// closePath re-pushes the ring's first point, as upstream does, and ends
// the part. The pen does not move. **A ring whose last point is already its
// first is closed, and gets nothing more**: repeating the point would add a
// segment of no length that paulmach/orb's decoder - the proven decoder the
// oracle compares against (tools/oracle) - does not have (v0.2.0 D-16).
func (d *layerDecoder) closePath(c *cursor, count uint64) error {
	if count != 1 {
		return malformed()
	}
	if c.partStart < 0 || len(d.layer.Coords)-c.partStart < 2 {
		return malformed()
	}
	first, last := c.partStart, len(d.layer.Coords)-2
	if d.layer.Coords[last] != d.layer.Coords[first] || d.layer.Coords[last+1] != d.layer.Coords[first+1] {
		d.layer.Coords = append(d.layer.Coords, d.layer.Coords[first], d.layer.Coords[first+1])
	}
	d.endPart(c)
	return nil
}

// endPart closes the open part, if there is one.
func (d *layerDecoder) endPart(c *cursor) {
	if c.partStart < 0 {
		return
	}
	d.layer.Parts = append(d.layer.Parts, uint32(len(d.layer.Coords)))
	c.partStart = -1
}

// groupRings splits a polygon feature as upstream does (P-39): a ring whose
// signed area is zero or more starts a new polygon, a negative one is a
// hole in the polygon before it, and each polygon becomes a feature of its
// own. A hole with no outline before it is a polygon of its own.
func (d *layerDecoder) groupRings(feature scene.Feature, firstPart, endPart uint32) error {
	if firstPart >= endPart || int(endPart) > len(d.layer.Parts) {
		return malformed()
	}
	start := firstPart
	for p := firstPart + 1; p <= endPart; p++ {
		if p < endPart && d.signedArea(p) < 0 {
			continue // a hole: it stays with the polygon before it
		}
		d.features++
		if d.features > d.lim.Features {
			return overLimit()
		}
		feature.FirstPart, feature.EndPart = start, p
		d.layer.Features = append(d.layer.Features, feature)
		start = p
	}
	d.features-- // the tile's own feature was counted once already
	return nil
}

// signedArea is twice the ring's area by upstream's own sum: positive for a
// ring drawn clockwise on a screen, where y grows downward - an outline.
func (d *layerDecoder) signedArea(part uint32) int64 {
	from := uint32(0)
	if part > 0 {
		from = d.layer.Parts[part-1]
	}
	ring := d.layer.Coords[from:d.layer.Parts[part]]
	if len(ring) < 6 {
		return 0
	}
	sum := int64(0)
	j := len(ring) - 2
	for i := 0; i < len(ring); i += 2 {
		sum += (int64(ring[j]) - int64(ring[i])) * (int64(ring[j+1]) + int64(ring[i+1]))
		j = i
	}
	return sum
}

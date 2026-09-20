package render

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// firstLiteral is the first ink that stands for a user's style's literal
	// colour. The inks below it are the tokens' own numbers.
	firstLiteral = 96
	// maxLabels bounds the label candidates kept for one frame.
	maxLabels = 4096
)

// Label is a place name waiting to be placed: where it belongs, in dots, and
// how important it is - the lower the rank, the more.
type Label struct {
	X, Y       int // its first vertex
	Name       string
	Rank       int32
	Ink        uint8
	first, end int // its vertices, among the painter's: each is tried in turn (P-32)
}

// Painter paints the basemap of one frame: line work as dots, areas as the
// cells they own, and the names that want placing.
type Painter struct {
	lines      *Canvas
	areas      *Canvas
	literals   []colour.RGB
	labels     []Label
	labelPts   []Point
	ring       []Point
	edge       []bool // for each point of ring: the segment that ends there lies along the tile's border
	rings      [][]Point
	culled     int
	lastPoints int
}

// NewPainter makes a painter for a map of cols by rows cells.
func NewPainter(cols, rows int) (*Painter, error) {
	lines, err := NewCanvas(cols, rows)
	if err != nil {
		return nil, err
	}
	areas, err := NewCanvas(cols, rows)
	if err != nil {
		return nil, err
	}
	return &Painter{lines: lines, areas: areas}, nil
}

// Reset clears the painter for another frame, keeping its buffers.
func (p *Painter) Reset() {
	if p == nil {
		return
	}
	p.lines.Wipe()
	p.areas.Wipe()
	p.literals, p.labels, p.labelPts = p.literals[:0], p.labels[:0], p.labelPts[:0]
	p.culled, p.lastPoints = 0, 0
}

// Lines is the canvas of line work.
func (p *Painter) Lines() *Canvas {
	if p == nil {
		return nil
	}
	return p.lines
}

// Labels are the names found, in the order the tiles gave them.
func (p *Painter) Labels() []Label {
	if p == nil {
		return nil
	}
	return p.labels
}

// Points are the vertices a label may be anchored at, in order.
func (p *Painter) Points(l Label) []Point {
	if p == nil || l.first < 0 || l.end > len(p.labelPts) || l.first > l.end {
		return nil
	}
	return p.labelPts[l.first:l.end]
}

// Culled is how many features this frame were passed over because their box
// misses the view (P-28).
func (p *Painter) Culled() int {
	if p == nil {
		return 0
	}
	return p.culled
}

// World makes everything beyond the world's edge ocean, in the style's water
// colour (P-22, L-6). It is called before any tile is painted.
func (p *Painter) World(v project.View) error {
	if p == nil {
		return badTile()
	}
	x, y, side, err := v.TilePlace(scene.TileID{})
	if err != nil {
		return err
	}
	w, h := p.areas.Dots()
	left, top, right, bottom := toDot(x), toDot(y), toDot(x+side), toDot(y+side)
	ink := uint8(colour.WaterFill)
	for _, r := range [4][4]int{{0, 0, left, h}, {right, 0, w, h}, {0, 0, w, top}, {0, bottom, w, h}} {
		if r[0] >= r[2] || r[1] >= r[3] {
			continue
		}
		p.areas.Fill([][]Point{{{r[0], r[1]}, {r[2], r[1]}, {r[2], r[3]}, {r[0], r[3]}}}, ink)
	}
	return nil
}

// Area is the ink of the area that owns a cell: the area covering at least
// half of the cell's dots. Water owns its cells as their background; it is
// not drawn in dots (L2 Render, step 2).
func (p *Painter) Area(col, row int) (uint8, bool) {
	if p == nil {
		return 0, false
	}
	glyph, ink := p.areas.Cell(col, row)
	lit := 0
	for mask := uint8(glyph - blank); mask != 0; mask &= mask - 1 {
		lit++
	}
	return ink, lit >= 4
}

// Water reports whether water owns a cell.
func (p *Painter) Water(col, row int) bool {
	ink, owned := p.Area(col, row)
	return owned && ink == uint8(colour.WaterFill)
}

// Colour is an ink's colour: a token's, through the palette and the ground in
// effect, or a user's style's literal colour.
func (p *Painter) Colour(ink uint8, palette colour.Palette, ground colour.GroundKind) (colour.RGB, bool) {
	if p == nil || ink == 0 {
		return colour.RGB{}, false
	}
	if ink < firstLiteral {
		return palette.Resolve(colour.Token(ink), ground)
	}
	if int(ink-firstLiteral) >= len(p.literals) {
		return colour.RGB{}, false
	}
	return p.literals[ink-firstLiteral], true
}

// inkFor is the ink a rule draws in at a zoom.
func (p *Painter) inkFor(r *style.Rule, zoom float64) uint8 {
	literal, isLiteral := r.Colour(zoom)
	if !isLiteral {
		return uint8(r.Token)
	}
	for i, c := range p.literals {
		if c == literal {
			return uint8(firstLiteral + i)
		}
	}
	if len(p.literals) >= 256-firstLiteral {
		return uint8(firstLiteral) // a style with more colours than inks shares the first
	}
	p.literals = append(p.literals, literal)
	return uint8(firstLiteral + len(p.literals) - 1)
}

// toDot rounds a coordinate to 1/256 of a dot and takes the dot it falls in,
// so that the same tile gives the same dots wherever it is computed (NFR-6).
func toDot(v float64) int {
	return int(math.Floor(math.Round(v*256) / 256))
}

// place is a tile coordinate's dot: the tile's corner plus the coordinate
// scaled. The product is converted before it is added, so that no machine
// fuses the two into one rounding (constants, section 6).
func place(corner float64, coord int16, scale float64) int {
	return toDot(corner + float64(float64(coord)*scale))
}

func badTile() error {
	return fault.Make(fault.Internal, textsafe.Const("a tile could not be drawn"),
		textsafe.Const("the painter was given no tile or no style, or the tile's parts run past its coordinates"),
		textsafe.Const("this is a defect in the library; report it"))
}

// frame is one tile's place in the view.
type frame struct {
	x, y, scale float64
	extent      int16
	zoom        float64
}

// Tile paints one tile on hand. at is the tile it is - the wanted tile, or an
// ancestor standing in for it, which is drawn larger (D-30).
func (p *Painter) Tile(v project.View, tile *scene.Tile, at scene.TileID, s *style.Style) error {
	if p == nil || tile == nil || s == nil {
		return badTile()
	}
	x, y, side, err := v.TilePlace(at)
	if err != nil {
		return err
	}
	// Layers are drawn in one order, whatever order the tile carries them in
	// (P-23, P-24): areas, then line work from the least to the most
	// important, so that a border lying on a road shows as a border.
	for rank := range len(drawOrder()) + 1 {
		for i := range tile.Layers {
			l := &tile.Layers[i]
			if orderOf(l.Name) != rank || l.Extent == 0 || l.Extent > math.MaxInt16 {
				continue
			}
			f := frame{x: x, y: y, scale: side / float64(l.Extent), extent: int16(l.Extent), zoom: v.Zoom}
			for _, feature := range l.Features {
				err = p.feature(l, feature, f, s)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// drawOrder is the order layers are drawn in.
func drawOrder() []string {
	return []string{"water", "landcover", "park", "waterway", "aeroway", "transportation", "boundary", "water_name", "place", "aerodrome_label"}
}

// orderOf is a layer's place in the draw order; a layer not in it comes last.
func orderOf(name string) int {
	order := drawOrder()
	for i, n := range order {
		if n == name {
			return i
		}
	}
	return len(order)
}

// feature paints one feature by the rules that take it.
func (p *Painter) feature(l *scene.Layer, feature scene.Feature, f frame, s *style.Style) error {
	attrs := style.AttrsOf(feature)
	rule, drawn := s.Match(l.Name, attrs, f.zoom)
	fill, filled := s.Fill(l.Name, attrs, f.zoom)
	if !drawn && !filled {
		return nil
	}
	visible := p.project(l, feature, f)
	if p.lastPoints < 0 {
		return badTile()
	}
	if !visible {
		p.culled++ // its box misses the view: nothing of it is rastered (P-28)
		return nil
	}
	if drawn && rule.Kind == style.Line {
		at := 0
		for _, ring := range p.rings {
			p.stroke(ring, p.edge[at:at+len(ring)], feature.Kind == scene.GeomPolygon, f, rule)
			at += len(ring)
		}
	}
	if filled && feature.Kind == scene.GeomPolygon {
		p.areas.Fill(p.rings, p.inkFor(fill, f.zoom))
	}
	if drawn && rule.Kind == style.Symbol && len(p.ring) > 0 && len(p.labels) < maxLabels {
		first := len(p.labelPts)
		p.labelPts = append(p.labelPts, p.ring...)
		p.labels = append(p.labels, Label{X: p.ring[0].X, Y: p.ring[0].Y, Name: feature.Name, Rank: feature.Rank, Ink: p.inkFor(rule, f.zoom), first: first, end: len(p.labelPts)})
	}
	return nil
}

// project turns a feature's parts into dots: floored, with consecutive points
// that fall on one dot kept once (P-29). It reports whether the feature's box
// meets the view grown by the clip pad, and leaves lastPoints negative if a
// part runs past the layer's coordinates.
func (p *Painter) project(l *scene.Layer, feature scene.Feature, f frame) bool {
	p.rings, p.ring, p.edge = p.rings[:0], p.ring[:0], p.edge[:0]
	w, h := p.lines.Dots()
	box := [4]int{math.MaxInt, math.MaxInt, math.MinInt, math.MinInt}
	for part := feature.FirstPart; part < feature.EndPart; part++ {
		coords, err := l.Part(int(part))
		if err != nil {
			p.lastPoints = -1
			return false
		}
		start, prev := len(p.ring), -1
		for i := 0; i+1 < len(coords); i += 2 {
			pt := Point{X: place(f.x, coords[i], f.scale), Y: place(f.y, coords[i+1], f.scale)}
			if len(p.ring) > start && p.ring[len(p.ring)-1] == pt {
				continue
			}
			p.ring = append(p.ring, pt)
			p.edge = append(p.edge, prev >= 0 && onBorder(coords[prev], coords[prev+1], coords[i], coords[i+1], f.extent))
			prev = i
			box = [4]int{min(box[0], pt.X), min(box[1], pt.Y), max(box[2], pt.X), max(box[3], pt.Y)}
		}
		p.rings = append(p.rings, p.ring[start:len(p.ring):len(p.ring)])
	}
	p.lastPoints = len(p.ring)
	return box[2] >= -clipMargin && box[0] < w+clipMargin && box[3] >= -clipMargin && box[1] < h+clipMargin
}

// onBorder reports whether a polygon's edge lies along or beyond one side of
// its tile: that edge is where the tile - or at zoom 0 the world - was cut,
// not a coast. Real tiles cut a sliver inside the side as often as on it, so
// an edge counts when both its ends are within 1/2048 of the extent of the
// side: two units of 4096, a sixteenth of a dot at the tile's own zoom.
func onBorder(x0, y0, x1, y1, extent int16) bool {
	if extent <= 0 {
		return false
	}
	lo, hi := extent>>11, extent-extent>>11
	return (x0 <= lo && x1 <= lo) || (x0 >= hi && x1 >= hi) || (y0 <= lo && y1 <= lo) || (y0 >= hi && y1 >= hi)
}

// stroke draws a part as a line: a line feature, or a polygon's edge.
func (p *Painter) stroke(dots []Point, border []bool, ring bool, f frame, rule *style.Rule) {
	if len(dots) < 2 || rule == nil || len(border) != len(dots) {
		return
	}
	ink := p.inkFor(rule, f.zoom)
	width := int(math.Round(rule.Width(f.zoom)))
	for i := 0; i+1 < len(dots); i++ {
		if ring && border[i+1] {
			continue
		}
		p.lines.Line(dots[i].X, dots[i].Y, dots[i+1].X, dots[i+1].Y, width, ink)
	}
}

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
	X, Y int
	Name string
	Rank int32
	Ink  uint8
}

// Painter paints the basemap of one frame: line work as dots, areas as the
// cells they own, and the names that want placing.
type Painter struct {
	lines    *Canvas
	areas    *Canvas
	literals []colour.RGB
	labels   []Label
	ring     []Point
	rings    [][]Point
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
	p.literals, p.labels = p.literals[:0], p.labels[:0]
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
	return fault.New(fault.Internal, textsafe.Const("a tile could not be drawn"),
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
	for i := range tile.Layers {
		l := &tile.Layers[i]
		if l.Extent == 0 || l.Extent > math.MaxInt16 {
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
	return nil
}

// feature paints one feature by the rules that take it.
func (p *Painter) feature(l *scene.Layer, feature scene.Feature, f frame, s *style.Style) error {
	attrs := style.AttrsOf(feature)
	rule, drawn := s.Match(l.Name, attrs, f.zoom)
	fill, filled := s.Fill(l.Name, attrs, f.zoom)
	if !drawn && !filled {
		return nil
	}
	p.rings = p.rings[:0]
	p.ring = p.ring[:0]
	for part := feature.FirstPart; part < feature.EndPart; part++ {
		coords, err := l.Part(int(part))
		if err != nil {
			return badTile()
		}
		start := len(p.ring)
		for i := 0; i+1 < len(coords); i += 2 {
			p.ring = append(p.ring, Point{X: place(f.x, coords[i], f.scale), Y: place(f.y, coords[i+1], f.scale)})
		}
		p.rings = append(p.rings, p.ring[start:len(p.ring):len(p.ring)])
		if drawn && rule.Kind == style.Line {
			p.stroke(coords, p.ring[start:], feature.Kind == scene.GeomPolygon, f, rule)
		}
	}
	if filled && feature.Kind == scene.GeomPolygon {
		p.areas.Fill(p.rings, p.inkFor(fill, f.zoom))
	}
	if drawn && rule.Kind == style.Symbol && len(p.ring) > 0 && len(p.labels) < maxLabels {
		p.labels = append(p.labels, Label{X: p.ring[0].X, Y: p.ring[0].Y, Name: feature.Name, Rank: feature.Rank, Ink: p.inkFor(rule, f.zoom)})
	}
	return nil
}

// onBorder reports whether a polygon's edge lies along or beyond one side of
// its tile: that edge is where the tile was cut, not a coast.
func onBorder(coords []int16, i int, extent int16) bool {
	if i < 0 || i+3 >= len(coords) {
		return false
	}
	x0, y0, x1, y1 := coords[i], coords[i+1], coords[i+2], coords[i+3]
	return (x0 <= 0 && x1 <= 0) || (x0 >= extent && x1 >= extent) || (y0 <= 0 && y1 <= 0) || (y0 >= extent && y1 >= extent)
}

// stroke draws a part as a line: a line feature, or a polygon's edge.
func (p *Painter) stroke(coords []int16, dots []Point, ring bool, f frame, rule *style.Rule) {
	if len(dots) < 2 || rule == nil {
		return
	}
	ink := p.inkFor(rule, f.zoom)
	width := int(math.Round(rule.Width(f.zoom)))
	for i := 0; i+1 < len(dots); i++ {
		if ring && onBorder(coords, 2*i, f.extent) {
			continue
		}
		p.lines.Line(dots[i].X, dots[i].Y, dots[i+1].X, dots[i+1].Y, width, ink)
	}
}

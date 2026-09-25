package render

import (
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// MarkerPad is how far outside the rectangle a marker may be and still be
// drawn: twenty dots, as upstream (P-60). Beyond it nothing of it is drawn.
const MarkerPad = 20

// labelGap is how far right of a marker its label begins, in dots (P-60).
const labelGap = 4

// MarkerShape is how a place is drawn (P-58).
type MarkerShape uint8

// The marker shapes. A radius belongs to a ring and to a filled circle; the
// others are the sizes upstream draws them at.
const (
	MarkerDot     MarkerShape = iota + 1 // three dots by three
	MarkerCross                          // arms of three dots each way
	MarkerDiamond                        // radius three
	MarkerRing                           // a circle by the midpoint rule
	MarkerDisc                           // the same, filled
	MarkerGlyph                          // a character of the host's own
)

// Marker is one place drawn over the map (FR-26). Its ink is a colour token,
// which its label takes too (P-60).
type Marker struct {
	At     project.LonLat
	Shape  MarkerShape
	Radius int // a ring's and a filled circle's; three if it is not given
	Text   string
	Label  string
	Ink    uint8
	Blink  bool
	ID     string // the host's place, named in Frame.Dropped when its label cannot fit whole
}

// radius is the marker's radius in dots.
func (m Marker) radius() int {
	if m.Radius <= 0 {
		return 3
	}
	return min(m.Radius, 64)
}

// Mark draws one marker over everything beneath it (L2 Render, step 7), and
// keeps its label for the label pass. on is the phase of the blink: a marker
// that blinks is drawn in one half of it and not in the other (P-59a). A
// marker of a character is kept for the glyph pass, where text wins its cell.
func (p *Painter) Mark(v project.View, m Marker, on bool) error {
	if p == nil {
		return badTile()
	}
	x, y, err := v.ToDot(m.At)
	if err != nil {
		return err
	}
	at := Point{X: toDot(x), Y: toDot(y)}
	w, h := p.lines.Dots()
	if at.X < -MarkerPad || at.Y < -MarkerPad || at.X >= w+MarkerPad || at.Y >= h+MarkerPad {
		p.culled++ // more than twenty dots outside the rectangle (P-60)
		return nil
	}
	if m.Blink && !on {
		return nil
	}
	if m.Shape == MarkerGlyph {
		p.keepGlyph(at, m)
	} else {
		p.strokeMarker(at, m)
	}
	if m.Label != "" && len(p.markerLabels) < maxLabels {
		l := Label{X: at.X + labelGap, Y: at.Y, Name: textsafe.Clean(m.Label), Ink: m.Ink, fromPoint: true, Overlay: m.ID, place: true}
		if first, _, cut := strings.Cut(m.Label, " "); cut && first != "" {
			l.Short = textsafe.Clean(first) // a shorter form: the name's first word (L-8.9)
		}
		p.markerLabels = append(p.markerLabels, l)
	}
	return nil
}

// keepGlyph keeps a marker that is a character for the glyph pass.
func (p *Painter) keepGlyph(at Point, m Marker) {
	text := textsafe.Clean(m.Text)
	if textsafe.Width(text) == 0 || len(p.glyphs) >= maxLabels {
		return
	}
	p.glyphs = append(p.glyphs, Label{X: at.X, Y: at.Y, Name: text, Ink: m.Ink, fromPoint: true})
}

// strokeMarker lights the dots of a marker's shape. Each is forced, so that
// the marker shows whatever it is drawn over (P-09, FR-18a).
func (p *Painter) strokeMarker(at Point, m Marker) {
	switch m.Shape {
	case MarkerCross:
		p.markCross(at, m.Ink)
	case MarkerDiamond:
		p.markDiamond(at, 3, m.Ink)
	case MarkerRing:
		p.markCircle(at, m.radius(), false, m.Ink)
	case MarkerDisc:
		p.markCircle(at, m.radius(), true, m.Ink)
	default:
		p.markBlock(at, m.Ink)
	}
}

// markBlock is a dot: three dots by three (P-58).
func (p *Painter) markBlock(at Point, ink uint8) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			p.lines.SetForced(at.X+dx, at.Y+dy, ink)
		}
	}
}

// markCross is a cross: arms of three dots each way from its centre.
func (p *Painter) markCross(at Point, ink uint8) {
	p.lines.SetForced(at.X, at.Y, ink)
	for d := 1; d <= 3; d++ {
		p.lines.SetForced(at.X+d, at.Y, ink)
		p.lines.SetForced(at.X-d, at.Y, ink)
		p.lines.SetForced(at.X, at.Y+d, ink)
		p.lines.SetForced(at.X, at.Y-d, ink)
	}
}

// markDiamond is a diamond's edge: the dots a given number of steps from the
// centre, counted along the two axes.
func (p *Painter) markDiamond(at Point, radius int, ink uint8) {
	for dx := -radius; dx <= radius; dx++ {
		dy := radius - abs(dx)
		p.lines.SetForced(at.X+dx, at.Y+dy, ink)
		p.lines.SetForced(at.X+dx, at.Y-dy, ink)
	}
}

// markCircle draws a circle by the midpoint rule, as an edge or filled.
func (p *Painter) markCircle(at Point, radius int, fill bool, ink uint8) {
	x, err := radius, 1-radius
	// One step a row, and never more rows than the radius: the octant the
	// rule walks ends where the row meets the diagonal.
	for y := range radius + 1 {
		if x < y {
			return
		}
		p.circlePoints(at, x, y, fill, ink)
		if err < 0 {
			err += 2*(y+1) + 1
			continue
		}
		x--
		err += 2*(y+1-x) + 1
	}
}

// circlePoints draws one step of a circle in each of its eight octants, or
// the rows between them when it is filled.
func (p *Painter) circlePoints(at Point, x, y int, fill bool, ink uint8) {
	if fill {
		p.markRow(at, x, y, ink)
		p.markRow(at, y, x, ink)
		return
	}
	for _, d := range [8][2]int{{x, y}, {y, x}, {-x, y}, {-y, x}, {x, -y}, {y, -x}, {-x, -y}, {-y, -x}} {
		p.lines.SetForced(at.X+d[0], at.Y+d[1], ink)
	}
}

// markRow lights the two rows of a filled circle at a step: from -half to
// half, above and below the centre.
func (p *Painter) markRow(at Point, half, row int, ink uint8) {
	for dx := -half; dx <= half; dx++ {
		p.lines.SetForced(at.X+dx, at.Y+row, ink)
		p.lines.SetForced(at.X+dx, at.Y-row, ink)
	}
}

// Glyphs are the markers drawn as a character, in the order given.
func (p *Painter) Glyphs() []Label {
	if p == nil {
		return nil
	}
	return p.glyphs
}

// MarkerLabels are the markers' labels, in the order given. They are placed
// before every other name on the frame but the furniture: the host's places
// outrank alert labels and the basemap's names (L-8.9, D-60, D-80). Upstream
// places them after the map's names (P-60); the named place is the one thing
// the listener must never lose, so this departs from it.
func (p *Painter) MarkerLabels() []Label {
	if p == nil {
		return nil
	}
	return p.markerLabels
}

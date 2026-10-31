package render

import (
	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// hatchStrokes are the three strokes an area may be hatched with, from the
// closed list of characters the renderer may emit (constants, section 3).
const hatchStrokes = "╱╲╳"

// stroke is a broken line's pattern: dots on, dots off, and how thick.
type stroke struct{ on, off, thick int }

// dashStroke is the dash a line overlay is drawn with when there is no colour to
// tell it from the basemap's own lines: five dots on, four off, two dots
// thick, set from specimen 19c (D-77). dotStroke is the terminator's, at every
// depth: a dot on, two off, one thick, so it never reads as a contour, which
// is unbroken (L-4.4).
var (
	dashStroke = stroke{on: 5, off: 4, thick: 2}
	dotStroke  = stroke{on: 1, off: 2, thick: 1}
)

// hatchOf is the stroke an alert's tint is hatched with, and how far apart:
// a cell is hatched when its column and row together are a multiple of the
// stride. A graver alert is hatched more closely and in a different stroke,
// so that two areas can be ranked with no colour at all (FR-18a).
func hatchOf(ink uint8) (textsafe.Text, int) {
	switch colour.Token(ink) {
	case colour.AlertExtremeTint:
		return textsafe.Const("╳"), 2
	case colour.AlertSevereTint:
		return textsafe.Const("╳"), 3
	case colour.AlertModerateTint:
		return textsafe.Const("╲"), 4
	case colour.AlertMinorTint:
		return textsafe.Const("╱"), 5
	case colour.AlertUnknownTint:
		return textsafe.Const("╱"), 6
	}
	return textsafe.Text{}, 0
}

// hatch fills an alert area's empty cells with its stroke. It runs last of
// all, over the cells nothing else has taken: a marker, a name and the
// basemap's own line work all keep their cells, and the hatch fills what is
// left (FR-18a, specimen 13a).
func (r *Renderer) hatch() {
	g := r.grid
	for row := range g.rows {
		for col := range g.cols {
			c := &g.cells[row*g.cols+col]
			stroke, stride := hatchOf(c.area)
			if stride == 0 || c.taken || c.glyph != blank || (col+row)%stride != 0 {
				continue
			}
			g.write(col, row, stroke, c.area)
		}
	}
}

// dash draws a line as a dashed one, the dash carrying on from one part of
// the line to the next so that a bend does not restart it.
func (p *Painter) dash(ring []Point, ink uint8, with stroke) {
	run := 0
	for i := 0; i+1 < len(ring); i++ {
		run = p.dashSegment(ring[i], ring[i+1], run, ink, with)
	}
}

// dashSegment dashes one straight part, from where the dash had got to, and
// answers with where it has got to now.
func (p *Painter) dashSegment(a, b Point, run int, ink uint8, with stroke) int {
	steps := max(abs(b.X-a.X), abs(b.Y-a.Y))
	if steps == 0 {
		return run
	}
	for from := 0; from < steps; {
		phase := run % (with.on + with.off)
		length := min(with.on+with.off-phase, steps-from)
		if phase < with.on {
			length = min(with.on-phase, steps-from)
			one, two := along(a, b, from, steps), along(a, b, from+length, steps)
			p.lines.Line(one.X, one.Y, two.X, two.Y, with.thick, ink)
		}
		from += length
		run += length
	}
	return run
}

// along is the point a given number of steps along a part of a line.
func along(a, b Point, at, steps int) Point {
	return Point{X: a.X + (b.X-a.X)*at/steps, Y: a.Y + (b.Y-a.Y)*at/steps}
}

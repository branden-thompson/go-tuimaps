package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// TestEveryCellGetsItsOwnColoursThroughTheMemo is watchpost D-209: a frame's
// colours are worked out once a key, so two cells that differ in anything a
// colour is worked out from - its area, its field's ink, its own ink, the
// night, a zone, a faint band, text's contrast - each get the colours worked
// out for it, never the other's.
func TestEveryCellGetsItsOwnColoursThroughTheMemo(t *testing.T) {
	v := gulfView()
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), OverlaysVersion: 1, Depth: colour.Truecolor}
	r, _ := drawn(t, in)
	g := ground{kind: in.Ground.Kind(in.Palette)}
	g.colour, g.painted = in.Ground.InEffect(in.Palette)
	ground := cell{glyph: ' ', ink: uint8(colour.LabelRegion)}
	field := cell{glyph: ' ', ink: uint8(colour.LabelRegion), under: uint8(colour.Temperature1)}
	differ := 0
	for name, c := range map[string][2]cell{
		"the night":       {ground, func() cell { c := ground; c.night = true; return c }()},
		"a zone":          {ground, func() cell { c := ground; c.zone = true; return c }()},
		"an alert's area": {ground, func() cell { c := ground; c.area = uint8(colour.AlertSevereTint); return c }()},
		"a field's ink":   {field, func() cell { c := field; c.under = uint8(colour.Temperature1) + 5; return c }()},
		"a faint band":    {field, func() cell { c := field; c.faint = true; return c }()},
		"text's contrast": {field, func() cell { c := field; c.strict = true; return c }()},
		"its own ink":     {ground, func() cell { c := ground; c.ink = uint8(colour.AlertSevereOutline); return c }()},
	} {
		r.memo = map[colourKey][2]colour.RGB{}
		first, second := &pen{fresh: true}, &pen{fresh: true}
		r.colourCell(c[0], in, g, first)
		r.colourCell(c[1], in, g, second)
		fg, bg := r.colours(c[1], in, g.colour, g.kind)
		if second.fg != fg || second.bg != bg {
			t.Errorf("%s: the second cell drew %v on %v; its own colours are %v on %v", name, second.fg, second.bg, fg, bg)
		}
		if first.fg != second.fg || first.bg != second.bg {
			differ++
		}
	}
	if differ < 4 {
		t.Errorf("only %d of the pairs draw differently: the check measures little", differ)
	}
}

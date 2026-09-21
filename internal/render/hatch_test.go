package render

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// alertArea is one alert polygon covering the middle of the view.
func alertArea(t *testing.T, role colour.Token, label string) scene.Shape {
	t.Helper()
	return scene.Shape{Kind: scene.ShapeArea, Role: uint8(role), Label: label,
		Rings: [][]scene.Vertex{lonLatBox(t, -86, 24, -82, 28)}}
}

// hatched is how many cells of a frame carry a hatch stroke.
func hatched(f Frame) map[rune]int {
	out := map[rune]int{}
	for _, line := range f.Lines {
		for _, r := range plain(line) {
			if strings.ContainsRune(hatchStrokes, r) {
				out[r]++
			}
		}
	}
	return out
}

// TestHatchAndLabelNoColour is plan task 09.15 (FR-18a): with no colour an
// alert area is hatched and carries its plain-word label, so that it is not
// another braille shape among the basemap's; the hatch is denser and its
// stroke different for a graver severity, so two areas can be ranked.
func TestHatchAndLabelNoColour(t *testing.T) {
	v, in := gulf(t)
	in.Depth = NoColour
	in.Shapes = []scene.Shape{alertArea(t, colour.AlertSevereOutline, "Tornado Warning")}
	in.OverlaysVersion = 1
	f := render(t, in)
	got := hatched(f)
	if len(got) == 0 {
		t.Fatalf("an alert area with no colour is drawn with no hatch at all:\n%s", strings.Join(f.Lines, "\n"))
	}
	if !strings.Contains(plain(strings.Join(f.Lines, "\n")), "Tornado Warning") {
		t.Error("the alert's plain-word label is not on the frame")
	}
	// With colour there is no hatch: the tint carries it.
	coloured := in
	coloured.Depth = Truecolor
	if len(hatched(render(t, coloured))) != 0 {
		t.Error("an alert area is hatched even with colour to draw it in")
	}
	// A graver severity is hatched more heavily, and in its own stroke.
	counts := map[colour.Token]int{}
	strokes := map[colour.Token]rune{}
	for _, role := range []colour.Token{colour.AlertExtremeOutline, colour.AlertModerateOutline, colour.AlertMinorOutline} {
		one := in
		one.Shapes = []scene.Shape{alertArea(t, role, "Warning")}
		one.OverlaysVersion++
		for stroke, n := range hatched(render(t, one)) {
			counts[role] += n
			strokes[role] = stroke
		}
	}
	if counts[colour.AlertExtremeOutline] <= counts[colour.AlertModerateOutline] ||
		counts[colour.AlertModerateOutline] <= counts[colour.AlertMinorOutline] {
		t.Errorf("hatch by severity: extreme %d, moderate %d, minor %d; a graver one is drawn more heavily",
			counts[colour.AlertExtremeOutline], counts[colour.AlertModerateOutline], counts[colour.AlertMinorOutline])
	}
	if strokes[colour.AlertExtremeOutline] == strokes[colour.AlertMinorOutline] {
		t.Error("the gravest and the slightest alert are hatched with the same stroke")
	}
	_ = v
}

// TestDashedLineNoColour is the other half of 09.15 (FR-18a, D-77, specimen
// 19c): with no colour an overlay's line is dashed - five dots on, four off,
// two dots thick - so that it is not the unbroken river beside it.
func TestDashedLineNoColour(t *testing.T) {
	v, _ := gulf(t)
	track := scene.Shape{Kind: scene.ShapeLine, Role: uint8(colour.Track),
		Rings: [][]scene.Vertex{{vertex(t, -92, 26), vertex(t, -78, 28)}}}
	solid, dashed := painter(t, v), painter(t, v)
	dashed.SetDepth(NoColour)
	for _, p := range []*Painter{solid, dashed} {
		if err := p.Shape(v, track); err != nil {
			t.Fatal(err)
		}
	}
	if lit(dashed) == 0 {
		t.Fatal("a dashed line drew nothing")
	}
	if gaps(solid) != 0 {
		t.Fatalf("the line drawn with colour has %d gaps of its own", gaps(solid))
	}
	if gaps(dashed) < 4 {
		t.Errorf("a line with no colour has %d gaps along it; five dots on and four off leaves many", gaps(dashed))
	}
	if dashOn != 5 || dashOff != 4 || dashThick != 2 {
		t.Errorf("the dash is %d dots on, %d off, %d thick; the constants document fixes five, four and two", dashOn, dashOff, dashThick)
	}
	// An area's outline is not dashed: it is closed, and its hatch tells it
	// from the basemap.
	area, plainArea := painter(t, v), painter(t, v)
	area.SetDepth(NoColour)
	shape := alertArea(t, colour.AlertSevereOutline, "")
	for _, p := range []*Painter{area, plainArea} {
		if err := p.Shape(v, shape); err != nil {
			t.Fatal(err)
		}
	}
	if lit(area) != lit(plainArea) {
		t.Errorf("an alert's outline is %d dots with no colour and %d with it; an outline is not dashed", lit(area), lit(plainArea))
	}
}

// TestMarkerOverHatch is plan task 09.12 (FR-18a, specimen 13a): a marker
// inside a hatched area is drawn over the hatch, never under it. The defect
// this guards against is the hatch taking the cell first.
func TestMarkerOverHatch(t *testing.T) {
	v, in := gulf(t)
	in.Depth = NoColour
	in.Shapes = []scene.Shape{alertArea(t, colour.AlertExtremeOutline, "Warning")}
	in.OverlaysVersion = 1
	in.Labels = false
	at, _, row := quarter(t, v)
	// The marker sits inside the area, and is a character so that it and the
	// hatch want the same cell.
	inside, err := v.FromDot(float64(v.Cols), float64(v.Rows))
	if err != nil {
		t.Fatal(err)
	}
	_ = at
	in.Markers = []Marker{{At: inside, Shape: MarkerGlyph, Text: "◉", Ink: uint8(colour.Marker)}}
	line := plain(render(t, in).Lines[row])
	if !strings.Contains(line, "◉") {
		t.Errorf("the marker is not in its row; the hatch took its cell:\n%s", line)
	}
}

// gaps is how many runs of empty columns a painter's line work has between
// its leftmost and its rightmost lit dot: none for an unbroken line.
func gaps(p *Painter) int {
	w, h := p.Lines().Dots()
	var cols []bool
	for x := range w {
		lit := false
		for y := range h {
			lit = lit || p.Lines().Lit(x, y)
		}
		cols = append(cols, lit)
	}
	first, last := -1, -1
	for x, on := range cols {
		if on && first < 0 {
			first = x
		}
		if on {
			last = x
		}
	}
	n, inGap := 0, false
	for x := first; x >= 0 && x <= last; x++ {
		if !cols[x] && !inGap {
			n++
		}
		inGap = !cols[x]
	}
	return n
}

package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestTheTerminatorIsMarkedWithoutColour is L-4.4: the terminator is dotted
// at every depth - colour or none - so it never reads as a contour, which is
// unbroken, nor as a track, whose dash is longer.
func TestTheTerminatorIsMarkedWithoutColour(t *testing.T) {
	v, _ := gulf(t)
	line := scene.Shape{Kind: scene.ShapeLine, Role: uint8(colour.Terminator),
		Rings: [][]scene.Vertex{{vertex(t, -92, 26), vertex(t, -78, 28)}}}
	for _, d := range []Depth{colour.Truecolor, NoColour} {
		p := painter(t, v)
		p.SetDepth(d)
		if err := p.Shape(v, line); err != nil {
			t.Fatal(err)
		}
		if lit(p) == 0 || gaps(p) < 8 {
			t.Errorf("depth %v: the terminator lit %d dots with %d gaps; a dot on and two off leaves a gap at every third", d, lit(p), gaps(p))
		}
	}
}

// TestTheNightTintCarriesNoAlertWords is L-4.3: the night side fills its own
// canvas - the areas the water test reads are left as they were - and keeps
// no words, no digits and no outline.
func TestTheNightTintCarriesNoAlertWords(t *testing.T) {
	v, _ := gulf(t)
	night := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.Night), Label: "NIGHT", Word: "NIGHT", Mark: "4",
		Rings: [][]scene.Vertex{{vertex(t, -95, 22), vertex(t, -80, 22), vertex(t, -80, 30), vertex(t, -95, 30), vertex(t, -95, 22)}}}
	p := painter(t, v)
	if err := p.Shape(v, night); err != nil {
		t.Fatal(err)
	}
	inNight, areas := 0, 0
	cols, rows := p.areas.cols, p.areas.rows
	for row := range rows {
		for col := range cols {
			if p.InNight(col, row) {
				inNight++
			}
			if _, owned := p.Area(col, row); owned {
				areas++
			}
		}
	}
	if inNight == 0 {
		t.Fatal("the night side owns no cell")
	}
	if areas != 0 {
		t.Errorf("the night filled %d cells of the areas the water test reads", areas)
	}
	if len(p.OverlayLabels()) != 0 || len(p.outlines) != 0 || lit(p) != 0 {
		t.Errorf("the night kept %d labels, %d outlines and %d dots of line", len(p.OverlayLabels()), len(p.outlines), lit(p))
	}
}

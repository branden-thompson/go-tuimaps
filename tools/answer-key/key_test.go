package main

import (
	"math"
	"os"
	"strings"
	"testing"
)

// TestKeyModuleRequiresNothing is plan task 11.1 (D-43): this tool is the
// independent half of M1b, so its module file must require nothing at all.
// It cannot import the library even by accident, which is the whole point:
// two implementations that agree are evidence, and one implementation
// checked against itself is not.
func TestKeyModuleRequiresNothing(t *testing.T) {
	body, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "require") {
		t.Errorf("the module file requires something:\n%s", text)
	}
	if strings.Contains(text, "replace") {
		t.Errorf("the module file replaces something:\n%s", text)
	}
	// And no source file of this tool names the library.
	names, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if !strings.HasSuffix(name.Name(), ".go") || strings.HasSuffix(name.Name(), "_test.go") {
			continue // this file names the library in order to forbid it
		}
		source, err := os.ReadFile(name.Name())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(source), "go-tuimaps/internal") {
			t.Errorf("%s reaches into the library", name.Name())
		}
	}
}

// square is a scenario shape: a square from 10 to 20 east, 10 to 20 north.
func square() Shape {
	return Shape{ID: "square", Kind: "area", Rings: [][][]float64{{
		{10, 10}, {20, 10}, {20, 20}, {10, 20},
	}}}
}

// triangle is a right-angled triangle with its corner at 0,0.
func triangle() Shape {
	return Shape{ID: "triangle", Kind: "area", Rings: [][][]float64{{
		{0, 0}, {4, 0}, {0, 3},
	}}}
}

// TestKeyOnHandMadeShapes is the rest of plan task 11.1: shapes whose
// answers can be worked out by hand, so that the key itself is checked
// before anything is checked against it.
func TestKeyOnHandMadeShapes(t *testing.T) {
	// The middle of the square is inside it. The square runs from 10 to 20
	// each way, so the middle is five degrees from every edge: five degrees
	// of longitude at 15 north is 5 * 111.32 * cos(15) = 537.7 km, which is
	// nearer than five degrees of latitude at 555.8 km.
	middle := Place{Name: "middle", Lon: 15, Lat: 15}
	got := answer(middle, square(), nil)
	if got.Inside == nil || !*got.Inside {
		t.Errorf("the middle of a square is %+v", got.Inside)
	}
	if math.Abs(got.Km-537.7) > 4 {
		t.Errorf("the nearest edge is %v km, want about 538: five degrees of longitude at 15 north", got.Km)
	}

	// A place two degrees west of the square's west edge, at its middle
	// latitude: the edge is due east, and two degrees of longitude at 15
	// north is 2 * 111.32 * cos(15) = 215.1 km.
	west := Place{Name: "west", Lon: 8, Lat: 15}
	got = answer(west, square(), nil)
	if got.Inside == nil || *got.Inside {
		t.Error("a place west of the square is inside it")
	}
	if math.Abs(got.Km-215.1) > 3 {
		t.Errorf("it is %v km from the edge, want about 215", got.Km)
	}
	if got.Compass != "east" {
		t.Errorf("the edge is to the %q", got.Compass)
	}
	if math.Abs(got.Bearing-90) > 1 {
		t.Errorf("the bearing is %v", got.Bearing)
	}

	// The triangle: a place inside it, and one outside past the hypotenuse.
	if in := answer(Place{Name: "in", Lon: 0.5, Lat: 0.5}, triangle(), nil); in.Inside == nil || !*in.Inside {
		t.Errorf("a place inside the triangle is %+v", in.Inside)
	}
	if out := answer(Place{Name: "out", Lon: 3, Lat: 2}, triangle(), nil); out.Inside == nil || *out.Inside {
		t.Errorf("a place past the hypotenuse is %+v", out.Inside)
	}
	// A point shape: the distance is to the position itself.
	point := Shape{ID: "strike", Kind: "point", Rings: [][][]float64{{{15, 16}}}}
	if got := answer(middle, point, nil); math.Abs(got.Km-111.2) > 2 || got.Compass != "north" {
		t.Errorf("a point one degree north is %v km to the %q", got.Km, got.Compass)
	}
	// A line shape is not closed: the way back from its end is not an edge.
	line := Shape{ID: "track", Kind: "line", Rings: [][][]float64{{{10, 16}, {20, 16}}}}
	if got := answer(middle, line, nil); math.Abs(got.Km-111.2) > 2 || got.Compass != "north" {
		t.Errorf("a line one degree north is %v km to the %q", got.Km, got.Compass)
	}
}

// TestKeyNamesDimensionByBearing is plan task 11.2 (S17-2): a cell is twice
// as tall as it is wide in ground terms, so "under one cell" means the row
// for something north or south and the column for something east or west.
func TestKeyNamesDimensionByBearing(t *testing.T) {
	view := View{Name: "149x38", CellKm: 1.6, CellRowKm: 3.2}
	// Two km due east is more than a column and less than a row.
	if got := cellsAcross(2, 90, view); math.Abs(got-1.25) > 0.01 {
		t.Errorf("2 km east is %v cells, want it measured by the column's 1.6 km", got)
	}
	if got := cellsAcross(2, 0, view); math.Abs(got-0.625) > 0.01 {
		t.Errorf("2 km north is %v cells, want it measured by the row's 3.2 km", got)
	}
	// And the frame's answer follows: east it is outside, north it is on the
	// edge, from the same two kilometres.
	no := false
	if frameAnswer(cellsAcross(2, 90, view), &no) != "outside" {
		t.Error("2 km east is not outside")
	}
	if frameAnswer(cellsAcross(2, 0, view), &no) != "on the edge" {
		t.Error("2 km north is not on the edge")
	}
	// A view that says nothing about its rows falls back to its columns.
	plain := View{Name: "plain", CellKm: 1.6}
	if got := cellsAcross(2, 0, plain); math.Abs(got-1.25) > 0.01 {
		t.Errorf("with no row height given, 2 km north is %v cells", got)
	}
}

// TestKeyWholeScenario: the key for a scenario is every place against every
// shape, with the cells and the frame's answer at each view.
func TestKeyWholeScenario(t *testing.T) {
	scenario := Scenario{
		Name:   "hand-made",
		Places: []Place{{Name: "middle", Lon: 15, Lat: 15}},
		Shapes: []Shape{square(), {ID: "strike", Kind: "point", Rings: [][][]float64{{{15, 15.01}}}}},
		Views:  []View{{Name: "149x38", CellKm: 1.6, CellRowKm: 3.2}},
	}
	key := Key(scenario)
	if len(key) != 2 {
		t.Fatalf("%d answers for one place and two shapes", len(key))
	}
	for _, a := range key {
		if a.Place != "middle" || a.Compass == "" {
			t.Errorf("%+v", a)
		}
		if len(a.Cells) != 1 || len(a.Frame) != 1 {
			t.Errorf("%+v has %d cell counts and %d frame answers", a, len(a.Cells), len(a.Frame))
		}
	}
	// The strike is about a kilometre north: under one cell's height, so the
	// frame cannot settle it.
	if key[1].Frame["149x38"] != "on the edge" {
		t.Errorf("a strike 1.1 km north reads %q at 149 by 38", key[1].Frame["149x38"])
	}
}

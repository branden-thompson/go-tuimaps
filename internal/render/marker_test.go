package render

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// centre is the place the marker tests draw at: the middle of the tile the
// painter tests use.
func centre() project.LonLat {
	at, err := project.FromTile(float64(here.X)+0.5, float64(here.Y)+0.5, here.Z)
	if err != nil {
		panic(err)
	}
	return at
}

// dotsOf draws one marker and answers with the dots it lit, as offsets from
// the marker's own dot.
func dotsOf(t *testing.T, m Marker) map[[2]int]bool {
	t.Helper()
	v := worldView()
	p := painter(t, v)
	if err := p.Mark(v, m, true); err != nil {
		t.Fatal(err)
	}
	w, h := p.Lines().Dots()
	at := [2]int{w / 2, h / 2}
	out := map[[2]int]bool{}
	for y := range h {
		for x := range w {
			if p.Lines().Lit(x, y) {
				out[[2]int{x - at[0], y - at[1]}] = true
			}
		}
	}
	return out
}

// TestParityP58_MarkerShapes is plan task 09.29: upstream's six marker
// shapes, each the size upstream draws it.
func TestParityP58_MarkerShapes(t *testing.T) {
	at := centre()
	cases := []struct {
		name   string
		m      Marker
		count  int
		holds  [][2]int
		misses [][2]int
	}{
		{"a dot is three by three", Marker{At: at, Shape: MarkerDot}, 9,
			[][2]int{{0, 0}, {-1, -1}, {1, 1}}, [][2]int{{2, 0}, {0, -2}}},
		{"a cross reaches three each way", Marker{At: at, Shape: MarkerCross}, 13,
			[][2]int{{0, 0}, {3, 0}, {-3, 0}, {0, 3}, {0, -3}}, [][2]int{{4, 0}, {1, 1}, {-2, 2}}},
		{"a diamond has radius three", Marker{At: at, Shape: MarkerDiamond}, 12,
			[][2]int{{3, 0}, {0, 3}, {-3, 0}, {0, -3}, {1, 2}, {-2, -1}}, [][2]int{{0, 0}, {1, 1}, {4, 0}}},
		{"a filled circle of radius three", Marker{At: at, Shape: MarkerDisc, Radius: 3}, 37,
			[][2]int{{0, 0}, {3, 0}, {2, 2}, {-1, -2}, {3, 1}}, [][2]int{{3, 3}, {4, 0}}},
	}
	for _, c := range cases {
		got := dotsOf(t, c.m)
		if len(got) != c.count {
			t.Errorf("%s: %d dots, want %d", c.name, len(got), c.count)
		}
		for _, d := range c.holds {
			if !got[d] {
				t.Errorf("%s: no dot at %v", c.name, d)
			}
		}
		for _, d := range c.misses {
			if got[d] {
				t.Errorf("%s: a dot at %v, which is outside the shape", c.name, d)
			}
		}
	}
	// A ring is the midpoint circle of its radius: every dot about that far
	// out, and one at each of the four points of the compass.
	ring := dotsOf(t, Marker{At: at, Shape: MarkerRing, Radius: 4})
	if ring[[2]int{0, 0}] {
		t.Error("a ring is drawn filled")
	}
	for d := range ring {
		if r2 := d[0]*d[0] + d[1]*d[1]; r2 < 9 || r2 > 25 {
			t.Errorf("a ring of radius 4 has a dot at %v", d)
		}
	}
	for _, d := range [][2]int{{4, 0}, {-4, 0}, {0, 4}, {0, -4}} {
		if !ring[d] {
			t.Errorf("a ring of radius 4 is open at %v", d)
		}
	}
	// A filled circle is the same circle filled: it holds every dot of the ring.
	disc := dotsOf(t, Marker{At: at, Shape: MarkerDisc, Radius: 4})
	for d := range ring {
		if !disc[d] {
			t.Errorf("a filled circle of radius 4 is missing the ring's dot at %v", d)
		}
	}
}

// gulf is the view the frame-level marker tests draw, with the embedded
// tiles behind it: a marker on an empty frame would meet the no-tiles notice.
func gulf(t *testing.T) (project.View, Input) {
	t.Helper()
	v, err := project.WholeWorld(149, 38)
	if err != nil {
		t.Fatal(err)
	}
	v.Zoom, v.Centre = 3, project.LonLat{Lon: -84, Lat: 26}
	return v, input(t, v)
}

// quarter is the place a quarter of the way down the view, with the cell its
// marker's label begins in - four dots to its right (P-60) - and its row.
func quarter(t *testing.T, v project.View) (project.LonLat, int, int) {
	t.Helper()
	x, y := float64(v.Cols), float64(v.Rows) // in dots: the middle across, a quarter down
	at, err := v.FromDot(x, y)
	if err != nil {
		t.Fatal(err)
	}
	return at, (int(x) + labelGap) / 2, int(y) / 4
}

// TestMarkerGlyphIsACell: a marker of a character is drawn as text, which
// wins the cell over any dots in it (P-10).
func TestMarkerGlyphIsACell(t *testing.T) {
	v, in := gulf(t)
	at, _, row := quarter(t, v)
	glyph := "\u25c9"
	in.Markers = []Marker{{At: at, Shape: MarkerGlyph, Text: glyph, Ink: uint8(colour.Marker)}}
	f := render(t, in)
	if !strings.Contains(plain(f.Lines[row]), glyph) {
		t.Errorf("the marker's own character is not in its row:\n%s", plain(f.Lines[row]))
	}
}

// TestParityP60_MarkerCullAndLabel is the rest of 09.29: a marker more than
// twenty dots outside the rectangle is not drawn, its label sits four dots
// to its right, and the label is in the marker's own colour.
func TestParityP60_MarkerCullAndLabel(t *testing.T) {
	v := worldView()
	outside := centre()
	outside.Lon += 180 // a world away: far outside this tile's view
	p := painter(t, v)
	if err := p.Mark(v, Marker{At: outside, Shape: MarkerDot}, true); err != nil {
		t.Fatal(err)
	}
	if lit(p) != 0 {
		t.Errorf("%d dots drawn for a marker outside the view", lit(p))
	}
	if p.Culled() != 1 {
		t.Errorf("a marker outside the view was not counted as culled: %d", p.Culled())
	}
	// Up to twenty dots past the edge a marker is kept, because one big
	// enough still reaches into the rectangle; beyond that it is culled.
	w, h := p.Lines().Dots()
	for _, c := range []struct {
		dots int
		kept bool
	}{{0, true}, {MarkerPad - 2, true}, {MarkerPad + 8, false}} {
		p := painter(t, v)
		at, err := v.FromDot(float64(w+c.dots), float64(h/2))
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Mark(v, Marker{At: at, Shape: MarkerRing, Radius: 24}, true); err != nil {
			t.Fatal(err)
		}
		if kept := p.Culled() == 0; kept != c.kept {
			t.Errorf("a marker %d dots past the edge kept %v, want %v", c.dots, kept, c.kept)
		}
		if drew := lit(p) > 0; drew != c.kept {
			t.Errorf("a marker %d dots past the edge drew %d dots", c.dots, lit(p))
		}
	}
	view, in := gulf(t)
	at, wantCol, row := quarter(t, view)
	in.Labels = false // the map's own names are the next test's; here the cells are clear
	in.Markers = []Marker{{At: at, Shape: MarkerDot, Label: "Base", Ink: uint8(colour.Marker)}}
	f := render(t, in)
	line := plain(f.Lines[row])
	cut := strings.Index(line, "Base")
	if cut < 0 {
		t.Fatalf("the marker's label is not in its row:\n%s", line)
	}
	col := utf8.RuneCountInString(line[:cut]) // cells, not bytes: braille takes three each
	if col != wantCol {
		t.Errorf("the label starts at cell %d, want %d: four dots right of the marker", col, wantCol)
	}
	other := in
	other.Markers = []Marker{{At: at, Shape: MarkerDot, Label: "Base", Ink: uint8(colour.River)}}
	other.OverlaysVersion = 1
	if colours(f.Lines[row], "Base") == colours(render(t, other).Lines[row], "Base") {
		t.Error("the label is drawn in the same colour whatever the marker's own")
	}
}

// TestMarkerLabelAfterTheMapsNames is P-60's other half: a marker's label is
// collision-checked after the map's own names, so a name already placed
// keeps its cells.
func TestMarkerLabelAfterTheMapsNames(t *testing.T) {
	v, err := project.WholeWorld(149, 38)
	if err != nil {
		t.Fatal(err)
	}
	v.Zoom, v.Centre = 3, project.LonLat{Lon: -84, Lat: 26}
	in := input(t, v)
	before := plain(strings.Join(render(t, in).Lines, "\n"))
	if !strings.Contains(before, "Miami") {
		t.Fatal("the frame this test builds on no longer places Miami")
	}
	in.Markers = []Marker{{At: project.LonLat{Lon: -80.19, Lat: 25.77}, Shape: MarkerDot, Label: "Miami", Ink: uint8(colour.Marker)}}
	in.OverlaysVersion = 1
	after := plain(strings.Join(render(t, in).Lines, "\n"))
	if strings.Count(after, "Miami") != 1 {
		t.Errorf("the marker's label was placed over the name already there: %d of them", strings.Count(after, "Miami"))
	}
}

// TestBlinkingMarkerFollowsThePhase is plan task 09.13 in the frame: a
// blinking marker is drawn in one half of the blink and not in the other,
// while a steady marker is drawn in both.
func TestBlinkingMarkerFollowsThePhase(t *testing.T) {
	v, both := gulf(t)
	at, _, row := quarter(t, v)
	both.Markers = []Marker{{At: at, Shape: MarkerDot, Blink: true}}
	both.MarkerPhase = true
	on := render(t, both)
	both.MarkerPhase = false
	off := render(t, both)
	if plain(on.Lines[row]) == plain(off.Lines[row]) {
		t.Error("a blinking marker is drawn the same in both halves of its blink")
	}
	steady := both
	steady.Markers = []Marker{{At: at, Shape: MarkerDot}}
	steady.MarkerPhase = true
	shown := render(t, steady)
	steady.MarkerPhase = false
	again := render(t, steady)
	if plain(shown.Lines[row]) != plain(again.Lines[row]) {
		t.Error("a marker that does not blink changed with the phase")
	}
}

// colours is the colour sequence in force where a piece of text begins.
func colours(line, text string) string {
	cut := strings.Index(plain(line), text)
	if cut < 0 {
		return ""
	}
	cut = utf8.RuneCountInString(plain(line)[:cut])
	seen, col, i := "", 0, 0
	for i < len(line) {
		if loc := sgr.FindStringIndex(line[i:]); loc != nil && loc[0] == 0 {
			if col > cut {
				break
			}
			seen, i = line[i:i+loc[1]], i+loc[1]
			continue
		}
		_, size := utf8.DecodeRuneInString(line[i:])
		i += size
		col++
	}
	return seen
}

// TestParityP35_PoiGlyph is the parity row of the same name: a symbol with
// no name of its own draws upstream's place glyph, and one with a name
// draws the name.
func TestParityP35_PoiGlyph(t *testing.T) {
	v := worldView()
	g, err := newGrid(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if !placed(g, Label{X: 40, Y: 8, Ink: uint8(colour.LabelPlace)}) {
		t.Fatal("a symbol with no name was not placed")
	}
	if got := g.cells[2*v.Cols+20].text; got != placeGlyph {
		t.Errorf("a symbol with no name drew %q, want upstream's %q", got, placeGlyph)
	}
	named, err := newGrid(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if !placed(named, Label{X: 40, Y: 8, Name: textsafe.Clean("Quay"), Ink: uint8(colour.LabelPlace)}) {
		t.Fatal("a symbol with a name was not placed")
	}
	if got := named.cells[2*named.cols+18].text; got != "Q" { // centred on its point
		t.Errorf("a symbol with a name drew %q; the name is what is drawn", got)
	}
}

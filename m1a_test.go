package tuimaps_test

import (
	"context"
	"path/filepath"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// areaScenarios are the M1 scenarios that ask "inside or outside", which is
// the question a frame can be read for mechanically.
func areaScenarios() []string { return []string{"scenario-1", "scenario-2"} }

// TestM1aFrameDoesNotContradictKey is plan task 14.4, and the machine's
// half of M1a. **It reads the picture** - not the geometry the picture was
// drawn from - and asks one question of it: does the frame put the place on
// the side of the boundary the independent key puts it on?
//
// Where the key puts the place closer to the edge than one cell, the frame
// cannot settle the question and is not asked to: that is D-67, and it is
// why the description exists. HUM LEAD's own reading is 14.15; this is the
// guard that stops the picture drifting away from the key unnoticed.
func TestM1aFrameDoesNotContradictKey(t *testing.T) {
	for _, name := range areaScenarios() {
		for _, size := range m1Sizes() {
			t.Run(name+"/"+sizeName(size), func(t *testing.T) {
				judgeFrame(t, name, size)
			})
		}
	}
}

// sizeName is a size written the way the frames are named.
func sizeName(size tuimaps.Size) string {
	return itoa(size.Cols) + "x" + itoa(size.Rows)
}

// itoa is a small number written out.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// judgeFrame reads one scenario's frame at one size and holds it to the key.
func judgeFrame(t *testing.T, name string, size tuimaps.Size) {
	t.Helper()
	s := scenarioNamed(t, name)
	var key wholeKey
	read(t, filepath.Join("06_docs", "02_features", "go-tuimaps", "02-analysis", "scenarios", name+"-key.json"), &key)

	m := world(t, size.Cols, size.Rows)
	m.ColourDepth(tuimaps.NoColour) // one rune a cell, which is what makes a frame readable
	places := make([]tuimaps.Place, 0, len(s.Places))
	at := make([]tuimaps.LonLat, 0, len(s.Places))
	for _, p := range s.Places {
		places = append(places, tuimaps.Place{Name: p.Name, At: tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat}})
		at = append(at, tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat})
	}
	if _, err := m.SetPlaces(places); err != nil {
		t.Fatal(err)
	}
	putScenario(t, m, s)
	if err := m.FitTo(at, m.Overlays(), 2); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	with := cellsOf(t, m, size)

	// The same view with nothing on it: what the two frames differ in is
	// exactly what the overlay drew, read from the picture itself.
	ids := m.Overlays()
	for _, id := range ids {
		if _, err := m.Remove(id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	without := cellsOf(t, m, size)

	centre, zoom := m.Centre()
	// The public LonLat is the projection's own type under another name, so
	// nothing is converted here.
	view := project.View{Centre: centre, Zoom: zoom, Cols: size.Cols, Rows: size.Rows}
	perCol, _, err := view.CellSpanKm()
	if err != nil {
		t.Fatal(err)
	}
	judged := 0
	for _, answer := range key.Shapes {
		if answer.Inside == nil {
			continue
		}
		if answer.Km < perCol {
			// The picture cannot settle this and is not asked to (D-67).
			t.Logf("%s in %s: %.2f km from the edge, under one cell of %.2f km - the words settle this one",
				answer.Place, answer.Shape, answer.Km, perCol)
			continue
		}
		place, ok := placeNamed(places, answer.Place)
		if !ok {
			t.Fatalf("the key names the place %q, which the scenario does not", answer.Place)
		}
		col, row, ok := cellOf(view, place.At)
		if !ok {
			t.Errorf("%s is not on the frame at all", answer.Place)
			continue
		}
		inside := enclosed(with, without, col, row)
		if inside != *answer.Inside {
			t.Errorf("%s in %s: the frame reads %s and the key says %s",
				answer.Place, answer.Shape, saidInside(inside), saidInside(*answer.Inside))
		}
		judged++
	}
	// Scenario 1 puts one place inside its polygon and one well outside it,
	// at both sizes: a run that judged neither would be a test that had
	// quietly stopped reading anything, and would pass for ever.
	if name == "scenario-1" && judged < 2 {
		t.Errorf("only %d of this scenario's answers were read from the frame; both should be", judged)
	}
	if judged == 0 {
		t.Logf("nothing here could be judged from the frame at this size; every answer is under one cell")
	}
}

// saidInside is a verdict in words.
func saidInside(inside bool) string {
	if inside {
		return "inside"
	}
	return "outside"
}

// placeNamed is one of the scenario's places by name.
func placeNamed(places []tuimaps.Place, name string) (tuimaps.Place, bool) {
	for _, p := range places {
		if p.Name == name {
			return p, true
		}
	}
	return tuimaps.Place{}, false
}

// cellsOf is a frame as rows of cells, one rune each, which is what a frame
// with no colour in it is.
func cellsOf(t *testing.T, m *tuimaps.Map, size tuimaps.Size) [][]rune {
	t.Helper()
	frame, err := m.Render(size, noon)
	if err != nil {
		t.Fatal(err)
	}
	rows := make([][]rune, 0, len(frame.Lines))
	for _, line := range frame.Lines {
		rows = append(rows, []rune(line))
	}
	return rows
}

// cellOf is the cell a position falls in, or false where it is off the
// frame.
func cellOf(view project.View, at tuimaps.LonLat) (int, int, bool) {
	x, y, err := view.ToDot(at)
	if err != nil {
		return 0, 0, false
	}
	col, row := int(x)/2, int(y)/4
	if col < 0 || row < 0 || col >= view.Cols || row >= view.Rows {
		return 0, 0, false
	}
	return col, row, true
}

// enclosed reads the picture: starting at the place's own cell and stepping
// only through cells the overlay did not draw in, can the edge of the frame
// be reached? If it cannot, the overlay's drawing surrounds the place, and
// the frame says "inside".
//
// The overlay's cells are the ones the two frames differ in - the outline,
// the hatch and the label - which is read from the pictures themselves and
// never from the geometry they were drawn from.
func enclosed(with, without [][]rune, col, row int) bool {
	drawnBy := func(c, r int) bool {
		if r < 0 || r >= len(with) || r >= len(without) {
			return false
		}
		if c < 0 || c >= len(with[r]) || c >= len(without[r]) {
			return false
		}
		return with[r][c] != without[r][c]
	}
	seen := map[[2]int]bool{}
	stack := [][2]int{{col, row}}
	for len(stack) > 0 {
		at := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[at] {
			continue
		}
		seen[at] = true
		c, r := at[0], at[1]
		if r < 0 || r >= len(with) || c < 0 || c >= len(with[r]) {
			return false // the edge of the frame was reached: not enclosed
		}
		if drawnBy(c, r) && at != [2]int{col, row} {
			continue // the overlay's own cells are the wall
		}
		stack = append(stack, [2]int{c - 1, r}, [2]int{c + 1, r}, [2]int{c, r - 1}, [2]int{c, r + 1})
	}
	return true
}

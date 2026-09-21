package render

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

func canvas(t *testing.T, cols, rows int) *Canvas {
	t.Helper()
	c, err := NewCanvas(cols, rows)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// picture draws the canvas's dots, one character a dot, for golden tests.
func picture(c *Canvas) string {
	var b strings.Builder
	w, h := c.Dots()
	for y := range h {
		for x := range w {
			if c.Lit(x, y) {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func want(rows ...string) string { return strings.Join(rows, "\n") + "\n" }

// TestEmptyCellIsBlankBraille is plan task 09.1, the mapping's
// TestParityP03a_EmptyCell; with it P-01, the grid, and P-02, the dot masks.
func TestEmptyCellIsBlankBraille(t *testing.T) { testEmptyCell(t) }
func TestParityP03a_EmptyCell(t *testing.T)    { testEmptyCell(t) }

func testEmptyCell(t *testing.T) {
	c := canvas(t, 3, 2)
	for row := range 2 {
		for col := range 3 {
			if g, _ := c.Cell(col, row); g != rune(0x2800) {
				t.Errorf("cell %d,%d of an empty canvas is %q, want the blank braille character, not a space", col, row, g)
			}
		}
	}
}

func TestParityP01_PixelGrid(t *testing.T) {
	c := canvas(t, 5, 3)
	if w, h := c.Dots(); w != 10 || h != 12 {
		t.Errorf("%dx%d dots for 5x3 cells, want two across and four down a cell", w, h)
	}
	c.Set(7, 9, 1)
	for row := range 3 {
		for col := range 5 {
			g, _ := c.Cell(col, row)
			if lit := g != rune(0x2800); lit != (col == 3 && row == 2) {
				t.Errorf("dot 7,9 lit cell %d,%d", col, row)
			}
		}
	}
	for _, bad := range [][2]int{{0, 1}, {1, 0}, {-1, 5}, {100000, 100000}} {
		if _, err := NewCanvas(bad[0], bad[1]); !isKind(err, fault.NoSize) {
			t.Errorf("%v: %v; want the no-size kind", bad, err)
		}
	}
}

func TestParityP02_DotBitmask(t *testing.T) {
	masks := [4][2]rune{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}
	for y := range 4 {
		for x := range 2 {
			c := canvas(t, 1, 1)
			c.Set(x, y, 1)
			if g, _ := c.Cell(0, 0); g != 0x2800+masks[y][x] {
				t.Errorf("dot %d,%d gives %U, want %U", x, y, g, 0x2800+masks[y][x])
			}
		}
	}
}

// TestLineRaster is plan task 09.2; TestParityP13_ThinLine is Bresenham.
func TestLineRaster(t *testing.T)         { testThinLine(t) }
func TestParityP13_ThinLine(t *testing.T) { testThinLine(t) }

func testThinLine(t *testing.T) {
	c := canvas(t, 4, 2) // 8 by 8 dots
	for _, to := range [][2]int{{7, 3}, {3, 7}} {
		c.Line(0, 0, to[0], to[1], 1, 1)
	}
	got := picture(c)
	golden := want(
		"##......",
		"#.##....",
		".#..##..",
		".#....##",
		"..#.....",
		"..#.....",
		"...#....",
		"...#....")
	if got != golden {
		t.Errorf("two shallow lines from the corner:\n%s\nwant\n%s", got, golden)
	}
	// Eight directions from the centre reach the same dots as their mirror.
	star := canvas(t, 5, 3) // 10 by 12
	for _, to := range [][2]int{{9, 5}, {0, 5}, {5, 0}, {5, 11}, {9, 1}, {1, 1}, {9, 9}, {1, 9}} {
		star.Line(5, 5, to[0], to[1], 1, 1)
	}
	back := canvas(t, 5, 3)
	for _, from := range [][2]int{{9, 5}, {0, 5}, {5, 0}, {5, 11}, {9, 1}, {1, 1}, {9, 9}, {1, 9}} {
		back.Line(from[0], from[1], 5, 5, 1, 1)
	}
	count := strings.Count(picture(star), "#")
	if count != strings.Count(picture(back), "#") || count < 30 {
		t.Errorf("%d dots outward, %d inward", count, strings.Count(picture(back), "#"))
	}
	for _, p := range [][2]int{{9, 5}, {0, 5}, {5, 0}, {5, 11}, {9, 1}, {1, 1}, {9, 9}, {1, 9}, {5, 5}} {
		if !star.Lit(p[0], p[1]) {
			t.Errorf("the line's end %v is not lit", p)
		}
	}
	dot := canvas(t, 1, 1)
	dot.Line(1, 2, 1, 2, 1, 1)
	if strings.Count(picture(dot), "#") != 1 {
		t.Error("a line from a dot to itself is that dot")
	}
}

// TestParityP14_ThickLine: upstream's thick line, with its half width of
// (width-1+1)/2.
func TestParityP14_ThickLine(t *testing.T) {
	c := canvas(t, 6, 3) // 12 by 12
	c.Line(1, 6, 10, 6, 3, 1)
	got := picture(c)
	// Upstream's half width for a width of 3 is (3-1+1)/2 = 1.5, and it is
	// spread to one side of the line: two rows of dots, not three. Matched as
	// it is (P-14); the style's widths are chosen knowing it.
	for y, row := range strings.Split(strings.TrimSpace(got), "\n") {
		lit := strings.Count(row, "#")
		if (y == 5 || y == 6) != (lit == 10) || (y != 5 && y != 6 && lit != 0) {
			t.Errorf("row %d has %d dots; upstream draws a level line of width 3 as rows 5 and 6:\n%s", y, lit, got)
		}
	}
	thin := canvas(t, 6, 3)
	thin.Line(1, 1, 10, 9, 1, 1)
	thick := canvas(t, 6, 3)
	thick.Line(1, 1, 10, 9, 4, 1)
	if strings.Count(picture(thick), "#") < 2*strings.Count(picture(thin), "#") {
		t.Errorf("a slanted line four wide:\n%s", picture(thick))
	}
	for y := range 12 {
		for x := range 12 {
			if thin.Lit(x, y) && !thick.Lit(x, y) {
				t.Errorf("dot %d,%d is on the thin line and not the thick one", x, y)
			}
		}
	}
}

// TestFillEvenOdd is plan task 09.3 (FR-11); TestParityP15 and P16 with it.
func TestFillEvenOdd(t *testing.T) {
	c := canvas(t, 6, 3) // 12 by 12
	outer := []Point{{1, 1}, {10, 1}, {10, 10}, {1, 10}, {1, 1}}
	hole := []Point{{4, 4}, {7, 4}, {7, 7}, {4, 7}, {4, 4}}
	c.Fill([][]Point{outer, hole}, 1)
	if !c.Lit(2, 2) || !c.Lit(9, 9) || !c.Lit(2, 5) {
		t.Errorf("the body is not filled:\n%s", picture(c))
	}
	if c.Lit(5, 5) || c.Lit(6, 6) {
		t.Errorf("the hole is filled:\n%s", picture(c))
	}
	if c.Lit(0, 0) || c.Lit(11, 11) || c.Lit(11, 5) {
		t.Errorf("the fill left its polygon:\n%s", picture(c))
	}
	// Two features that overlap: each is filled, so the overlap is inside.
	two := canvas(t, 6, 3)
	two.Fill([][]Point{{{1, 1}, {7, 1}, {7, 7}, {1, 7}, {1, 1}}}, 1)
	two.Fill([][]Point{{{4, 4}, {10, 4}, {10, 10}, {4, 10}, {4, 4}}}, 1)
	if !two.Lit(5, 5) || !two.Lit(6, 6) {
		t.Errorf("the overlap of two parts is not inside:\n%s", picture(two))
	}
}

func TestParityP15_PolygonFill(t *testing.T) {
	c := canvas(t, 6, 3)
	c.Fill([][]Point{{{1, 1}, {10, 1}, {5, 10}, {1, 1}}}, 1)
	widths := []int{}
	for _, row := range strings.Split(strings.TrimSpace(picture(c)), "\n") {
		widths = append(widths, strings.Count(row, "#"))
	}
	for y := 2; y < 9; y++ {
		if widths[y] > widths[y-1] || widths[y] == 0 {
			t.Errorf("a triangle pointing down narrows row by row: %v\n%s", widths, picture(c))
			break
		}
	}
	// An unclosed ring is closed.
	open := canvas(t, 6, 3)
	open.Fill([][]Point{{{1, 1}, {10, 1}, {10, 10}, {1, 10}}}, 1)
	if !open.Lit(5, 5) {
		t.Error("a ring that does not repeat its first point was not filled")
	}
}

func TestParityP16_DegenerateRings(t *testing.T) {
	c := canvas(t, 6, 3)
	c.Fill([][]Point{{{1, 1}, {10, 10}}, {{4, 4}, {7, 4}, {7, 7}, {4, 4}}}, 1)
	if strings.Contains(picture(c), "#") {
		t.Error("an outer ring of two points drew something; it aborts the polygon")
	}
	c.Fill([][]Point{{{1, 1}, {10, 1}, {10, 10}, {1, 10}, {1, 1}}, {{5, 5}, {6, 6}}}, 1)
	if !c.Lit(5, 5) {
		t.Error("a hole of two points was not skipped")
	}
	c.Fill(nil, 1)
	c.Fill([][]Point{nil}, 1)
}

// TestClipNothingOutside is plan task 09.4 (FR-11): nothing is ever drawn
// outside the rectangle, and geometry far outside it costs nothing.
func TestClipNothingOutside(t *testing.T) {
	c := canvas(t, 4, 2)
	c.Line(-1_000_000_000, -1_000_000_000, 1_000_000_000, 1_000_000_000, 1, 1)
	c.Line(-2_000_000_000, 3, 2_000_000_000, 3, 5, 1)
	c.Line(100, 100, 200, 300, 1, 1)
	c.Fill([][]Point{{{-1_000_000_000, -1_000_000_000}, {1_000_000_000, -1_000_000_000}, {1_000_000_000, 2}, {-1_000_000_000, 2}}}, 1)
	c.Set(-1, 0, 1)
	c.Set(0, 99, 1)
	c.SetForced(8, 8, 1)
	got := picture(c)
	if len(got) != 8*9 {
		t.Fatalf("the picture is %d bytes", len(got))
	}
	for x := range 8 {
		if !c.Lit(x, 3) || !c.Lit(x, 0) {
			t.Errorf("column %d: the parts inside the rectangle are drawn:\n%s", x, got)
			break
		}
	}
	if !c.Lit(6, 6) {
		t.Errorf("the long diagonal does not cross the rectangle where it should:\n%s", got)
	}
}

// TestCellColourVote is plan task 09.5, the mapping's TestParityP08.
func TestCellColourVote(t *testing.T)           { testVote(t) }
func TestParityP08_CellColourVote(t *testing.T) { testVote(t) }

func testVote(t *testing.T) {
	const road, water = 1, 2
	c := canvas(t, 3, 3)
	// The middle cell: three dots of water, two of road. The majority wins.
	for _, d := range [][2]int{{2, 4}, {2, 5}, {2, 6}} {
		c.Set(d[0], d[1], water)
	}
	for _, d := range [][2]int{{3, 4}, {3, 5}} {
		c.Set(d[0], d[1], road)
	}
	if _, ink := c.Cell(1, 1); ink != water {
		t.Errorf("three dots against two: ink %d, want the majority's", ink)
	}
	// A tie goes to the colour commoner among the eight neighbours.
	c.Set(3, 6, road)
	for _, d := range [][2]int{{0, 0}, {1, 1}, {4, 8}} {
		c.Set(d[0], d[1], road)
	}
	c.Set(5, 0, water)
	if _, ink := c.Cell(1, 1); ink != road {
		t.Errorf("a tie, with road commoner around it: ink %d", ink)
	}
	// When the neighbours tie as well, the colour met first in the cell
	// wins, scanning its dots row by row: the same answer on every machine.
	tie := canvas(t, 1, 1)
	tie.Set(1, 0, road)
	tie.Set(0, 0, water)
	tie.Set(0, 1, road)
	tie.Set(1, 1, water)
	if _, ink := tie.Cell(0, 0); ink != water {
		t.Errorf("a tie all round: ink %d, want the colour of the first dot in the scan", ink)
	}
}

// TestParityP09_ForcedPixels: a forced dot locks its whole cell's colour.
func TestParityP09_ForcedPixels(t *testing.T) {
	const road, marker = 1, 3
	c := canvas(t, 1, 1)
	for _, d := range [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}} {
		c.Set(d[0], d[1], road)
	}
	c.SetForced(1, 3, marker)
	if _, ink := c.Cell(0, 0); ink != marker {
		t.Errorf("one forced dot among five: ink %d, want the forced one's", ink)
	}
	c.Wipe()
	if g, ink := c.Cell(0, 0); g != rune(0x2800) || ink != 0 {
		t.Errorf("after a wipe: %U ink %d", g, ink)
	}
	c.Set(0, 0, road)
	if _, ink := c.Cell(0, 0); ink != road {
		t.Error("the lock outlived the wipe")
	}
}

// Package render turns what is on hand into a frame. It reads only what it is
// given, never waits, and is a pure function of its inputs (FR-23, NFR-6).
package render

import (
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxCells bounds a canvas: far larger than any terminal, small enough
	// that its buffers are a few megabytes at most.
	maxCells = 1 << 20
	// clipMargin is how far outside the rectangle, in dots, a line is still
	// rastered as it stands. Beyond it the line is cut to the margin first,
	// so geometry far away costs nothing (FR-11).
	clipMargin = 64
	// blank is the empty cell: the blank braille character, not a space (P-03a).
	blank = rune(0x2800)
)

// Point is a position in dot space: two dots across and four down a cell.
type Point struct{ X, Y int }

// Canvas is the braille canvas (L2 Render): a mask of eight dots a cell, the
// ink each dot was drawn in, and which cells have their ink locked. An ink
// is a small number the caller gives meaning to; zero is "none".
type Canvas struct {
	cols, rows int
	w, h       int
	mask       []uint8   // a cell's lit dots (P-02)
	ink        []uint8   // each dot's ink
	last       []uint8   // the ink last drawn in a cell
	locked     []bool    // a forced dot locks its cell's ink (P-09)
	crossings  []float64 // the fill's scratch space, kept between fills
}

// dotMask is a dot's bit within its cell, by row and column (P-02).
func dotMask(x, y int) uint8 {
	return [4][2]uint8{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}[y&3][x&1]
}

// NewCanvas makes a canvas of cols by rows cells.
func NewCanvas(cols, rows int) (*Canvas, error) {
	if cols <= 0 || rows <= 0 || cols > maxCells || rows > maxCells || cols*rows > maxCells {
		return nil, fault.Make(fault.NoSize,
			textsafe.Const("the map has no size to draw at"),
			textsafe.Const("its width and height in cells must each be at least 1, and together no more than a million cells"),
			textsafe.Const("give the map a size before drawing"))
	}
	c := &Canvas{cols: cols, rows: rows, w: 2 * cols, h: 4 * rows}
	c.mask = make([]uint8, cols*rows)
	c.last = make([]uint8, cols*rows)
	c.locked = make([]bool, cols*rows)
	c.ink = make([]uint8, c.w*c.h)
	return c, nil
}

// Dots is the canvas's size in dots.
func (c *Canvas) Dots() (w, h int) {
	if c == nil {
		return 0, 0
	}
	return c.w, c.h
}

// Wipe clears the canvas for another frame, keeping its buffers.
func (c *Canvas) Wipe() {
	if c == nil {
		return
	}
	clear(c.mask)
	clear(c.ink)
	clear(c.last)
	clear(c.locked)
}

// Set lights a dot in an ink. A dot outside the canvas is not drawn (FR-11).
func (c *Canvas) Set(x, y int, ink uint8) {
	if c == nil || x < 0 || y < 0 {
		return
	}
	if x >= c.w || y >= c.h {
		return
	}
	cell := (x >> 1) + c.cols*(y>>2) // P-01
	c.mask[cell] |= dotMask(x, y)
	c.last[cell] = ink
	c.ink[y*c.w+x] = ink
}

// SetForced lights a dot and locks its whole cell to the dot's ink, past the
// vote: a marker must show whatever it is drawn over (P-09).
func (c *Canvas) SetForced(x, y int, ink uint8) {
	if c == nil || x < 0 || y < 0 {
		return
	}
	if x >= c.w || y >= c.h {
		return
	}
	c.Set(x, y, ink)
	c.locked[(x>>1)+c.cols*(y>>2)] = true
}

// Lit reports whether a dot is lit.
func (c *Canvas) Lit(x, y int) bool {
	if c == nil || x < 0 || y < 0 {
		return false
	}
	if x >= c.w || y >= c.h {
		return false
	}
	return c.mask[(x>>1)+c.cols*(y>>2)]&dotMask(x, y) != 0
}

// Cell is a cell's braille character and its one ink: the majority among its
// lit dots (P-08). An empty cell is the blank braille character and no ink.
func (c *Canvas) Cell(col, row int) (rune, uint8) {
	if c == nil || col < 0 || row < 0 {
		return blank, 0
	}
	if col >= c.cols || row >= c.rows {
		return blank, 0
	}
	cell := col + c.cols*row
	if c.mask[cell] == 0 {
		return blank, 0
	}
	return blank + rune(c.mask[cell]), c.vote(col, row)
}

// tally counts a cell's lit dots by ink, in the order the inks are first met
// scanning the cell row by row. It returns how many inks it found.
func (c *Canvas) tally(col, row int, inks *[8]uint8, counts *[8]int) int {
	n := 0
	for dy := range 4 {
		for dx := range 2 {
			x, y := 2*col+dx, 4*row+dy
			if !c.Lit(x, y) {
				continue
			}
			ink, at := c.ink[y*c.w+x], -1
			for i := range n {
				if inks[i] == ink {
					at = i
				}
			}
			if at < 0 {
				at = n
				inks[n] = ink
				n++
			}
			counts[at]++
		}
	}
	return n
}

// around counts the lit dots of one ink in the eight cells around a cell.
func (c *Canvas) around(col, row int, ink uint8) int {
	score := 0
	for ny := max(row-1, 0); ny <= min(row+1, c.rows-1); ny++ {
		for nx := max(col-1, 0); nx <= min(col+1, c.cols-1); nx++ {
			if nx == col && ny == row {
				continue
			}
			for dy := range 4 {
				for dx := range 2 {
					x, y := 2*nx+dx, 4*ny+dy
					if c.Lit(x, y) && c.ink[y*c.w+x] == ink {
						score++
					}
				}
			}
		}
	}
	return score
}

// vote is upstream's rule (P-08, D-83): the majority ink among the cell's lit
// dots; a tie goes to the tied ink commoner among the eight neighbouring
// cells, so an area wins over a line one dot wide; and if that ties too, the
// tied ink met first in the cell - the same answer on every machine.
func (c *Canvas) vote(col, row int) uint8 {
	cell := col + c.cols*row
	if c.locked[cell] {
		return c.last[cell]
	}
	var inks [8]uint8
	var counts [8]int
	n := c.tally(col, row, &inks, &counts)
	if n == 0 {
		return c.last[cell]
	}
	most := 0
	for i := range n {
		most = max(most, counts[i])
	}
	best, bestScore := inks[0], -1
	for i := range n {
		if counts[i] != most {
			continue
		}
		if score := c.around(col, row, inks[i]); score > bestScore {
			best, bestScore = inks[i], score
		}
	}
	return best
}

// Line draws a line. One dot wide it is Bresenham's (P-13); wider, upstream's
// thick line with its half width (P-14).
func (c *Canvas) Line(x0, y0, x1, y1, width int, ink uint8) {
	if c == nil || width > 64 {
		return
	}
	x0, y0, x1, y1, ok := c.clip(x0, y0, x1, y1, width)
	if !ok {
		return
	}
	if width <= 1 {
		c.thin(x0, y0, x1, y1, ink)
		return
	}
	c.thick(x0, y0, x1, y1, width, ink)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(from, to int) int {
	if from < to {
		return 1
	}
	return -1
}

func (c *Canvas) thin(x0, y0, x1, y1 int, ink uint8) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x0, x1), sign(y0, y1)
	err := dx + dy
	for range dx - dy + 1 {
		c.Set(x0, y0, ink)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

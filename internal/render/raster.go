package render

import (
	"math"
	"sort"
)

// clip cuts a line to the canvas grown by a margin, so that a line from far
// away costs what its visible part costs. A line wholly inside the margin is
// returned as it stands, and rasters exactly as upstream's does.
func (c *Canvas) clip(x0, y0, x1, y1, width int) (int, int, int, int, bool) {
	m := float64(clipMargin + max(width, 1))
	left, top, right, bottom := -m, -m, float64(c.w-1)+m, float64(c.h-1)+m
	fx0, fy0, fx1, fy1 := float64(x0), float64(y0), float64(x1), float64(y1)
	inside := func(x, y float64) bool { return x >= left && x <= right && y >= top && y <= bottom }
	if inside(fx0, fy0) && inside(fx1, fy1) {
		return x0, y0, x1, y1, true
	}
	t0, t1 := 0.0, 1.0
	dx, dy := fx1-fx0, fy1-fy0
	for _, edge := range [4][2]float64{{-dx, fx0 - left}, {dx, right - fx0}, {-dy, fy0 - top}, {dy, bottom - fy0}} {
		p, q := edge[0], edge[1]
		if p == 0 && q < 0 {
			return 0, 0, 0, 0, false
		}
		if p == 0 {
			continue
		}
		r := q / p
		if p < 0 {
			t0 = math.Max(t0, r)
		} else {
			t1 = math.Min(t1, r)
		}
	}
	if t0 > t1 {
		return 0, 0, 0, 0, false
	}
	round := func(v float64) int { return int(math.Round(v)) }
	return round(fx0 + t0*dx), round(fy0 + t0*dy), round(fx0 + t1*dx), round(fy0 + t1*dy), true
}

// thick is upstream's thick line (P-14): an error-carrying walk along the
// line that also lights the dots within half the width to either side. Every
// loop is bounded by the line's length and width.
func (c *Canvas) thick(x0, y0, x1, y1, width int, ink uint8) {
	dx, dy := abs(x1-x0), abs(y1-y0)
	sx, sy := sign(x0, x1), sign(y0, y1)
	err := dx - dy
	ed := 1.0
	if dx+dy != 0 {
		ed = math.Sqrt(float64(dx*dx + dy*dy))
	}
	reach := ed * (float64(width-1) + 1) / 2
	span := dx + dy + width + 2
	for range span {
		c.Set(x0, y0, ink)
		e2, x2 := err, x0
		if 2*e2 >= -dx {
			e2 += dy
			y2 := y0
			for range span {
				if !(float64(e2) < reach && (y1 != y2 || dx > dy)) {
					break
				}
				y2 += sy
				c.Set(x0, y2, ink)
				e2 += dx
			}
			if x0 == x1 {
				return
			}
			e2 = err
			err -= dy
			x0 += sx
		}
		if 2*e2 <= dy {
			e2 = dx - e2
			for range span {
				if !(float64(e2) < reach && (x1 != x2 || dx < dy)) {
					break
				}
				x2 += sx
				c.Set(x2, y0, ink)
				e2 += dy
			}
			if y0 == y1 {
				return
			}
			err += dx
			y0 += sy
		}
	}
}

// usable returns the rings a fill draws: an outer ring of fewer than three
// points aborts the polygon, and such a hole is skipped (P-16).
func usable(rings [][]Point) [][]Point {
	if len(rings) == 0 || len(rings[0]) < 3 {
		return nil
	}
	kept := rings[:1]
	for _, hole := range rings[1:] {
		if len(hole) >= 3 {
			kept = append(kept[:len(kept):len(kept)], hole)
		}
	}
	return kept
}

// Fill fills one polygon: its outer ring and its holes, by the even-odd rule,
// a row of dots at a time. A dot is inside if its centre is. Two polygons
// that overlap are two fills, so their overlap is inside (FR-11). The fill
// method is free (D-11); that areas with holes are filled is what is matched
// (P-15).
func (c *Canvas) Fill(rings [][]Point, ink uint8) {
	if c == nil {
		return
	}
	rings = usable(rings)
	if rings == nil {
		return
	}
	top, bottom := c.h, -1
	for _, p := range rings[0] {
		top, bottom = min(top, p.Y), max(bottom, p.Y)
	}
	for y := max(top, 0); y <= min(bottom, c.h-1); y++ {
		c.crossings = c.crossings[:0]
		for _, ring := range rings {
			c.cross(ring, float64(y)+0.5)
		}
		sort.Float64s(c.crossings)
		for i := 0; i+1 < len(c.crossings); i += 2 {
			from := math.Max(math.Ceil(c.crossings[i]-0.5), 0)
			to := math.Min(math.Ceil(c.crossings[i+1]-0.5)-1, float64(c.w-1))
			for x := int(from); x <= int(to); x++ {
				c.Set(x, y, ink)
			}
		}
	}
}

// cross adds where a ring's edges cross the level of a row's dot centres.
// The ring is closed whether or not it repeats its first point.
func (c *Canvas) cross(ring []Point, level float64) {
	if len(ring) < 3 {
		return
	}
	prev := ring[len(ring)-1]
	for _, p := range ring {
		a, b := prev, p
		prev = p
		if a.Y == b.Y {
			continue
		}
		lo, hi := math.Min(float64(a.Y), float64(b.Y)), math.Max(float64(a.Y), float64(b.Y))
		if level < lo || level >= hi {
			continue
		}
		c.crossings = append(c.crossings, float64(a.X)+(level-float64(a.Y))*float64(b.X-a.X)/float64(b.Y-a.Y))
	}
}

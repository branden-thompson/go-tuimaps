package render

// SimplifyTolerance is upstream's tolerance for simplifying a line, in dots
// (P-31). Upstream leaves simplification off by default, and so does this.
const SimplifyTolerance = 0.5

// simplify drops the points of a line that lie within the tolerance of the
// line between the points kept either side of them: Ramer-Douglas-Peucker,
// the algorithm upstream uses (P-31), worked out with an explicit stack
// rather than by calling itself.
func simplify(ring []Point, into []Point, tolerance float64) []Point {
	into = into[:0]
	if len(ring) < 3 || tolerance <= 0 {
		return append(into, ring...)
	}
	keep := make([]bool, len(ring))
	keep[0], keep[len(ring)-1] = true, true
	stack := [][2]int{{0, len(ring) - 1}}
	for range len(ring) * 2 { // each span is split at most once a vertex
		if len(stack) == 0 {
			break
		}
		span := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		at, far := furthest(ring, span[0], span[1])
		if far <= tolerance*tolerance {
			continue
		}
		keep[at] = true
		stack = append(stack, [2]int{span[0], at}, [2]int{at, span[1]})
	}
	for i, p := range ring {
		if keep[i] {
			into = append(into, p)
		}
	}
	return into
}

// furthest is the point of a span that lies furthest from the line between
// its ends, and the square of how far that is.
func furthest(ring []Point, from, to int) (int, float64) {
	at, far := from, 0.0
	ax, ay := float64(ring[from].X), float64(ring[from].Y)
	bx, by := float64(ring[to].X), float64(ring[to].Y)
	dx, dy := bx-ax, by-ay
	length := dx*dx + dy*dy
	for i := from + 1; i < to; i++ {
		px, py := float64(ring[i].X), float64(ring[i].Y)
		var d float64
		if length == 0 {
			d = (px-ax)*(px-ax) + (py-ay)*(py-ay)
		} else {
			cross := (px-ax)*dy - (py-ay)*dx
			d = cross * cross / length
		}
		if d > far {
			at, far = i, d
		}
	}
	return at, far
}

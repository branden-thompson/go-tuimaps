package colour

// Depth is the colour depth a frame is drawn at: a hint from the host, never
// found out by asking the terminal (FR-17).
type Depth uint8

// The depths. Truecolor is the zero value.
const (
	Truecolor  Depth = iota
	Colours256       // indices 16 to 255 only: the ones every terminal agrees on
	Colours16        // the terminal owns these colours; checks are indicative
	NoColour         // no colour sequence at all (NFR-15)
)

// String names the depth.
func (d Depth) String() string {
	switch d {
	case Colours256:
		return "256 colours"
	case Colours16:
		return "16 colours"
	case NoColour:
		return "no colour"
	}
	return "truecolor"
}

// cubeLevel is the nearest of the 256 palette's six channel levels, and its
// place among them.
func cubeLevel(v uint8) (int, uint8) {
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	best := 0
	for i, l := range levels {
		if absDiff(v, l) < absDiff(v, levels[best]) {
			best = i
		}
	}
	return best, levels[best]
}

func absDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func squared(a, b RGB) int {
	dr, dg, db := absDiff(a.R, b.R), absDiff(a.G, b.G), absDiff(a.B, b.B)
	return dr*dr + dg*dg + db*db
}

// To256 is the entry of the 256-colour palette nearest a colour, and the
// colour that entry shows. Only entries 16 to 255 are used: the colour cube
// and the grey ramp, which are fixed. Entries 0 to 15 are whatever the
// person's theme says, so nothing can be checked against them (FR-16).
func To256(c RGB) (uint8, RGB) {
	ri, r := cubeLevel(c.R)
	gi, g := cubeLevel(c.G)
	bi, b := cubeLevel(c.B)
	cube := RGB{r, g, b}
	step := (int(c.R) + int(c.G) + int(c.B)) / 3
	gi24 := min(max((step-8+5)/10, 0), 23)
	level := uint8(8 + 10*gi24)
	grey := RGB{level, level, level}
	if squared(c, grey) < squared(c, cube) {
		return uint8(232 + gi24), grey
	}
	return uint8(16 + 36*ri + 6*gi + bi), cube
}

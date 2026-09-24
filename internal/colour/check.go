package colour

// threshold is the least difference between two classes of a ramp, and
// between a class and its ground, under every kind of colour vision (D-88).
const threshold = 10.0

// Rule is one of the things the checker holds a ramp to (D-53, FR-16).
type Rule uint8

// The rules.
const (
	Ordered       Rule = iota + 1 // relative luminance moves one way, or one way on each side of a midpoint
	Distinct                      // every class is a different palette entry at the depth in use
	Readable                      // line work 3:1 against its ground; text 4.5:1 on every class
	VisionSafe                    // every pair of classes, and every class against the ground, at least 10 apart
	BlendSeparate                 // blended toward an alert's tint, every class still at least 10 from every other (L-11.2)
	BlendVisible                  // blended, every class at least 5 from the same class outside the area (L-11.5)
)

// String names the rule.
func (r Rule) String() string {
	switch r {
	case Ordered:
		return "ordered"
	case Distinct:
		return "distinct"
	case Readable:
		return "readable"
	case VisionSafe:
		return "colour-vision-safe"
	case BlendSeparate:
		return "blend-separate"
	case BlendVisible:
		return "blend-visible"
	}
	return "unknown"
}

// Finding is one way a ramp breaks a rule: the two classes concerned, and
// the figure that fell short. B is -1 when the second party is the ground.
type Finding struct {
	Rule  Rule
	A, B  int
	Value float64
}

// RampCheck says how a ramp is used, which decides what is asked of it.
type RampCheck struct {
	Ground RGB
	Depth  Depth
	// Midpoint is the class a diverging ramp is lightest or darkest at, such
	// as the pale class at freezing; -1 for a ramp that runs one way.
	Midpoint int
	// Lines says the ramp colours line work drawn on the ground, such as
	// alert outlines. Otherwise it colours areas that text is drawn over.
	Lines bool
}

// CheckRamp holds a ramp to the four rules and returns every breach. The
// library's own presets must pass with none; a host's colours are reported
// on and never refused (D-53). A host can run it in its own tests.
func Check(ramp []RGB, c RampCheck) []Finding {
	if len(ramp) == 0 {
		return nil
	}
	shown := make([]RGB, len(ramp))
	for i, colour := range ramp {
		shown[i] = colour
		if c.Depth == Colours256 {
			_, shown[i] = To256(colour)
		}
	}
	var out []Finding
	out = ordered(out, shown, c.Midpoint)
	out = distinct(out, shown)
	out = readable(out, shown, c)
	return visionSafe(out, shown, c.Ground)
}

// ordered asks that luminance move one way along each run: the whole ramp,
// or the two sides of the midpoint, which run opposite ways.
func ordered(out []Finding, ramp []RGB, midpoint int) []Finding {
	if len(ramp) < 2 {
		return out
	}
	if midpoint <= 0 || midpoint >= len(ramp)-1 {
		return run(out, ramp, 0, len(ramp)-1, Luminance(ramp[len(ramp)-1]) > Luminance(ramp[0]))
	}
	rising := Luminance(ramp[midpoint]) > Luminance(ramp[0])
	out = run(out, ramp, 0, midpoint, rising)
	return run(out, ramp, midpoint, len(ramp)-1, !rising)
}

// run reports each step from one class to the next that goes the wrong way.
func run(out []Finding, ramp []RGB, from, to int, rising bool) []Finding {
	if from < 0 || to >= len(ramp) {
		return out
	}
	for i := from; i < to; i++ {
		step := Luminance(ramp[i+1]) - Luminance(ramp[i])
		if (rising && step <= 0) || (!rising && step >= 0) {
			out = append(out, Finding{Rule: Ordered, A: i, B: i + 1, Value: step})
		}
	}
	return out
}

func distinct(out []Finding, ramp []RGB) []Finding {
	for i := range ramp {
		for j := i + 1; j < len(ramp); j++ {
			if ramp[i] == ramp[j] {
				out = append(out, Finding{Rule: Distinct, A: i, B: j})
			}
		}
	}
	return out
}

// readable asks 3:1 of line work against its ground, and of an area that the
// text the foreground rule would draw on it reaches 4.5:1.
func readable(out []Finding, ramp []RGB, c RampCheck) []Finding {
	for i, colour := range ramp {
		if c.Lines {
			if ratio := Contrast(colour, c.Ground); ratio < LineContrast {
				out = append(out, Finding{Rule: Readable, A: i, B: -1, Value: ratio})
			}
			continue
		}
		text := Foreground(colour, colour, TextContrast)
		if ratio := Contrast(text, colour); ratio < TextContrast {
			out = append(out, Finding{Rule: Readable, A: i, B: i, Value: ratio})
		}
	}
	return out
}

// visionSafe is D-88: every pair of classes, not only neighbours, and every
// class against the ground.
func visionSafe(out []Finding, ramp []RGB, ground RGB) []Finding {
	for i := range ramp {
		for j := i + 1; j < len(ramp); j++ {
			if d := Difference(ramp[i], ramp[j]); d < threshold {
				out = append(out, Finding{Rule: VisionSafe, A: i, B: j, Value: d})
			}
		}
		if d := Difference(ramp[i], ground); d < threshold {
			out = append(out, Finding{Rule: VisionSafe, A: i, B: -1, Value: d})
		}
	}
	return out
}

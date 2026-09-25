package colour

import "math"

// visibleFloor is the least difference between a class inside an alert's
// area and the same class outside it, so that the warning stays visible
// inside the blend (L-11.5, D-45).
const visibleFloor = 5.0

// Blend shifts an image's class colour toward an alert's tint, in linear
// light, by a strength from 0 (the image alone) to 1 (the tint alone): the
// radar stays readable inside the area and the area stays visible (L-11.1).
func Blend(img, tint RGB, strength float64) RGB {
	strength = math.Max(0, math.Min(1, strength))
	encode := func(x float64) uint8 {
		if x <= 0.0031308 {
			x *= 12.92
		} else {
			x = 1.055*math.Pow(x, 1/2.4) - 0.055
		}
		return uint8(math.Round(math.Max(0, math.Min(1, x)) * 255))
	}
	mix := func(i, t uint8) uint8 { return encode((1-strength)*linear(i) + strength*linear(t)) }
	return RGB{R: mix(img.R, tint.R), G: mix(img.G, tint.G), B: mix(img.B, tint.B)}
}

// CheckBlend holds a ramp blended toward one tint to L-11.2 and L-11.5, as
// the depth will show it: every tinted class at least 10 from every other
// class, tinted or plain, and from the ground (BlendSeparate); and every
// tinted class at least 5 from the same class outside the area
// (BlendVisible). A is the tinted class; B the other class, or -1 for the
// ground. It takes one tint, so that a finding says which tint it is about.
func CheckBlend(ramp []RGB, tint, ground RGB, strength float64, depth Depth) []Finding {
	var out []Finding
	tinted, plain := make([]RGB, len(ramp)), make([]RGB, len(ramp))
	for i, c := range ramp {
		tinted[i], plain[i] = shownAt(Blend(c, tint, strength), depth), shownAt(c, depth)
	}
	ground = shownAt(ground, depth)
	ramp = plain // every comparison is of colours as the depth shows them
	for i := range tinted {
		for j := range ramp {
			if j > i {
				if d := Difference(tinted[i], tinted[j]); d < threshold {
					out = append(out, Finding{Rule: BlendSeparate, A: i, B: j, Value: d})
				}
			}
			if j != i {
				if d := Difference(tinted[i], ramp[j]); d < threshold {
					out = append(out, Finding{Rule: BlendSeparate, A: i, B: j, Value: d})
				}
			}
		}
		if d := Difference(tinted[i], ground); d < threshold {
			out = append(out, Finding{Rule: BlendSeparate, A: i, B: -1, Value: d})
		}
		if d := Difference(tinted[i], ramp[i]); d < visibleFloor {
			out = append(out, Finding{Rule: BlendVisible, A: i, B: i, Value: d})
		}
	}
	return out
}

// blendStrengths are the strengths the search tries, strongest first: D-14's
// specimens went to 50 %, and below 10 % a tint is not there to see.
var blendStrengths = []float64{0.50, 0.45, 0.40, 0.35, 0.30, 0.25, 0.20, 0.15, 0.10}

// Blends is how strongly each alert tint is blended over each image ramp, as
// the search found it for one palette, ground and depth (L-11.3). A tint that
// does not blend over a ramp has the image drawn over it instead: the
// warning is carried by the outline, label and severity digit (L-11.4, D-27).
type Blends struct {
	strength [2][5]float64 // by preset (temperature, radar), then tint (extreme to unknown); zero: not blended
}

// Strength is how strongly a tint blends over a preset's image, and whether
// it blends at all.
func (b Blends) Strength(p Preset, tint int) (float64, bool) {
	if p < Temperature || p > Radar || tint < 0 || tint >= 5 {
		return 0, false
	}
	s := b.strength[p-Temperature][tint]
	return s, s > 0
}

// Fallbacks are the tints, by preset, that do not blend and have the image
// drawn over them.
func (b Blends) Fallbacks() (out [][2]int) {
	for p := range 2 {
		for tint := range 5 {
			if b.strength[p][tint] == 0 {
				out = append(out, [2]int{p + int(Temperature), tint})
			}
		}
	}
	return out
}

// SearchBlends finds, for each image preset and alert tint, the strongest
// strength that keeps every class separated (L-11.2), reading the palette's
// colours at the ground and depth given, so that a host palette is searched
// as the library's own is (L3.8). It runs whenever the palette, the ground or
// the depth changes, never on a frame.
func SearchBlends(p Palette, ground GroundKind, groundColour RGB, depth Depth) Blends {
	var b Blends
	under := seenAs(shownAt(groundColour, depth))
	for _, preset := range []Preset{Temperature, Radar} {
		ramp := rampOf(p, preset, ground, depth)
		var plain [17]seen // the largest ramp, temperature's
		for i := range min(len(ramp), len(plain)) {
			plain[i] = seenAs(shownAt(ramp[i], depth))
		}
		for tint := range 5 {
			colour := tintOf(p, tint, ground, depth)
			for _, s := range blendStrengths {
				if separates(ramp, plain[:len(ramp)], under, colour, s, depth) {
					b.strength[preset-Temperature][tint] = s
					break
				}
			}
		}
	}
	return b
}

// separates is CheckBlend's BlendSeparate rule alone, stopping at the first
// pair too close: the search needs only whether a strength passes, and it
// runs on every change of look, so it must cost little (a new map's first
// frame pays for it).
func separates(ramp []RGB, plain []seen, under seen, tint RGB, strength float64, depth Depth) bool {
	var tinted [17]seen // the largest ramp, temperature's
	n := min(len(ramp), len(tinted))
	for i := range n {
		tinted[i] = seenAs(shownAt(Blend(ramp[i], tint, strength), depth))
	}
	for i := range n {
		if apart(tinted[i], under) < threshold {
			return false
		}
		for j := range n {
			if j > i && apart(tinted[i], tinted[j]) < threshold {
				return false
			}
			if j != i && apart(tinted[i], plain[j]) < threshold {
				return false
			}
		}
	}
	return true
}

// shownAt is a colour as the depth will show it.
func shownAt(c RGB, depth Depth) RGB {
	if depth == Colours256 {
		_, c = To256(c)
	}
	return c
}

// seen is a colour in Lab under each of the four kinds of vision, worked out
// once so that comparing it many times costs a distance each (Difference
// works both colours out again on every call).
type seen [4]Lab

func seenAs(c RGB) seen {
	var out seen
	for v := Normal; v <= Tritanopia; v++ {
		out[v-Normal] = InLab(c, v)
	}
	return out
}

// apart is Difference over colours already seen: the least of their
// distances under the four kinds of vision.
func apart(a, b seen) float64 {
	least := math.Inf(1)
	for v := range a {
		least = math.Min(least, a[v].Distance(b[v]))
	}
	return least
}

// rampOf is a preset's class colours as the palette resolves them.
func rampOf(p Palette, preset Preset, ground GroundKind, depth Depth) []RGB {
	first, n := Radar1, 6
	if preset == Temperature {
		first, n = Temperature1, 17
	}
	out := make([]RGB, 0, n)
	for i := range n {
		c, _ := p.ResolveAt(first+Token(i), ground, depth)
		out = append(out, c)
	}
	return out
}

// tintOf is an alert's tint, extreme (0) to unknown (4), as the palette
// resolves it.
func tintOf(p Palette, tint int, ground GroundKind, depth Depth) RGB {
	c, _ := p.ResolveAt(AlertExtremeTint+Token(2*tint), ground, depth)
	return c
}

// CheckOverBlend holds an alert's outline and its label to their contrast
// over every background an image ramp takes inside the alert's area: each
// class blended toward the tint at the strength given, or where the strength
// is zero, the class itself, drawn over the tint (L-8.8). The outline must
// reach 3:1 and the label 4.5:1, as the depth shows them, after the
// foreground rule has chosen what is drawn. A is the class; B is 0 for the
// outline and 1 for the label.
func CheckOverBlend(ink RGB, ramp []RGB, tint RGB, strength float64, depth Depth) []Finding {
	var out []Finding
	for i, class := range ramp {
		bg := class
		if strength > 0 {
			bg = Blend(class, tint, strength)
		}
		bg = shownAt(bg, depth)
		own := shownAt(ink, depth)
		for which, need := range []float64{LineContrast, TextContrast} {
			if got := Contrast(Foreground(own, bg, need), bg); got < need {
				out = append(out, Finding{Rule: Readable, A: i, B: which, Value: got})
			}
		}
	}
	return out
}

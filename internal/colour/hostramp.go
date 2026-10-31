package colour

import "math"

// HostRamp is a host type's colour for each of n classes (L-2.2): the
// host's own, one a class, when it gave them; otherwise the low, middle and
// high tokens' colours spread across the classes, low to middle over the
// first half and middle to high over the second, evenly in Lab. Nil at a
// depth with no ramp, or for no classes.
//
// Source: the 1976 CIE L*a*b* space under D65 (toLab and its inverse).
func HostRamp(own []RGB, n int, p Palette, ground GroundKind, depth Depth) []RGB {
	if n <= 0 || (depth != Truecolor && depth != Colours256) {
		return nil
	}
	if len(own) == n {
		return append([]RGB(nil), own...)
	}
	low, okL := p.ResolveAt(Low, ground, depth)
	middle, okM := p.ResolveAt(Middle, ground, depth)
	high, okH := p.ResolveAt(High, ground, depth)
	if !okL || !okM || !okH {
		return nil
	}
	if n == 1 {
		return []RGB{middle}
	}
	a, b, c := InLab(low, Normal), InLab(middle, Normal), InLab(high, Normal)
	out := make([]RGB, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		from, to, at := a, b, 2*t
		if t > 0.5 {
			from, to, at = b, c, 2*t-1
		}
		out[i] = fromLab(Lab{L: from.L + (to.L-from.L)*at, A: from.A + (to.A-from.A)*at, B: from.B + (to.B-from.B)*at})
	}
	return out
}

// fromLab takes Lab under D65 back to sRGB, each channel held to the gamut.
func fromLab(c Lab) RGB {
	fy := (c.L + 16) / 116
	fx, fz := fy+c.A/500, fy-c.B/200
	f := func(t float64) float64 {
		if t > 0.206893 {
			return t * t * t
		}
		return (t - 16.0/116) / 7.787
	}
	x, y, z := 0.95047*f(fx), f(fy), 1.08883*f(fz)
	r := 3.2404542*x - 1.5371385*y - 0.4985314*z
	g := -0.9692660*x + 1.8760108*y + 0.0415560*z
	b := 0.0556434*x - 0.2040259*y + 1.0572252*z
	return RGB{R: encode(r), G: encode(g), B: encode(b)}
}

// encode applies the sRGB transfer curve to a linear channel and rounds it
// to a byte, held to 0 to 255.
func encode(v float64) uint8 {
	v = math.Max(0, math.Min(1, v))
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return uint8(math.Round(v * 255))
}

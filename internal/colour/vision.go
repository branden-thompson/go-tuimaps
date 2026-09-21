package colour

import "math"

// Vision is a kind of colour vision a ramp is checked under (D-88).
type Vision uint8

// The kinds of vision: the common one, and the three deficiencies at full
// severity.
const (
	Normal Vision = iota
	Protanopia
	Deuteranopia
	Tritanopia
)

// String names the kind.
func (v Vision) String() string {
	switch v {
	case Protanopia:
		return "protanopia"
	case Deuteranopia:
		return "deuteranopia"
	case Tritanopia:
		return "tritanopia"
	}
	return "normal"
}

// Lab is a colour in the 1976 CIE L*a*b* space, under the D65 white.
type Lab struct{ L, A, B float64 }

// Distance is how far apart two colours are in Lab: the 1976 difference.
func (p Lab) Distance(q Lab) float64 {
	return math.Sqrt((p.L-q.L)*(p.L-q.L) + (p.A-q.A)*(p.A-q.A) + (p.B-q.B)*(p.B-q.B))
}

// matrix is the published simulation of a deficiency at full severity
// (Machado, Oliveira and Fernandes, 2009). It was fitted to linear light and
// is applied there, never to gamma-encoded values: the same two colours can
// score 27 one way and 7 the other (D-88).
func matrix(v Vision) ([3][3]float64, bool) {
	switch v {
	case Protanopia:
		return [3][3]float64{{0.152286, 1.052583, -0.204868}, {0.114503, 0.786281, 0.099216}, {-0.003882, -0.048116, 1.051998}}, true
	case Deuteranopia:
		return [3][3]float64{{0.367322, 0.860646, -0.227968}, {0.280085, 0.672501, 0.047413}, {-0.011820, 0.042940, 0.968881}}, true
	case Tritanopia:
		return [3][3]float64{{1.255528, -0.076749, -0.178779}, {-0.078411, 0.930809, 0.147602}, {0.004733, 0.691367, 0.303900}}, true
	}
	return [3][3]float64{}, false
}

// apply multiplies and holds each channel to 0 to 1.
func apply(m [3][3]float64, ch [3]float64) [3]float64 {
	var out [3]float64
	for i := range 3 {
		out[i] = math.Max(0, math.Min(1, m[i][0]*ch[0]+m[i][1]*ch[1]+m[i][2]*ch[2]))
	}
	return out
}

// linearOf undoes the sRGB transfer curve for a channel given as 0 to 1.
func linearOf(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// toLab takes linear sRGB to Lab under D65.
func toLab(l [3]float64) Lab {
	x := 0.4124564*l[0] + 0.3575761*l[1] + 0.1804375*l[2]
	y := 0.2126729*l[0] + 0.7151522*l[1] + 0.0721750*l[2]
	z := 0.0193339*l[0] + 0.1191920*l[1] + 0.9503041*l[2]
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	fx, fy, fz := f(x/0.95047), f(y), f(z/1.08883)
	return Lab{L: 116*fy - 16, A: 500 * (fx - fy), B: 200 * (fy - fz)}
}

// InLab is a colour as a kind of vision sees it, in Lab: the deficiency is
// simulated in linear light, then the result is taken to Lab.
func InLab(c RGB, v Vision) Lab {
	ch := [3]float64{linear(c.R), linear(c.G), linear(c.B)}
	if m, ok := matrix(v); ok {
		ch = apply(m, ch)
	}
	return toLab(ch)
}

// Difference is how far apart two colours are for whoever tells them apart
// least: the smallest of their Lab distances under the four kinds of vision.
// D-88 asks for at least 10 between every pair of classes of a ramp, and
// between every class and its ground.
func Difference(a, b RGB) float64 {
	if a == b {
		return 0
	}
	least := math.Inf(1)
	for v := Normal; v <= Tritanopia; v++ {
		least = math.Min(least, InLab(a, v).Distance(InLab(b, v)))
	}
	return least
}

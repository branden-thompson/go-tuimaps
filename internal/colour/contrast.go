package colour

import "math"

// The contrast a foreground must reach on the cell it is drawn in (FR-16).
const (
	LineContrast = 3.0 // line work and glyphs
	TextContrast = 4.5 // labels and legend text
)

// linear undoes the sRGB transfer curve for one channel.
func linear(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Luminance is the relative luminance of a colour, 0 for black to 1 for
// white, as the contrast formula defines it.
func Luminance(c RGB) float64 {
	return 0.2126*linear(c.R) + 0.7152*linear(c.G) + 0.0722*linear(c.B)
}

// Contrast is the contrast ratio of two colours, 1 to 21, whichever way
// round they are given.
func Contrast(a, b RGB) float64 {
	la, lb := Luminance(a), Luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Foreground is the foreground rule (FR-16, D-77): a line's or a label's own
// colour where it reaches the contrast asked for on that cell's background;
// otherwise black or white, whichever contrasts more, found by computing
// both.
func Foreground(own, background RGB, atLeast float64) RGB {
	if !(atLeast >= 1 && atLeast <= 21) {
		atLeast = TextContrast // a threshold that is no ratio: hold the colour to the stricter one
	}
	if Contrast(own, background) >= atLeast {
		return own
	}
	if Contrast(RGB{}, background) > Contrast(RGB{255, 255, 255}, background) {
		return RGB{}
	}
	return RGB{255, 255, 255}
}

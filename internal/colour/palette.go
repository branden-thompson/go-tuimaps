package colour

import (
	"sort"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// GroundKind is whether the ground in effect is dark or light, which picks
// between the library's two sets of defaults (FR-20, D-64).
type GroundKind uint8

// The two kinds of ground. Dark is the zero value, as the painted default is.
const (
	Dark GroundKind = iota
	Light
)

// String names the kind.
func (g GroundKind) String() string {
	if g == Light {
		return "light"
	}
	return "dark"
}

// Palette is a host's colours for whichever tokens it chose to set. The
// zero value sets none, and every token keeps the library's default.
type Palette struct {
	set  map[Token]RGB
	safe bool // safe ramps: the host's colours for ramp tokens are passed over (D-63)
}

// NewPalette makes a palette from token names. Names that are no token are
// returned, sorted, for the caller to report; the rest take effect.
func NewPalette(values map[string]RGB) (Palette, []string) {
	if len(values) == 0 {
		return Palette{}, nil
	}
	p := Palette{set: make(map[Token]RGB, len(values))}
	var unknown []string
	for name, c := range values {
		tok, ok := ParseToken(name)
		if !ok {
			unknown = append(unknown, name)
			continue
		}
		p.set[tok] = c
	}
	sort.Strings(unknown)
	return p, unknown
}

// Resolve is the colour of a token: the host's if it set one, else the
// library's default for the kind of ground in effect. It is false for a
// value that is no token, and for a token with no default yet.
func (p Palette) Resolve(t Token, ground GroundKind) (RGB, bool) {
	return p.ResolveAt(t, ground, Truecolor)
}

// ResolveAt is Resolve at a depth. A ramp token's default is the library's
// own ramp for that depth; every other colour is the same at every depth and
// is mapped to the depth's palette when it is drawn.
func (p Palette) ResolveAt(t Token, ground GroundKind, depth Depth) (RGB, bool) {
	if t < Ground || t > TrackLabel {
		return RGB{}, false
	}
	ramp := t >= AlertExtremeOutline && t <= High
	if c, ok := p.set[t]; ok && !(ramp && p.safe) {
		return c, true
	}
	if depth == Colours16 {
		index, ok := Sixteen(t, ground)
		return SixteenReference()[index], ok
	}
	if c, ok := rampDefault(t, ground, depth); ok {
		return c, true
	}
	if ground == Light {
		return lightDefault(t)
	}
	return darkDefault(t)
}

// KeepRamps returns the palette with safe ramps on or off. On, every ramp is
// the library's own, whatever the host set for its tokens; the host's other
// colours stand (D-63).
func (p Palette) KeepRamps(on bool) Palette {
	p.safe = on
	return p
}

// sets reports whether the host's colour for any token from first to last
// is in effect.
func (p Palette) sets(first, last Token) bool {
	if p.safe {
		return false
	}
	for t := first; t <= last; t++ {
		if _, ok := p.set[t]; ok {
			return true
		}
	}
	return false
}

// effective is the ramp from first to last, a step apart, as it is drawn.
func (p Palette) effective(first, last, step Token, ground GroundKind, depth Depth) []RGB {
	var ramp []RGB
	for t := first; t <= last; t += step {
		if c, ok := p.ResolveAt(t, ground, depth); ok {
			ramp = append(ramp, c)
		}
	}
	return ramp
}

// Warnings reports each of the library's ramps that a host's colours have
// made break a rule. The colours are used all the same: a host's palette is
// reported on, never refused (D-53, D-69).
func (p Palette) Warnings(ground GroundKind, depth Depth) []fault.Warning {
	if depth != Truecolor && depth != Colours256 {
		return nil
	}
	under, _ := p.ResolveAt(Ground, ground, depth)
	var out []fault.Warning
	report := func(name textsafe.Text, breaches int) {
		if breaches > 0 {
			out = append(out, fault.Warning{Kind: fault.RampRuleBroken, Subject: name, Count: breaches})
		}
	}
	if p.sets(Radar1, Radar6) {
		report(textsafe.Const("radar"), len(CheckRamp(p.effective(Radar1, Radar6, 1, ground, depth), RampCheck{Ground: under, Depth: depth, Midpoint: -1})))
	}
	if p.sets(Temperature1, Temperature17) {
		mid := Midpoint(Temperature, ground, depth)
		report(textsafe.Const("temperature"), len(CheckRamp(p.effective(Temperature1, Temperature17, 1, ground, depth), RampCheck{Ground: under, Depth: depth, Midpoint: mid})))
	}
	if p.sets(AlertExtremeOutline, AlertUnknownTint) {
		lines := CheckRamp(p.effective(AlertExtremeOutline, AlertUnknownOutline, 2, ground, depth), RampCheck{Ground: under, Depth: depth, Midpoint: -1, Lines: true})
		areas := CheckRamp(p.effective(AlertExtremeTint, AlertUnknownTint, 2, ground, depth), RampCheck{Ground: under, Depth: depth, Midpoint: -1})
		report(textsafe.Const("alerts"), countUnordered(lines)+countUnordered(areas))
	}
	return out
}

// countUnordered counts breaches other than of order: alert colours are a
// set of five, not a scale, and are not asked to run light to dark.
func countUnordered(findings []Finding) int {
	n := 0
	for _, f := range findings {
		if f.Rule != Ordered {
			n++
		}
	}
	return n
}

// GroundChoice is a host's choice about the ground (D-64). The zero value is
// the default: the library paints every cell's background with the ground
// token, so the map reads the same on any terminal and contrast is exact.
type GroundChoice struct {
	declared bool
	colour   RGB
}

// Declared is the other choice: the library paints no ground, and the host
// says what colour is behind the map. "Do not paint" cannot be chosen
// without saying so, because every contrast check reads the ground.
func Declared(c RGB) (GroundChoice, error) {
	return GroundChoice{declared: true, colour: c}, nil
}

// InEffect is the ground every check reads, and whether the library paints it.
func (g GroundChoice) InEffect(p Palette) (RGB, bool) {
	if g.declared {
		return g.colour, false
	}
	if c, ok := p.set[Ground]; ok {
		return c, true
	}
	return RGB{16, 22, 28}, true
}

// Kind is whether the ground in effect is dark or light: light when black
// contrasts with it more than white does, the same test as the foreground
// rule's.
func (g GroundChoice) Kind(p Palette) GroundKind {
	c, _ := g.InEffect(p)
	if Contrast(RGB{}, c) > Contrast(RGB{255, 255, 255}, c) {
		return Light
	}
	return Dark
}

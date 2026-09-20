package colour

import "sort"

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
	set map[Token]RGB
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
	if c, ok := p.set[t]; ok {
		return c, true
	}
	if c, ok := rampDefault(t, ground, depth); ok {
		return c, true
	}
	if ground == Light {
		return lightDefault(t)
	}
	return darkDefault(t)
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

package colour

import "testing"

// TestTheTerminatorAndNightTokensHaveDefaults is L-4.1: the terminator's
// line and the night's tint are tokens with the library's own colours on
// both grounds and at sixteen; the line reads on its ground; neither is any
// preset's class, nor an alert's.
func TestTheTerminatorAndNightTokensHaveDefaults(t *testing.T) {
	for _, c := range []struct {
		tok  Token
		name string
	}{{Terminator, "terminator"}, {Night, "night"}} {
		if c.tok.Name() != c.name {
			t.Errorf("%d is named %q; want %q", c.tok, c.tok.Name(), c.name)
		}
		if back, ok := ParseToken(c.name); !ok || back != c.tok {
			t.Errorf("%q does not parse back to its token", c.name)
		}
		if ScaleClass(c.tok) || (c.tok >= AlertExtremeOutline && c.tok <= AlertUnknownTint) {
			t.Errorf("%s is a preset's class or an alert's", c.name)
		}
		for _, g := range []GroundKind{Dark, Light} {
			if _, ok := (Palette{}).Resolve(c.tok, g); !ok {
				t.Errorf("%s has no default on the %v ground", c.name, g)
			}
			if _, ok := Sixteen(c.tok, g); !ok {
				t.Errorf("%s has no sixteen-colour entry on the %v ground", c.name, g)
			}
		}
	}
	for _, g := range []GroundKind{Dark, Light} {
		ground, _ := (Palette{}).Resolve(Ground, g)
		line, _ := (Palette{}).Resolve(Terminator, g)
		if r := Contrast(line, ground); r < LineContrast {
			t.Errorf("the terminator reads at %.1f:1 on the %v ground; line work needs %.1f", r, g, LineContrast)
		}
	}
}

// TestNightTintKeepsTheFieldReadable is L-4.4: every preset's ramp passes
// the checks - order by lightness, colour-vision safety, reading on its
// ground - against the ground the night shades, on both grounds, as it
// passes against the ground itself.
func TestNightTintKeepsTheFieldReadable(t *testing.T) {
	for _, g := range []GroundKind{Dark, Light} {
		ground, _ := (Palette{}).Resolve(Ground, g)
		night, _ := (Palette{}).Resolve(Night, g)
		shaded := Blend(ground, night, NightShade(g))
		if shaded == ground {
			t.Errorf("the night shades nothing on the %v ground", g)
		}
		for p := Temperature; p <= FoF2; p++ {
			ramp, _ := Ramp(p, g, Truecolor)
			rc := RampCheck{Ground: shaded, Depth: Truecolor, Midpoint: -1, Lines: p == Wind}
			if p == Temperature {
				rc.Midpoint = Midpoint(Temperature, g, Truecolor)
			}
			if f := Check(ramp, rc); len(f) != 0 {
				t.Errorf("preset %d on the %v ground at night: %v", p, g, f)
			}
		}
	}
}

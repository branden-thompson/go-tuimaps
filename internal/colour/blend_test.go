package colour

import (
	"math"
	"testing"
)

// TestBlendMixesInLinearLight is L3.6's colour half (L-11.1): the image
// shifted toward the tint in linear light, not in encoded values.
func TestBlendMixesInLinearLight(t *testing.T) {
	black, white := RGB{}, RGB{R: 255, G: 255, B: 255}
	if got := Blend(black, white, 0.5); got != (RGB{R: 188, G: 188, B: 188}) {
		t.Errorf("half of white over black in linear light: %v; want 188 (128 would be the encoded average)", got)
	}
	c := RGB{R: 10, G: 120, B: 240}
	if Blend(c, white, 0) != c || Blend(c, white, 1) != white {
		t.Error("strength 0 is the image and 1 is the tint")
	}
}

func radarAt(t *testing.T, ground GroundKind, depth Depth) ([]RGB, [5]RGB, RGB) {
	t.Helper()
	var none Palette
	ramp := rampOf(none, Radar, ground, depth)
	var tints [5]RGB
	for i := range tints {
		tints[i] = tintOf(none, i, ground, depth)
	}
	return ramp, tints, groundOf(ground, depth)
}

// TestCheckBlendHoldsBothRules is L3.7 (L-11.2, L-11.5): a blend strong
// enough to wash classes together breaks BlendSeparate, and one too faint to
// see breaks BlendVisible.
func TestCheckBlendHoldsBothRules(t *testing.T) {
	ramp, tints, ground := radarAt(t, Dark, Truecolor)
	count := func(f []Finding, r Rule) int {
		n := 0
		for _, one := range f {
			if one.Rule == r {
				n++
			}
		}
		return n
	}
	if f := CheckBlend(ramp, tints[1], ground, 0.20, Truecolor); count(f, BlendSeparate) != 0 {
		t.Errorf("radar under the severe tint at 20 %% on dark: %v; D-14's specimen passes", f)
	}
	washed := 0
	for _, tint := range tints {
		washed += count(CheckBlend(ramp, tint, ground, 0.50, Truecolor), BlendSeparate)
	}
	if washed == 0 {
		t.Error("at 50 % nothing washes out, so the separation rule is not biting")
	}
	if f := CheckBlend(ramp, tints[1], ground, 0.10, Truecolor); count(f, BlendVisible) == 0 {
		t.Error("at 10 % every class is visible inside the area, so the visibility rule is not biting")
	}
	// A tint the colour of the ground pulls a class onto it.
	onGround := CheckBlend([]RGB{{R: 40, G: 50, B: 60}}, ground, ground, 0.50, Truecolor)
	near := false
	for _, f := range onGround {
		near = near || (f.Rule == BlendSeparate && f.B == -1)
	}
	if !near {
		t.Errorf("a class blended half-way to the ground's own colour: %v; want it too near the ground", onGround)
	}
	// At 256 colours the check reads the colours the palette will show.
	one, reddish, white := []RGB{{R: 51, G: 51, B: 120}}, RGB{R: 200, G: 40, B: 40}, RGB{R: 255, G: 255, B: 255}
	if count(CheckBlend(one, reddish, white, 0.20, Truecolor), BlendVisible) != 0 {
		t.Fatal("the truecolour blend is not visible, so this proves nothing")
	}
	if count(CheckBlend(one, reddish, white, 0.20, Colours256), BlendVisible) == 0 {
		t.Error("at 256 colours the blend and the class land on one palette entry, and the check did not see it")
	}
}

// TestTheSearchPicksTheStrongestThatSeparates is L3.8 (L-11.3, L-11.4,
// D-27): for each image ramp, ground, depth and tint, the strength chosen is
// the strongest tried that keeps every class separated; where none does, the
// tint does not blend and the image is drawn over it. The PLAN measurements
// hold: radar blends under every tint on the dark ground, and on the light
// ground no strength passes (S29-6).
func TestTheSearchPicksTheStrongestThatSeparates(t *testing.T) {
	var none Palette
	for _, ground := range []GroundKind{Dark, Light} {
		for _, depth := range []Depth{Truecolor, Colours256} {
			b := SearchBlends(none, ground, groundOf(ground, depth), depth)
			for _, p := range []Preset{Radar, Temperature} {
				ramp := rampOf(none, p, ground, depth)
				for tint := range 5 {
					colour := tintOf(none, tint, ground, depth)
					s, blended := b.Strength(p, tint)
					separates := func(at float64) bool {
						for _, f := range CheckBlend(ramp, colour, groundOf(ground, depth), at, depth) {
							if f.Rule == BlendSeparate {
								return false
							}
						}
						return true
					}
					if !blended {
						if separates(blendStrengths[len(blendStrengths)-1]) {
							t.Errorf("%v %v %v tint %d: not blended, yet the faintest strength separates", p, ground, depth, tint)
						}
						continue
					}
					if !separates(s) {
						t.Errorf("%v %v %v tint %d: %.2f chosen and it does not separate", p, ground, depth, tint, s)
					}
					if stronger := s + 0.05; stronger <= blendStrengths[0]+1e-9 && separates(stronger) {
						t.Errorf("%v %v %v tint %d: %.2f chosen, yet %.2f separates too", p, ground, depth, tint, s, stronger)
					}
				}
			}
			for tint := range 5 {
				_, blended := b.Strength(Radar, tint)
				if ground == Dark && depth == Truecolor && !blended {
					t.Errorf("radar on the dark ground under tint %d does not blend; D-14's specimens did", tint)
				}
				if ground == Light && depth == Truecolor && blended {
					t.Errorf("radar on the light ground under tint %d blends; S29-6 found no strength that passes", tint)
				}
			}
		}
	}
	if math.Abs(blendStrengths[0]-0.50) > 1e-9 || math.Abs(blendStrengths[len(blendStrengths)-1]-0.10) > 1e-9 {
		t.Errorf("the strengths tried run %v; D-14's specimens went to 50 %%, and the faintest tried is 10 %%", blendStrengths)
	}
}

// TestAHostPaletteReRunsTheSearch: the search reads the host's colours. A
// palette that puts two radar classes almost together leaves no strength
// that keeps a tinted class apart from its twin outside the area, so no tint
// blends over radar and the image is drawn over it.
func TestAHostPaletteReRunsTheSearch(t *testing.T) {
	var none Palette
	ramp := rampOf(none, Radar, Dark, Truecolor)
	twin := ramp[3]
	twin.R++
	host, bad := NewPalette(map[string]RGB{(Radar1 + 2).Name(): twin})
	if len(bad) != 0 {
		t.Fatalf("the palette refused %v", bad)
	}
	if got := rampOf(host, Radar, Dark, Truecolor)[2]; got != twin {
		t.Fatalf("the host's radar.3 did not take: %v", got)
	}
	if _, blended := SearchBlends(none, Dark, groundOf(Dark, Truecolor), Truecolor).Strength(Radar, 1); !blended {
		t.Fatal("the library's own severe tint does not blend, so this proves nothing")
	}
	for tint := range 5 {
		if _, blended := SearchBlends(host, Dark, groundOf(Dark, Truecolor), Truecolor).Strength(Radar, tint); blended {
			t.Errorf("tint %d still blends over a radar ramp whose third and fourth classes are one colour", tint)
		}
	}
}

// TestTheSearchAgreesWithTheChecker: the search's fast path, which stops at
// the first pair too close, passes exactly the strengths CheckBlend passes,
// for every preset, ground, depth, tint and strength tried.
func TestTheSearchAgreesWithTheChecker(t *testing.T) {
	var none Palette
	for _, ground := range []GroundKind{Dark, Light} {
		for _, depth := range []Depth{Truecolor, Colours256} {
			gc := groundOf(ground, depth)
			under := seenAs(shownAt(gc, depth))
			for _, p := range []Preset{Radar, Temperature} {
				ramp := rampOf(none, p, ground, depth)
				plain := make([]seen, len(ramp))
				for i, c := range ramp {
					plain[i] = seenAs(shownAt(c, depth))
				}
				for tint := range 5 {
					tc := tintOf(none, tint, ground, depth)
					for _, s := range blendStrengths {
						slow := true
						for _, f := range CheckBlend(ramp, tc, gc, s, depth) {
							slow = slow && f.Rule != BlendSeparate
						}
						if fast := separates(ramp, plain, under, tc, s, depth); fast != slow {
							t.Errorf("%v %v %v tint %d at %.2f: the search says %v, the checker %v", p, ground, depth, tint, s, fast, slow)
						}
					}
				}
			}
		}
	}
	// A ramp where the ground alone decides: its dark class blended toward a
	// tint the ground's own colour lands on the ground, and nothing else is near.
	ground := groundOf(Dark, Truecolor)
	ramp := []RGB{{R: 40, G: 50, B: 60}, {R: 250, G: 250, B: 250}}
	plain := []seen{seenAs(ramp[0]), seenAs(ramp[1])}
	slow := true
	for _, f := range CheckBlend(ramp, ground, ground, 0.50, Truecolor) {
		slow = slow && f.Rule != BlendSeparate
	}
	if slow {
		t.Fatal("the ground case does not fail the checker, so it proves nothing")
	}
	if separates(ramp, plain, seenAs(ground), ground, 0.50, Truecolor) {
		t.Error("the search passes a blend that lands a class on the ground")
	}
	cs := []RGB{{}, {R: 255, G: 255, B: 255}, {R: 10, G: 120, B: 240}, {R: 200, G: 40, B: 40}, {R: 51, G: 51, B: 120}}
	for _, a := range cs {
		for _, b := range cs {
			if d, e := Difference(a, b), apart(seenAs(a), seenAs(b)); d != e {
				t.Errorf("%v %v: Difference %v, apart %v", a, b, d, e)
			}
		}
	}
}

// BenchmarkSearchBlends is what a change of look costs a map: the search
// runs on a new map's first frame and whenever the palette, ground or depth
// changes. It was 17 ms before the fast path, and a fuzz leg that makes a
// map an input timed out on it.
func BenchmarkSearchBlends(b *testing.B) {
	var none Palette
	g := groundOf(Dark, Truecolor)
	for b.Loop() {
		SearchBlends(none, Dark, g, Truecolor)
	}
}

package colour

import "testing"

// TestFireHasItsColoursOnBothGrounds is L-18.1's defaults (watchpost D-121):
// both fire tokens have a colour of their own on either ground, apart from
// each other, from the track's (the earthquakes') and from the alerts', and
// readable as line work on the ground.
func TestFireHasItsColoursOnBothGrounds(t *testing.T) {
	for _, ground := range []GroundKind{Dark, Light} {
		behind, _ := Palette{}.ResolveAt(Ground, ground, Truecolor)
		fire, ok1 := Palette{}.ResolveAt(Fire, ground, Truecolor)
		faint, ok2 := Palette{}.ResolveAt(FireFaint, ground, Truecolor)
		if !ok1 || !ok2 || fire == faint {
			t.Fatalf("ground %v: fire %v (%v), faint %v (%v); want two colours of their own", ground, fire, ok1, faint, ok2)
		}
		for _, other := range []Token{Track, AlertExtremeOutline, AlertSevereOutline, AlertModerateOutline} {
			if c, _ := (Palette{}).ResolveAt(other, ground, Truecolor); c == fire || c == faint {
				t.Errorf("ground %v: fire shares %v's colour %v", ground, other.Name(), c)
			}
		}
		for name, c := range map[string]RGB{"fire": fire, "fire.faint": faint} {
			if r := Contrast(c, behind); r < 3 {
				t.Errorf("ground %v: %s %v is %.1f:1 on the ground %v; want 3:1 for line work", ground, name, c, r, behind)
			}
		}
	}
}

// TestTheQuakesHaveTheirColoursOnBothGrounds is L-19.3's defaults: the three
// ages apart from each other, from fire's and from the track's, readable as
// line work on either ground.
func TestTheQuakesHaveTheirColoursOnBothGrounds(t *testing.T) {
	for _, ground := range []GroundKind{Dark, Light} {
		behind, _ := Palette{}.ResolveAt(Ground, ground, Truecolor)
		seen := map[RGB]string{}
		for _, tok := range []Token{QuakeHour, QuakeDay, QuakeOlder, Fire, FireFaint, Track} {
			c, ok := Palette{}.ResolveAt(tok, ground, Truecolor)
			if !ok {
				t.Fatalf("ground %v: %s has no colour", ground, tok.Name())
			}
			if other, dup := seen[c]; dup {
				t.Errorf("ground %v: %s shares %s's colour %v", ground, tok.Name(), other, c)
			}
			seen[c] = tok.Name()
			if tok >= QuakeHour && Contrast(c, behind) < 3 {
				t.Errorf("ground %v: %s %v is %.1f:1 on the ground; want 3:1", ground, tok.Name(), c, Contrast(c, behind))
			}
		}
	}
}

// TestTheWaveScalePassesOnTheSea is L-20.2's scale: six classes that pass
// the checker as areas on each ground's water - the waves are the sea's
// alone - at truecolor and at 256 colours.
func TestTheWaveScalePassesOnTheSea(t *testing.T) {
	for _, ground := range []GroundKind{Dark, Light} {
		water, _ := Palette{}.ResolveAt(WaterFill, ground, Truecolor)
		for _, depth := range []Depth{Truecolor, Colours256} {
			ramp, ok := Ramp(Waves, ground, depth)
			if !ok || len(ramp) != 6 {
				t.Fatalf("%v %v: the wave scale is %d classes", ground, depth, len(ramp))
			}
			if depth == Colours256 {
				for i, c := range ramp {
					_, ramp[i] = To256(c)
				}
			}
			if f := Check(ramp, RampCheck{Ground: water, Depth: depth, Midpoint: -1}); len(f) != 0 {
				t.Errorf("%v %v: %+v", ground, depth, f)
			}
		}
	}
}

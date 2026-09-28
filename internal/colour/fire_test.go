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

package colour

import "testing"

// TestWarningsReportEveryScaleToken is L-6.1: a host's colour for any
// preset's class (ScaleClass) that breaks its ramp's rules is reported by
// Warnings - here the ground's own colour, which no ramp may use.
func TestWarningsReportEveryScaleToken(t *testing.T) {
	under, _ := Palette{}.ResolveAt(Ground, Dark, Truecolor)
	for _, tok := range Tokens() {
		if !ScaleClass(tok) {
			continue
		}
		pal, bad := NewPalette(map[string]RGB{tok.Name(): under})
		if len(bad) != 0 {
			t.Fatalf("%s: %v", tok.Name(), bad)
		}
		if w := pal.Warnings(Dark, Truecolor); len(w) == 0 {
			t.Errorf("%s set to the ground's colour: no warning", tok.Name())
		}
	}
}

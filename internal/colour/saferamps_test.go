package colour

import "testing"

// TestSafeRampsKeepEveryScaleTheLibrarysOwn is D-5 and D-29: with safe ramps
// on, a host's colour for any preset's class, or an alert's, is passed over
// for the library's own; the host's other colours stand.
func TestSafeRampsKeepEveryScaleTheLibrarysOwn(t *testing.T) {
	host := RGB{1, 2, 3}
	for _, tok := range Tokens() {
		alert := tok >= AlertExtremeOutline && tok <= AlertUnknownTint
		pal, bad := NewPalette(map[string]RGB{tok.Name(): host})
		if len(bad) != 0 || !pal.Own() {
			t.Fatalf("%s: the host's colour was not taken: %v", tok.Name(), bad)
		}
		got, _ := pal.KeepRamps(true).ResolveAt(tok, Dark, Truecolor)
		kept := got != host
		if want := ScaleClass(tok) || alert; kept != want {
			t.Errorf("%s with safe ramps: host colour passed over %v; want %v", tok.Name(), kept, want)
		}
	}
}

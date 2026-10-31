package tuimaps_test

import (
	"reflect"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestTheNightIsOneOverlayInOneCall is watchpost D-216: NightOverlay is the
// night side in the night's role and the terminator in its own, for the
// moment, as a host composed them by hand; a map takes it as any overlay.
func TestTheNightIsOneOverlayInOneCall(t *testing.T) {
	o := tuimaps.NightOverlay("night", noon)
	want := []tuimaps.Feature{
		{Kind: tuimaps.Polygon, Rings: tuimaps.NightSide(noon), Role: tuimaps.NightTint},
		{Kind: tuimaps.Line, Rings: [][]tuimaps.LonLat{tuimaps.Terminator(noon)}, Role: tuimaps.TerminatorLine},
	}
	if o.ID != "night" || !o.Valid.Equal(noon) || o.Keeps != time.Hour || !reflect.DeepEqual(o.Features, want) {
		t.Fatalf("the night overlay is %q at %v keeping %v, features %d; want the night side and the terminator in their roles", o.ID, o.Valid, o.Keeps, len(o.Features))
	}
	m := world(t, 80, 24)
	must(t, m.FitWorld())
	mustSet(t, m, o)
}

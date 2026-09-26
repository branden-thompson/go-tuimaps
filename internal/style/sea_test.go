package style

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestTheSeaRulesAreNeverSwitched is watchpost's UAT-1 U1-39: switching the
// water off takes the lakes and inland water, never the sea or its coast -
// without them the land has no edge and the country's outline is lost. So
// the sea and the coast are drawn with every layer switched off, and the
// ocean is filled by the sea's rule while a lake is filled by the water's.
func TestTheSeaRulesAreNeverSwitched(t *testing.T) {
	all := Off(RoadLayer, RailLayer, ParkLayer, BorderLayer, RiverLayer, WaterLayer, LabelLayer, MinorRoadLayer)
	p := NewProfile(Bare, 149, 38, all)
	for _, id := range []string{"coast", "sea"} {
		if !p.Draws(ruleOf(t, id)) {
			t.Errorf("%s is not drawn with every layer switched off", id)
		}
	}
	s := BuiltIn()
	for class, want := range map[string]string{"ocean": "sea", "lake": "water", "": "water"} {
		r, ok := s.Fill("water", Attrs{Kind: scene.GeomPolygon, Class: class}, 5)
		if !ok || r.ID != want {
			t.Errorf("a %q water area is filled by %v, want %s", class, r, want)
		}
	}
}

package tuimaps_test

import (
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestTheSeaOutlastsTheWaterSwitch is L-14.2 (D-85; watchpost UAT-1 U1-39):
// WaterLayer is the lakes and inland water. With every layer a host can
// switch turned off, the world still has its coast - the land keeps its edge.
// Without the fix this frame is blank.
func TestTheSeaOutlastsTheWaterSwitch(t *testing.T) {
	const cols, rows = 80, 24
	m := world(t, cols, rows)
	m.ColourDepth(tuimaps.NoColour)
	for _, l := range []tuimaps.Layer{tuimaps.RoadLayer, tuimaps.MinorRoadLayer, tuimaps.RailLayer, tuimaps.ParkLayer,
		tuimaps.BorderLayer, tuimaps.RiverLayer, tuimaps.WaterLayer, tuimaps.LabelLayer} {
		m.Layers(l, false)
	}
	_, text := drawn(t, m, cols, rows)
	coast := 0
	for _, r := range text {
		if r > 0x2800 && r <= 0x28ff { // a braille cell with a dot in it
			coast++
		}
	}
	if coast < 20 {
		t.Errorf("with every layer off the world draws %d cells of coast; the sea and its coast are never switched:\n%s", coast, text)
	}
}

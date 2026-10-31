package tuimaps_test

import (
	"context"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestNoRoleRestsOnColourAlone is L-7.3 (OW-17): a quake's age, a fire's
// strength and a buoy against a tide station each draw a mark of their own,
// so with no colour at all the frames still differ - a ring, as a point, for
// the quakes; a point for the rest.
func TestNoRoleRestsOnColourAlone(t *testing.T) {
	const cols, rows = 149, 38
	at := tuimaps.LonLat{Lon: -90, Lat: 35.1}
	point := func(r tuimaps.Token) tuimaps.Feature {
		return tuimaps.Feature{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{at}}, Role: r}
	}
	ring := func(r tuimaps.Token) tuimaps.Feature {
		return tuimaps.Feature{Kind: tuimaps.Circle, Centre: at, RadiusDots: 8, Role: r}
	}
	drawn := func(f tuimaps.Feature) string {
		t.Helper()
		m := gulfMap(t, cols, rows)
		m.ColourDepth(tuimaps.NoColour)
		if _, err := m.Set(tuimaps.Overlay{ID: "x", Valid: noon, Keeps: 24 * time.Hour, Features: []tuimaps.Feature{f}}); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err)
		}
		return frameAtTime(t, m, cols, rows, noon)
	}
	for _, g := range []struct {
		name  string
		roles []tuimaps.Token
		as    func(tuimaps.Token) tuimaps.Feature
	}{
		{"a fire's strength", []tuimaps.Token{tuimaps.Fire, tuimaps.FireFaint}, point},
		{"a quake's age, as a ring", []tuimaps.Token{tuimaps.QuakeHour, tuimaps.QuakeDay, tuimaps.QuakeOlder}, ring},
		{"a quake's age, as a point", []tuimaps.Token{tuimaps.QuakeHour, tuimaps.QuakeDay, tuimaps.QuakeOlder}, point},
		{"a buoy against a tide station", []tuimaps.Token{tuimaps.Buoy, tuimaps.Tide}, point},
	} {
		seen := map[string]tuimaps.Token{}
		for _, r := range g.roles {
			f := drawn(g.as(r))
			if other, ok := seen[f]; ok {
				t.Errorf("%s: %s and %s draw the same frame with no colour", g.name, other.Name(), r.Name())
			}
			seen[f] = r
		}
	}
}

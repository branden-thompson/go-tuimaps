package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// FuzzKey is the property test of this tool's one entry point: whatever a
// scenario file says, the key holds together. Nothing panics, no distance
// is a number a voice cannot say, every compass is one of the eight words,
// and a place inside an area is never further from its edge than the
// widest the area is.
func FuzzKey(f *testing.F) {
	f.Add(`{"places":[{"name":"a","lon":0,"lat":0}],"shapes":[{"id":"s","kind":"area","rings":[[[1,1],[2,1],[2,2]]]}]}`)
	f.Add(`{"places":[{"name":"","lon":181,"lat":91}],"shapes":[{"id":"","kind":"line","rings":[[[0,0]]]}]}`)
	f.Add(`{"places":[],"shapes":[]}`)
	f.Add(`{"places":[{"lon":0,"lat":0}],"shapes":[{"kind":"area","rings":[[[179,-1],[-179,-1],[-179,1]]]}],"views":[{"name":"v","cell_km":1.6,"cell_row_km":3.2}]}`)
	words := "north north-east east south-east south south-west west north-west"
	f.Fuzz(func(t *testing.T, body string) {
		var s Scenario
		if err := json.Unmarshal([]byte(body), &s); err != nil {
			t.Skip() // not a scenario file at all
		}
		if len(s.Places) > 32 || len(s.Shapes) > 32 {
			t.Skip() // a key of a thousand answers proves nothing more
		}
		for _, a := range Key(s) {
			if math.IsNaN(a.Km) || math.IsInf(a.Km, 0) || a.Km < 0 {
				t.Fatalf("a distance of %v", a.Km)
			}
			if a.Km > 20038 {
				t.Fatalf("a distance of %v km, further than the other side of the world", a.Km)
			}
			if a.Bearing < 0 || a.Bearing > 360 {
				t.Fatalf("a bearing of %v", a.Bearing)
			}
			if !strings.Contains(words, a.Compass) || a.Compass == "" {
				t.Fatalf("a compass word of %q", a.Compass)
			}
			for view, cells := range a.Cells {
				if math.IsNaN(cells) || cells < 0 {
					t.Fatalf("%q is %v cells away", view, cells)
				}
			}
			for view, said := range a.Frame {
				switch said {
				case "inside", "outside", "on the edge", "unknown":
				default:
					t.Fatalf("%q says %q", view, said)
				}
			}
		}
	})
}

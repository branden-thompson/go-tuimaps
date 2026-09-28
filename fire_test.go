package tuimaps_test

// fire_test.go — v0.2.0 L-18 (watchpost D-121): fire is a role of its own,
// drawn in its own colours - a fire's perimeter, its incident and a strong
// hotspot in one, a weaker hotspot fainter - and never reported as an alert.

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// fireOverlay is a perimeter east of Memphis, its incident at its centre,
// and a weak hotspot beside it.
func fireOverlay() tuimaps.Overlay {
	ring := []tuimaps.LonLat{{Lon: -90.5, Lat: 34.8}, {Lon: -89.5, Lat: 34.8}, {Lon: -89.5, Lat: 35.4}, {Lon: -90.5, Lat: 35.4}, {Lon: -90.5, Lat: 34.8}}
	return tuimaps.Overlay{ID: "fire", Valid: noon, Keeps: 24 * time.Hour, Features: []tuimaps.Feature{
		{Kind: tuimaps.Polygon, Rings: [][]tuimaps.LonLat{ring}, Role: tuimaps.Fire, Label: "Test Fire"},
		{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -90, Lat: 35.1}}}, Role: tuimaps.Fire},
		{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -90.9, Lat: 35.1}}}, Role: tuimaps.FireFaint},
	}}
}

// TestFireIsARoleOfItsOwn is L-18.1 and L-18.2: the frame draws a fire in its
// own colour and a weak hotspot in its fainter one, both named tokens; a
// place inside a perimeter has no alert over it.
func TestFireIsARoleOfItsOwn(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	names := " " + strings.Join(tuimaps.TokenNames(), " ") + " "
	if !strings.Contains(names, " fire ") || !strings.Contains(names, " fire.faint ") {
		t.Fatalf("the tokens are%s; want fire and fire.faint", names)
	}
	colours := map[string]tuimaps.RGB{"fire": {R: 250, G: 1, B: 2}, "fire.faint": {R: 250, G: 231, B: 101}}
	if unknown, err := m.SetPalette(colours); err != nil || len(unknown) != 0 {
		t.Fatalf("the fire tokens are not a palette's: %v %v", unknown, err)
	}
	mustSet(t, m, fireOverlay())
	settle(t, m)
	frame := momentFrame(t, m, noon)
	for name, c := range colours {
		fg := "38;2;" + strconv.Itoa(int(c.R)) + ";" + strconv.Itoa(int(c.G)) + ";" + strconv.Itoa(int(c.B))
		if !strings.Contains(frame, fg) {
			t.Errorf("the frame draws nothing in %s's colour", name)
		}
	}
	r, err := m.Report([]tuimaps.Place{{ID: "p", Name: "Inside", At: tuimaps.LonLat{Lon: -90, Lat: 35.1}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Places[0].Alerts) != 0 {
		t.Errorf("a place inside a fire's perimeter has alerts %+v; a fire is no alert", r.Places[0].Alerts)
	}
}

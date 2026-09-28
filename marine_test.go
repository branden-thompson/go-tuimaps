package tuimaps_test

// marine_test.go — v0.2.0 L-21 (watchpost D-127, D-128): buoys and tide
// stations have roles of their own, drawn in their own colours and never
// reported as alerts.

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestBuoysAndTidesAreRolesOfTheirOwn is L-21.1: each a named token,
// settable, drawn in its colour; a place at a station has no alert.
func TestBuoysAndTidesAreRolesOfTheirOwn(t *testing.T) {
	m := reportMap(t)
	m.ColourDepth(tuimaps.Truecolor)
	names := " " + strings.Join(tuimaps.TokenNames(), " ") + " "
	if !strings.Contains(names, " buoy ") || !strings.Contains(names, " tide ") {
		t.Fatalf("the tokens are%s; want buoy and tide", names)
	}
	colours := map[string]tuimaps.RGB{"buoy": {R: 250, G: 1, B: 250}, "tide": {R: 1, G: 250, B: 250}}
	if unknown, err := m.SetPalette(colours); err != nil || len(unknown) != 0 {
		t.Fatalf("the marine tokens are not a palette's: %v %v", unknown, err)
	}
	mustSet(t, m, tuimaps.Overlay{ID: "marine", Valid: noon, Keeps: 24 * time.Hour, Features: []tuimaps.Feature{
		{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -90, Lat: 35.1}}}, Role: tuimaps.Buoy, Label: "5ft 64"},
		{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -90.9, Lat: 35.1}}}, Role: tuimaps.Tide, Label: "H 5.2ft"},
	}})
	settle(t, m)
	frame := momentFrame(t, m, noon)
	for name, c := range colours {
		if !strings.Contains(frame, "38;2;"+strconv.Itoa(int(c.R))+";"+strconv.Itoa(int(c.G))+";"+strconv.Itoa(int(c.B))) {
			t.Errorf("the frame draws nothing in %s's colour", name)
		}
	}
	r, err := m.Report([]tuimaps.Place{{ID: "p", Name: "At the buoy", At: tuimaps.LonLat{Lon: -90, Lat: 35.1}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Places[0].Alerts) != 0 {
		t.Errorf("a place at a buoy has alerts %+v; a station is no alert", r.Places[0].Alerts)
	}
}

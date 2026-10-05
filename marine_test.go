package tuimaps_test

// marine_test.go — v0.2.0 L-21 (watchpost D-127, D-128): buoys and tide
// stations have roles of their own, drawn in their own colours and never
// reported as alerts.

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
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

// TestTheNoticeWhileTheMapLoadsSpeaksToAPerson is L-22.2 (watchpost D-124,
// U2-34): the frame drawn before any tile has landed said "no map tiles yet:
// call Settle, or run Work until it has none left" - words for a programmer,
// shown to a listener. It says the map is loading, and names no call.
func TestTheNoticeWhileTheMapLoadsSpeaksToAPerson(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close() }()
	f, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	if err != nil {
		t.Fatal(err)
	}
	body := colours.ReplaceAllString(strings.Join(f.Lines, "\n"), "")
	if !strings.Contains(body, "Loading the map") || strings.Contains(body, "Settle") || strings.Contains(body, "Work") {
		t.Errorf("the frame before any tile landed says:\n%s", body)
	}
}

// TestTheBasemapDrawsUnderAFloodOfOverlays is L-22.1 (watchpost U2-34): a
// host handed the map one overlay a station - hundreds, past the queue's cap
// - after the view had asked for its tiles. The queue dropped the tiles, the
// oldest jobs, and ran the overlays first: markers over no basemap at all.
// The tiles are kept and run first; as many Work calls as there are tiles
// draw the whole basemap.
func TestTheBasemapDrawsUnderAFloodOfOverlays(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close() }()
	must(t, m.Recentre(tuimaps.LonLat{Lon: -90.5, Lat: 35.5}))
	must(t, m.Zoom(assets.MaxZoom))
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil {
		t.Fatal(err)
	}
	tiles := m.Pending()
	if tiles == 0 {
		t.Fatal("the view asked for no tiles; the test needs it to")
	}
	for i := range 400 {
		mustSet(t, m, tuimaps.Overlay{ID: "station/" + strconv.Itoa(i), Valid: noon, Keeps: 24 * time.Hour, Features: []tuimaps.Feature{
			{Kind: tuimaps.Point, Rings: [][]tuimaps.LonLat{{{Lon: -95 + float64(i%100)/10, Lat: 33 + float64(i/100)}}}, Role: tuimaps.Buoy},
		}})
	}
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon); err != nil { // asks for the overlays' work
		t.Fatal(err)
	}
	for range tiles {
		if _, err := m.Work(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	f, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if body := colours.ReplaceAllString(strings.Join(f.Lines, "\n"), ""); strings.Contains(body, "Loading the map") {
		t.Errorf("after %d Work calls - one a tile - under 400 overlays the map is still loading:\n%s", tiles, body)
	}
}

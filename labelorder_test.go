package tuimaps

// labelorder_test.go — v0.2.0 L-23 (watchpost UAT-2 U2-38, D-135): with every
// layer on, the place names were lost. Every overlay's words and every
// contour's value were placed before the basemap's names, into cells the
// names then could not have. The names come before the data's words.

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
)

// nameWords are a frame's words of four letters or more above its furniture:
// the place names, since a reading, a value or a mark is no such word.
var nameWords = regexp.MustCompile(`[A-Za-z]{4,}`)

// namesDrawn is the place names on a frame of the lower 48 with a field -
// banded, or lined with its values - and, when stations is set, three
// hundred labelled buoys.
func namesDrawn(t *testing.T, lined bool, stations int) (map[string]bool, string) {
	t.Helper()
	m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.ColourDepth(Truecolor)
	if err := m.Recentre(LonLat{Lon: -96, Lat: 37}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	g := warmEast()
	g.Lines = lined
	if _, err := m.Set(TemperatureGrid("temp", g, Celsius, drawnAt)); err != nil {
		t.Fatal(err)
	}
	if stations > 0 {
		var feats []Feature
		for i := range stations {
			feats = append(feats, Feature{Kind: Point, Rings: [][]LonLat{{{Lon: -124 + float64(i%60), Lat: 25 + float64(i/60)*5}}}, Role: Buoy, Label: "4ft " + strconv.Itoa(60+i%30)})
		}
		if _, err := m.Set(Overlay{ID: "buoys", Valid: drawnAt, Keeps: 24 * time.Hour, Features: feats}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	f, err := m.Render(Size{Cols: 120, Rows: 40}, drawnAt)
	if err != nil {
		t.Fatal(err)
	}
	body := escapes.ReplaceAllString(strings.Join(f.Lines[1:len(f.Lines)-1], "\n"), "")
	names := map[string]bool{}
	for _, w := range nameWords.FindAllString(body, -1) {
		names[w] = true
	}
	return names, body
}

// TestThePlaceNamesComeBeforeTheData is L-23: the place names a banded field
// leaves are all still drawn with the field's contour values and three
// hundred labelled stations on it - and the stations' words still take the
// room the names left.
func TestThePlaceNamesComeBeforeTheData(t *testing.T) {
	bare, _ := namesDrawn(t, false, 0)
	if len(bare) == 0 {
		t.Fatal("the frame names no place; the test needs names to lose")
	}
	full, body := namesDrawn(t, true, 300)
	for n := range bare {
		if !full[n] {
			t.Errorf("%s is drawn on the field alone and lost under its values and the stations", n)
		}
	}
	if !strings.Contains(body, "4ft") {
		t.Errorf("no station's reading is drawn in the room the names left:\n%s", body)
	}
}

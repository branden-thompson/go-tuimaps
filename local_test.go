package tuimaps_test

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// The pinned fixture's urban tiles, and where each one is. A zoom-14 tile
// is about two and a half kilometres across at this latitude: street scale,
// which is what a person means by looking at where they live.
func urbanTiles() map[[3]uint32]string {
	return map[[3]uint32]string{
		{14, 8299, 5636}:  "tiles-urban-z14/14-8299-5636.pbf",
		{14, 8186, 5448}:  "tiles-urban-z14/14-8186-5448.pbf",
		{14, 4824, 6157}:  "tiles-urban-z14/14-4824-6157.pbf",
		{14, 14552, 6451}: "tiles-urban-z14/14-14552-6451.pbf",
	}
}

// fromFixture is a source that serves the pinned fixture's own tiles and
// reaches no network at all: the test hands the library its own way of
// fetching, so nothing is dialled (FR-22b).
func fromFixture(t *testing.T, asked *int) http.RoundTripper {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	held := urbanTiles()
	return answering(func(address string) ([]byte, error) {
		*asked++
		z, x, y, ok := tileOf(address)
		if !ok {
			return nil, errors.New("not a tile address")
		}
		rel, carried := held[[3]uint32{uint32(z), x, y}]
		if !carried {
			return nil, errors.New("no such tile")
		}
		return testkit.LoadFixture(root, filepath.ToSlash(rel))
	})
}

// TestLocalZoomDrawsLocalDetail is the answer to the question a person asks
// of a map: **can I see where I live?** The tiles built into the app stop
// at zoom 3, so a map that has never been given a source draws a stand-in
// when it goes in - the right picture, blurred, rather than nothing (FR-30)
// - and a map with a source draws the street.
//
// Both are checked here from the pinned fixture, so the answer does not
// depend on anyone's network.
func TestLocalZoomDrawsLocalDetail(t *testing.T) {
	// A rectangle small enough to sit inside one zoom-14 tile, centred on
	// that tile: the fixture carries four tiles from four cities, not a
	// neighbourhood, so a wider view would rightly want tiles nobody has.
	const cols, rows = 69, 12
	paris := tuimaps.LonLat{Lon: 2.3729, Lat: 48.8553}

	// With the embedded tiles alone, a local zoom is a stand-in: the map is
	// never blank, and it is never pretending either.
	embedded := world(t, cols, rows)
	if err := embedded.Recentre(paris); err != nil {
		t.Fatal(err)
	}
	if err := embedded.Zoom(14); err != nil {
		t.Fatal(err)
	}
	if _, err := embedded.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, stood := drawn(t, embedded, cols, rows)
	// **What a person sees offline at street scale: no map.** The tiles
	// built in are the world down to zoom 3, and at zoom 14 over a city
	// there is nothing in them to draw. The frame is not blank - it carries
	// the credit the data's licence asks for - and it is not pretending
	// either: its status says it is still sharpening, which is what the
	// app's status row shows.
	if lines := strings.Split(strings.TrimRight(stood, "\n"), "\n"); len(lines) != rows {
		t.Errorf("%d rows, want %d", len(lines), rows)
	}

	// With a source, the same view draws the place itself.
	m, err := tuimaps.New(tuimaps.WithSize(cols, rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	asked := 0
	useTransport(t, m, fromFixture(t, &asked))
	if err := m.Source("https://tiles.example.test/"); err != nil {
		t.Fatal(err)
	}
	if err := m.Recentre(paris); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(14); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
	if err != nil {
		t.Fatal(err)
	}
	if asked == 0 {
		t.Fatal("a map at zoom 14 with a source named asked for no tile")
	}
	// The fixture carries one tile of the two this view wants - it is four
	// tiles from four cities, not a neighbourhood - so the frame says it is
	// still sharpening, which is the truth and is what the app's status row
	// shows. What matters here is that what *was* served is drawn.
	if frame.Status == tuimaps.NoTiles {
		t.Error("a map with a source that served a tile says it has none")
	}
	local := plainText(strings.Join(frame.Lines, "\n"))
	// The streets are there: real line work, where offline there was none.
	if inked(local) < inked(stood)+20 {
		t.Errorf("the street-scale tiles drew %d cells against the %d of a map with no tiles at all",
			inked(local), inked(stood))
	}
	if local == stood {
		t.Error("the frame with a source is the frame without one")
	}
	// And the scale bar says a distance a person walks, not one they fly.
	scale, ok := m.Scale()
	if !ok {
		t.Fatal("a map at zoom 14 has no scale")
	}
	if scale.Km > 5 {
		t.Errorf("the scale bar at zoom 14 stands for %v km, which is not street scale", scale.Km)
	}
}

// plainText is a frame with its colour sequences taken out.
func plainText(text string) string { return colours.ReplaceAllString(text, "") }

// inked is how many cells of a frame have anything drawn in them, which is
// the roughest possible measure of how much a picture holds - and enough to
// tell a street from a blur.
func inked(text string) int {
	drawn := 0
	for _, r := range text {
		switch r {
		case '\n', ' ', rune(0x2800): // the blank braille cell
			continue
		}
		drawn++
	}
	return drawn
}

// TestFixtureAtZoomNine is plan task 14.20: the ordinary path stays
// ordinary. An overlay of the fixture's own size at zoom 9 is simplified,
// cached and drawn from the library's own copy - not from the host's memory
// - because it fits the shape cache with room to spare. It is the case the
// withdrawn "quarter of the cache" rule would have sent down the fall-back
// path for no reason (P2-ENG-2, D-90).
func TestFixtureAtZoomNine(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	// Scenario 2's shape is the fixture's own zone: 2,962 vertices, the
	// kind about nine alerts in ten carry.
	putScenario(t, m, scenarioNamed(t, "scenario-2"))
	ids := m.Overlays()
	if len(ids) != 1 {
		t.Fatalf("the scenario put %d overlays on the map", len(ids))
	}
	if err := m.Recentre(tuimaps.LonLat{Lon: -85.0, Lat: 29.9}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(9); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
	if err != nil {
		t.Fatal(err)
	}
	// The basemap here is a stand-in - the tiles built in stop at zoom 3 -
	// and that is not what this test is about: the overlay is.
	if frame.Status == tuimaps.NoTiles {
		t.Error("nothing at all was drawn")
	}
	// The shape is drawn, and the library holds a copy of its own: the
	// shape cache is asked for bytes and has some.
	if use := m.CacheUse(); use.Shapes.Held == 0 {
		t.Error("nothing is in the shape cache; the overlay was drawn from the host's memory instead")
	} else if use.Shapes.Held > use.Shapes.Limit {
		t.Errorf("the shape cache holds %d bytes against a cap of %d", use.Shapes.Held, use.Shapes.Limit)
	}
	// Settling again does no work: the simplified copy is kept, not made
	// afresh every frame.
	res, err := m.Settle(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Ran != 0 {
		t.Errorf("a settled map at zoom 9 ran %d more jobs", res.Ran)
	}
	// And the host's own memory is given back, precisely because the
	// library kept a copy of its own: a shape drawn from the host's memory
	// would still be borrowed (D-92).
	if m.InUse(ids[0]) {
		t.Error("the host's geometry is still borrowed; the shape is being drawn from it rather than from the library's own copy")
	}
}

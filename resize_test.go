package tuimaps

import (
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
)

// drawnAt is the moment these tests draw at, so nothing moves under them.
var drawnAt = time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

// TestResizeKeepsWhereTheHostIsLooking is the defect the first host found
// (task 14.19). **A terminal host cannot know its size when it makes the map**:
// the size arrives with the first frame and changes again whenever the window
// does. So the size a map is drawn at is almost never the size it was made at,
// and a resize that threw the view away meant the host's chosen place was lost
// on the very first frame - silently, with no error and no status, leaving a
// map of the whole world where the host had asked for a town.
func TestResizeKeepsWhereTheHostIsLooking(t *testing.T) {
	m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	want := LonLat{Lon: -117.22, Lat: 33.29}
	if err := m.Recentre(want); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(Size{Cols: 73, Rows: 26}, drawnAt); err != nil {
		t.Fatal(err)
	}
	at, zoom := m.Centre()
	if at != want || zoom != 3 {
		t.Errorf("drawn at another size the map is at %v zoom %v; the host asked for %v zoom 3", at, zoom, want)
	}
	// And again, because a window is resized more than once.
	if _, err := m.Render(Size{Cols: 200, Rows: 50}, drawnAt); err != nil {
		t.Fatal(err)
	}
	if at, zoom := m.Centre(); at != want || zoom != 3 {
		t.Errorf("after a second resize the map is at %v zoom %v", at, zoom)
	}
}

// TestResizeStillFitsTheWorldWhenNobodyHasLooked holds the other half of the
// rule. A map nobody has pointed anywhere is showing the whole world, and the
// whole world at one size is a different zoom from the whole world at another,
// so that map must be refitted when the size changes - which is what makes the
// three-call world map work whatever size the terminal turns out to be
// (NFR-19).
func TestResizeStillFitsTheWorldWhenNobodyHasLooked(t *testing.T) {
	m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	_, wide := m.Centre()
	if _, err := m.Render(Size{Cols: 40, Rows: 12}, drawnAt); err != nil {
		t.Fatal(err)
	}
	_, narrow := m.Centre()
	if narrow >= wide {
		t.Errorf("a world map drawn into a smaller window is at zoom %v, no further out than the %v it had", narrow, wide)
	}
}

// TestTheHostsViewSurvivesEveryWayOfSettingIt checks the flag is set wherever a
// host says where to look, not only by Recentre: FitTo is how a host frames its
// own places, and it is the call a first host reaches for.
func TestTheHostsViewSurvivesEveryWayOfSettingIt(t *testing.T) {
	m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	places := []LonLat{{Lon: -117.2, Lat: 33.2}, {Lon: -117.0, Lat: 33.4}}
	if err := m.FitTo(places, nil, 2); err != nil {
		t.Fatal(err)
	}
	at, zoom := m.Centre()
	if _, err := m.Render(Size{Cols: 73, Rows: 26}, drawnAt); err != nil {
		t.Fatal(err)
	}
	if got, z := m.Centre(); got != at {
		t.Errorf("after FitTo and a resize the centre moved from %v to %v (zoom %v to %v)", at, got, zoom, z)
	}
}

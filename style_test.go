package tuimaps_test

import (
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// userStyle draws water and nothing else, in a colour of its own. It is the
// same JSON format as upstream's, with the legacy filter form.
const userStyle = `{
 "constants": {"@sea": "#102080"},
 "layers": [
  {"id": "bg", "type": "background", "paint": {"background-color": "#000000"}},
  {"id": "sea", "type": "fill", "source-layer": "water", "paint": {"fill-color": "@sea"}}
 ]}`

// TestStyleOfTheHostsOwn is FR-20's library half, which the app's --style
// flag needs (D-94): the library is handed a style as bytes and never reads
// a file. A style takes effect at the next frame, re-reads no tile, and one
// that cannot be read is refused with the map still drawing as it was.
func TestStyleOfTheHostsOwn(t *testing.T) {
	const cols, rows = 149, 38
	m := world(t, cols, rows)
	own, _ := drawn(t, m, cols, rows)

	// A style of the host's own draws its own map.
	if err := m.SetStyle([]byte(userStyle)); err != nil {
		t.Fatal(err)
	}
	theirs, theirText := drawn(t, m, cols, rows)
	if theirs == own {
		t.Error("the frame is unchanged after a style of the host's own was set")
	}
	if strings.TrimSpace(theirText) == "" {
		t.Error("the host's style drew nothing at all")
	}
	// Nothing was fetched or decoded again for it: a style is not part of a
	// tile's cache key, as a language is (D-82).
	if pending := m.Pending(); pending != 0 {
		t.Errorf("setting a style left %d jobs pending", pending)
	}

	// A style that cannot be read is refused, and the map draws on as it was.
	for _, bad := range []string{"layers:", `{"layers":{}}`,
		`{"layers":[{"id":"a","type":"line","source-layer":"water","filter":["match",["get","class"],"ocean",true,false]}]}`} {
		err := m.SetStyle([]byte(bad))
		if kind, ok := tuimaps.KindOf(err); !ok || kind != tuimaps.MalformedStyle {
			t.Errorf("%q: %v, want the malformed-style kind", bad, err)
		}
		if after, _ := drawn(t, m, cols, rows); after != theirs {
			t.Errorf("%q: a refused style changed the map", bad)
		}
	}

	// No style at all is the library's own again.
	if err := m.SetStyle(nil); err != nil {
		t.Fatal(err)
	}
	if back, _ := drawn(t, m, cols, rows); back != own {
		t.Error("with no style of the host's own the library's own is not back")
	}
}

// TestStyleOnAClosedMap: every call on a closed map says so.
func TestStyleOnAClosedMap(t *testing.T) {
	m := world(t, 80, 24)
	if inside := m.Close(); inside != 0 {
		t.Fatalf("%d calls were inside the map when it closed", inside)
	}
	err := m.SetStyle([]byte(userStyle))
	if kind, ok := tuimaps.KindOf(err); !ok || kind != tuimaps.Closed {
		t.Errorf("a style set on a closed map: %v", err)
	}
	if !isKind(err, fault.Closed) {
		t.Errorf("%v is not one of the library's own errors of the closed kind", err)
	}
}
